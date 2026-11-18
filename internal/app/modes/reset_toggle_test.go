package modes

import (
	"context"
	"image"
	"sync"
	"testing"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	rgcomponent "github.com/y3owk1n/neru/internal/app/components/recursivegrid"
	scrollcomponent "github.com/y3owk1n/neru/internal/app/components/scroll"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/domain/recursivegrid"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports/mocks"
)

func newResetToggleFixture(mode domain.Mode) (*Handler, *mocks.MockOverlayPort) {
	bounds := image.Rect(0, 0, 900, 600)
	g := grid.NewGridWithLabels("ABCD", "", "", bounds, zap.NewNop())
	gm := grid.NewManager(g, domain.GridDimensions{Rows: 3, Cols: 3}, "uiojklm,.", nil, nil, zap.NewNop())
	gc := &gridcomponent.Context{}
	gc.SetGridInstance(&g)
	gc.SetCursorFollowSelection(false)
	rgc := &rgcomponent.Context{}
	rgc.SetCursorFollowSelection(false)
	rgc.SetSelectionPoint(image.Pt(450, 300))
	app := state.NewAppState()
	app.SetMode(mode)
	overlay := &mocks.MockOverlayPort{}
	h := NewHandler(HandlerDeps{
		Ctx: context.Background(), Config: config.DefaultConfig(), AppState: app,
		CursorState: state.NewCursorState(), OverlayPort: overlay,
		ScrollComponent: &components.ScrollComponent{Context: &scrollcomponent.Context{}},
		GridComponent:   &components.GridComponent{Context: gc, Manager: gm, Router: grid.NewRouter(gm, nil)},
		RecursiveGridComponent: &components.RecursiveGridComponent{
			Context: rgc, Manager: recursivegrid.NewManager(bounds, "uiojklm,.", recursivegrid.SelectionCallbacks{}, zap.NewNop()),
		},
	})
	h.screenBounds = bounds

	return h, overlay
}

func TestResetCurrentMode_ToggleResetsDirtyGridAndExitsOnlyAtBase(t *testing.T) {
	tests := []struct {
		name     string
		mode     domain.Mode
		prepare  func(*testing.T, *Handler)
		toggle   bool
		wantMode domain.Mode
	}{
		{name: "grid base exits", mode: domain.ModeGrid, toggle: true, wantMode: domain.ModeIdle},
		{name: "partial label resets", mode: domain.ModeGrid, toggle: true, wantMode: domain.ModeGrid,
			prepare: func(t *testing.T, h *Handler) {
				h.grid.Manager.HandleInput(h.grid.Manager.CurrentGrid().AllCells()[0].Coordinate()[:1])
			}},
		{name: "selected fine cell resets", mode: domain.ModeGrid, toggle: true, wantMode: domain.ModeGrid,
			prepare: func(t *testing.T, h *Handler) {
				for _, key := range h.grid.Manager.CurrentGrid().AllCells()[0].Coordinate() {
					h.grid.Manager.HandleInput(string(key))
				}
				h.handleGridModeKey("j")
				if _, selected := h.grid.Context.SelectionPoint(); !selected {
					t.Fatal("fine selection missing")
				}
			}},
		{name: "root nudge resets", mode: domain.ModeGrid, toggle: true, wantMode: domain.ModeGrid,
			prepare: func(t *testing.T, h *Handler) {
				if err := h.NudgeGrid(domain.DirectionRight, 4, 1); err != nil {
					t.Fatal(err)
				}
			}},
		{name: "recursive depth one resets", mode: domain.ModeRecursiveGrid, toggle: true, wantMode: domain.ModeRecursiveGrid,
			prepare: func(t *testing.T, h *Handler) { h.recursiveGrid.Manager.HandleInput("j") }},
		{name: "recursive root center exits", mode: domain.ModeRecursiveGrid, toggle: true, wantMode: domain.ModeIdle},
		{name: "plain grid root stays", mode: domain.ModeGrid, wantMode: domain.ModeGrid},
		{name: "plain recursive root stays", mode: domain.ModeRecursiveGrid, wantMode: domain.ModeRecursiveGrid},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := newResetToggleFixture(tc.mode)
			if tc.prepare != nil {
				tc.prepare(t, h)
			}
			var err error
			if tc.toggle {
				err = h.ResetCurrentMode(true)
			} else {
				err = h.ResetCurrentMode()
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := h.appState.CurrentMode(); got != tc.wantMode {
				t.Fatalf("mode = %v, want %v", got, tc.wantMode)
			}
			if tc.wantMode != domain.ModeIdle {
				reporter, _ := activeModeExtension[baseStateReporter](&h.handlerState)
				if !reporter.AtBase() {
					t.Fatal("reset did not restore base state")
				}
			}
		})
	}
}

func TestResetCurrentMode_ToggleRefusesOtherModes(t *testing.T) {
	for _, mode := range []domain.Mode{domain.ModeIdle, domain.ModeHints, domain.ModeScroll, domain.ModeBisect, domain.ModeMonitorSelect, domain.ModeCustom, domain.ModeVim} {
		h, _ := newResetToggleFixture(mode)
		if err := h.ResetCurrentMode(true); err == nil {
			t.Fatalf("mode %v accepted toggle", mode)
		}
		if h.appState.CurrentMode() != mode {
			t.Fatalf("mode %v exited", mode)
		}
	}
}

func TestResetCurrentMode_ToggleConcurrentSelection(t *testing.T) {
	h, _ := newResetToggleFixture(domain.ModeGrid)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for range 100 {
				h.CurrentSelectionPoint()
				_ = h.ResetCurrentMode(true)
			}
		})
	}
	workers.Wait()
	if h.appState.CurrentMode() != domain.ModeIdle {
		t.Fatal("base did not exit")
	}
}
