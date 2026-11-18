package modes

import (
	"context"
	"image"
	"strings"
	"sync"
	"testing"

	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	"github.com/y3owk1n/neru/internal/app/services"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/action"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports"
	"github.com/y3owk1n/neru/internal/ports/mocks"
	"go.uber.org/zap"
)

func newGridNudgeFixture() (*Handler, *mocks.MockMousePort, *mocks.MockOverlayPort, *mocks.MockEventTapPort, *mocks.MockSystemPort) {
	g := domainGrid.NewGridWithLabels("ABCD", "", "", image.Rect(0, 0, 907, 607), zap.NewNop())
	manager := domainGrid.NewManager(g, domain.GridDimensions{Rows: 3, Cols: 3}, "uiojklm,.", nil, nil, zap.NewNop())
	ctx := &gridcomponent.Context{}
	ctx.SetGridInstance(&g)
	ctx.SetCursorFollowSelection(false)
	app := state.NewAppState()
	app.SetMode(domain.ModeGrid)
	mouse := &mocks.MockMousePort{}
	overlay := &mocks.MockOverlayPort{}
	tap := &mocks.MockEventTapPort{}
	system := &mocks.MockSystemPort{}
	h := newHandlerWithState(handlerState{ctx: context.Background(), appState: app, logger: zap.NewNop(),
		grid:         &components.GridComponent{Context: ctx, Manager: manager, Router: domainGrid.NewRouter(manager, zap.NewNop())},
		screenBounds: image.Rect(-500, 100, 407, 707), overlayPort: overlay, eventTap: tap,
		actionService: services.NewActionService(&mocks.MockAccessibilityPort{}, mouse, overlay, system, zap.NewNop()),
	})
	return h, mouse, overlay, tap, system
}

func TestNudgeGrid_ConcurrentSelectionAccess(t *testing.T) {
	h, _, _, _, _ := newGridNudgeFixture()
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for range 100 {
				if err := h.NudgeGrid(domain.DirectionRight, 4, 1); err != nil {
					t.Error(err)
				}
				h.CurrentSelectionPoint()
				h.ClearCurrentSelectionPoint()
			}
		})
	}
	workers.Wait()
}

func TestNudgeGrid_CoarsePrefixInitialCenter(t *testing.T) {
	h, _, _, _, _ := newGridNudgeFixture()
	manager := h.grid.Manager
	prefix := manager.CurrentGrid().AllCells()[0].Coordinate()[:1]
	manager.HandleInput(prefix)
	var expected image.Rectangle
	for _, cell := range manager.CurrentGrid().AllCells() {
		if strings.HasPrefix(cell.Coordinate(), prefix) {
			expected = expected.Union(cell.Bounds())
		}
	}
	if err := h.NudgeGrid(domain.DirectionRight, 4, 0); err != nil {
		t.Fatal(err)
	}
	want := image.Pt(expected.Min.X+expected.Dx()/2, expected.Min.Y+expected.Dy()/2).Add(h.screenBounds.Min)
	if got, ok := h.CurrentSelectionPoint(); !ok || got != want {
		t.Fatalf("selection %v, want %v", got, want)
	}
	if manager.CurrentInput() != prefix {
		t.Fatal("nudge changed prefix")
	}
}

func TestNudgeGrid_FineGranularityFourStepsExactAndSelectionResets(t *testing.T) {
	h, _, _, _, _ := newGridNudgeFixture()
	manager := h.grid.Manager
	// Pick an interior coarse cell to leave room for a full fine-cell nudge.
	cells := manager.CurrentGrid().AllCells()
	cell := cells[len(cells)/2]
	for _, key := range cell.Coordinate() {
		manager.HandleInput(string(key))
	}
	if bounds, index := gridNudgeBounds(manager); bounds != cell.Bounds() || index != -1 {
		t.Fatalf("coarse bounds %v, index %d", bounds, index)
	}
	h.handleGridModeKey("j")
	fine, _, ok := manager.LastSubgridCell()
	if !ok {
		t.Fatal("fine key not tracked")
	}
	if fine.Dx()%4 == 0 {
		t.Fatal("fixture must exercise fractional nudge widths")
	}
	start, _ := h.CurrentSelectionPoint()
	for range 4 {
		if err := h.NudgeGrid(domain.DirectionRight, 4, 1); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := h.CurrentSelectionPoint()
	if got != start.Add(image.Pt(fine.Dx(), 0)) {
		t.Fatalf("four steps: %v, start %v, width %d", got, start, fine.Dx())
	}
	h.handleGridModeKey("k")
	fine, _, _ = manager.LastSubgridCell()
	centered, _ := h.CurrentSelectionPoint()
	wantCenter := image.Pt(fine.Min.X+(fine.Dx()+1)/2, fine.Min.Y+(fine.Dy()+1)/2).Add(h.screenBounds.Min)
	if centered != wantCenter {
		t.Fatalf("new fine selection %v, want center %v", centered, wantCenter)
	}
	if err := h.NudgeGrid(domain.DirectionDown, 4, 1); err != nil {
		t.Fatal(err)
	}
	got, _ = h.CurrentSelectionPoint()
	if got.X != centered.X {
		t.Fatal("new selection retained old horizontal accumulator")
	}
	// An external selection write must also discard the fractional anchor.
	h.grid.Context.SetSelectionPoint(centered)
	if err := h.NudgeGrid(domain.DirectionDown, 4, 1); err != nil {
		t.Fatal(err)
	}
	reset, _ := h.CurrentSelectionPoint()
	if reset != got {
		t.Fatalf("selection write did not reset accumulator: %v vs %v", reset, got)
	}
	manager.ResetSilent()
	if _, _, ok := manager.LastSubgridCell(); ok {
		t.Fatal("reset kept fine cell")
	}
}

func TestNudgeGrid_NoClickCaptureExitOrFrameTransition(t *testing.T) {
	h, mouse, overlay, tap, system := newGridNudgeFixture()
	moves, clicks, transitions, releases := 0, 0, 0, 0
	system.MoveCursorToPointFunc = func(context.Context, image.Point, bool) error { moves++; return nil }
	mouse.PerformActionAtPointFunc = func(context.Context, action.Type, image.Point, action.Modifiers) error { clicks++; return nil }
	overlay.ShowFrameFunc = func(context.Context, ports.Frame) error { transitions++; return nil }
	overlay.ClearFrameFunc = func(context.Context) error { transitions++; return nil }
	tap.DisableFunc = func(context.Context) error { releases++; return nil }
	for _, count := range []int{0, 1, 4, 100} {
		if err := h.NudgeGrid(domain.DirectionDown, 4, count); err != nil {
			t.Fatal(err)
		}
	}
	if moves != 0 || clicks != 0 {
		t.Fatal("cursor-hold moved or clicked physical mouse")
	}
	if h.appState.CurrentMode() != domain.ModeGrid {
		t.Fatal("nudge exited grid")
	}
	if transitions != 0 {
		t.Fatal("nudge transitioned frame")
	}
	if releases != 0 {
		t.Fatal("nudge released capture")
	}
	point, _ := h.CurrentSelectionPoint()
	if !point.In(h.grid.Manager.CurrentGrid().Bounds().Add(h.screenBounds.Min)) {
		t.Fatal("nudge escaped grid bounds")
	}
	h.grid.Context.SetCursorFollowSelection(true)
	if err := h.NudgeGrid(domain.DirectionUp, 4, 1); err != nil {
		t.Fatal(err)
	}
	if moves != 1 || clicks != 0 {
		t.Fatal("cursor-follow did not move only")
	}
	h.appState.SetMode(domain.ModeRecursiveGrid)
	if h.NudgeGrid(domain.DirectionUp, 4, 1) == nil {
		t.Fatal("recursive grid accepted native nudge")
	}
	h.appState.SetMode(domain.ModeIdle)
	if h.NudgeGrid(domain.DirectionUp, 4, 0) == nil {
		t.Fatal("idle accepted native nudge")
	}
}
