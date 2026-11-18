package modes

import (
	"image"
	"math"
	"strings"

	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
)

var _ nativeNudger = (*GridMode)(nil)

// gridNudgeAccumulator is owned by handlerState and accessed only under h.mu.
// Integer step counts retain the fractional remainder rather than rounding
// each repeat separately. Selection writes by anyone else invalidate the anchor.
type gridNudgeAccumulator struct {
	manager   *domainGrid.Manager
	grid      *domainGrid.Grid
	prefix    string
	fineIndex int
	bounds    image.Rectangle
	writes    uint64
	divisor   int
	anchor    image.Point
	steps     image.Point
}

// NudgeGrid moves only an active native Grid selection. Count zero initializes
// an absent selection at the active rectangle's center without moving it.
func (h *Handler) NudgeGrid(direction domain.Direction, nudgesPerCell, count int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if direction.String() == domain.UnknownDirection || nudgesPerCell <= 0 || count < 0 {
		return derrors.New(derrors.CodeInvalidInput, "nudge-grid requires a cardinal direction, positive nudges-per-cell, and nonnegative count")
	}
	nudger, ok := activeModeEffect[nativeNudger](&h.handlerState, extensionNativeNudge)
	if !ok {
		return derrors.New(derrors.CodeInvalidInput, "nudge-grid requires an active native Grid")
	}
	return nudger.NativeNudge(direction, nudgesPerCell, count)
}

// gridNudgeBounds chooses fine-cell, coarse-cell, or the union of prefix matches.
func gridNudgeBounds(manager *domainGrid.Manager) (image.Rectangle, int) {
	if fine, index, ok := manager.LastSubgridCell(); ok {
		return fine, index
	}
	if cell, ok := manager.SelectedCell(); ok {
		return cell.Bounds(), -1
	}
	var bounds image.Rectangle
	if grid := manager.CurrentGrid(); grid != nil {
		for _, cell := range grid.AllCells() {
			if strings.HasPrefix(cell.Coordinate(), manager.CurrentInput()) {
				bounds = bounds.Union(cell.Bounds())
			}
		}
	}
	return bounds, -1
}

// NativeNudge leaves input, click disposition, keyboard capture and frame intact.
func (m *GridMode) NativeNudge(direction domain.Direction, nudgesPerCell, count int) error {
	h := m.handler
	if h.grid == nil || h.grid.Context == nil || h.grid.Manager == nil || h.grid.Manager.CurrentGrid() == nil {
		return derrors.New(derrors.CodeInvalidInput, "nudge-grid requires an active native Grid selection context")
	}
	manager := h.grid.Manager
	bounds, fineIndex := gridNudgeBounds(manager)
	bounds = bounds.Add(h.screenBounds.Min)
	whole := manager.CurrentGrid().Bounds().Add(h.screenBounds.Min)
	if bounds.Empty() || whole.Empty() {
		return derrors.New(derrors.CodeInvalidInput, "nudge-grid has no matching grid cells")
	}
	point, selected := m.SelectionPoint()
	if !selected {
		point = image.Pt(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2)
	}
	acc := &h.gridNudge
	if !selected || acc.manager != manager || acc.grid != manager.CurrentGrid() ||
		acc.prefix != manager.CurrentInput() || acc.fineIndex != fineIndex ||
		acc.bounds != bounds || acc.writes != m.SelectionWrites() || acc.divisor != nudgesPerCell {
		*acc = gridNudgeAccumulator{manager: manager, grid: manager.CurrentGrid(),
			prefix: manager.CurrentInput(), fineIndex: fineIndex, bounds: bounds,
			divisor: nudgesPerCell, anchor: point}
	}
	dx, dy := direction.Delta()
	acc.steps.X += dx * count
	acc.steps.Y += dy * count
	target := image.Pt(
		acc.anchor.X+int(math.Round(float64(acc.steps.X)*float64(bounds.Dx())/float64(nudgesPerCell))),
		acc.anchor.Y+int(math.Round(float64(acc.steps.Y)*float64(bounds.Dy())/float64(nudgesPerCell))),
	)
	clamped := image.Pt(max(whole.Min.X, min(whole.Max.X-1, target.X)), max(whole.Min.Y, min(whole.Max.Y-1, target.Y)))
	if clamped != target {
		acc.anchor = clamped
		acc.steps = image.Point{}
	}
	m.PlaceSelectionPoint(clamped)
	acc.writes = m.SelectionWrites()
	if h.grid.Context.CursorFollowSelection() && h.actionService != nil {
		return h.actionService.MoveCursorToPoint(h.ctx, clamped)
	}
	return nil
}
