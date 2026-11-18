package modes

import (
	"context"
	"image"
	"testing"

	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports"
	"github.com/y3owk1n/neru/internal/ports/mocks"
	"go.uber.org/zap"
)

func TestGridMode_ResetInput_RevealsBaseCellsAfterSubgrid(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Grid.HideUnmatched = true
	g := domainGrid.NewGridWithLabels("ABCD", "", "", image.Rect(0, 0, 900, 600), zap.NewNop())
	ctx := &gridcomponent.Context{}
	ctx.SetGridInstance(&g)
	ctx.SetCursorFollowSelection(false)
	app := state.NewAppState()
	app.SetMode(domain.ModeGrid)
	overlay := &mocks.MockOverlayPort{}
	h := newHandlerWithState(handlerState{ctx: context.Background(), logger: zap.NewNop(), config: cfg, appState: app,
		overlayPort: overlay, screenBounds: image.Rect(0, 0, 900, 600), grid: &components.GridComponent{Context: ctx}})
	h.initializeGridManager(g)
	h.grid.Router = domainGrid.NewRouter(h.grid.Manager, zap.NewNop())
	for _, key := range g.Cells()[0].Coordinate() {
		h.handleGridModeKey(string(key))
	}
	if !h.grid.Manager.InSubgrid() || !overlay.GridHideUnmatched() {
		t.Fatal("fixture did not reach a filtered subgrid")
	}
	draws, clears := 0, 0
	overlay.RedrawFrameFunc = func(_ context.Context, frame ports.Frame) error {
		draws++
		if overlay.GridHideUnmatched() {
			t.Error("base cells are still filtered when redraw lands")
		}
		if f, ok := frame.(ports.GridFrame); !ok || f.Input != "" {
			t.Error("reset did not redraw the unfiltered base grid")
		}
		return nil
	}
	overlay.ClearFrameFunc = func(context.Context) error { clears++; return nil }
	NewGridMode(&h.handlerState).ResetInput()
	if overlay.GridHideUnmatched() {
		t.Error("reset kept hide_unmatched enabled")
	}
	if draws != 1 || clears != 0 {
		t.Fatalf("draws=%d clears=%d; reset must update existing frame once", draws, clears)
	}
	if h.grid.Manager.InSubgrid() || app.CurrentMode() != domain.ModeGrid {
		t.Error("reset did not keep base Grid active")
	}
}
