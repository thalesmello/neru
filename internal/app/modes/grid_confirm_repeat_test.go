package modes

import (
	"context"
	"image"
	"testing"

	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	"github.com/y3owk1n/neru/internal/app/services"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/action"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports/mocks"
	"go.uber.org/zap"
)

func TestGridMode_ConfirmOnRepeat_ClicksOnlyAnUnchangedFinePoint(t *testing.T) {
	g := domainGrid.NewGridWithLabels("ABCD", "", "", image.Rect(0, 0, 900, 600), zap.NewNop())
	manager := domainGrid.NewManager(g, domain.GridDimensions{Rows: 3, Cols: 3}, "uiojklm,.", nil, nil, zap.NewNop())
	ctx := &gridcomponent.Context{}
	ctx.SetConfirmOnRepeat(true)
	ctx.SetCursorFollowSelection(false)
	ctx.SetGridInstance(&g)
	clicks := []image.Point{}
	mouse := &mocks.MockMousePort{PerformActionAtPointFunc: func(_ context.Context, kind action.Type, point image.Point, _ action.Modifiers) error {
		clicks = append(clicks, point)
		return nil
	}}
	h := newHandlerWithState(handlerState{ctx: context.Background(), logger: zap.NewNop(), config: config.DefaultConfig(), appState: state.NewAppState(), cursorState: state.NewCursorState(),
		grid:         &components.GridComponent{Manager: manager, Router: domainGrid.NewRouter(manager, zap.NewNop()), Context: ctx},
		screenBounds: image.Rect(0, 0, 900, 600), actionService: services.NewActionService(&mocks.MockAccessibilityPort{}, mouse, &mocks.MockOverlayPort{}, &mocks.MockSystemPort{}, zap.NewNop())})
	for _, key := range g.Cells()[0].Coordinate() {
		h.handleGridModeKey(string(key))
	}
	if !manager.InSubgrid() {
		t.Fatal("fine grid not opened")
	}
	h.handleGridModeKey("i")
	if len(clicks) != 0 {
		t.Fatal("first aim clicked")
	}
	point, _ := ctx.SelectionPoint()
	// There is no time condition: an arbitrarily later unchanged aim clicks.
	h.handleGridModeKey("i")
	h.handleGridModeKey("i")
	h.handleGridModeKey("i")
	if len(clicks) != 3 {
		t.Fatalf("got %d clicks, want 3", len(clicks))
	}
	for _, p := range clicks {
		if p != point {
			t.Fatal("repeat moved the point")
		}
	}
	if !manager.InSubgrid() {
		t.Fatal("confirmation exited fine grid")
	}
	h.handleGridModeKey("u")
	if len(clicks) != 3 {
		t.Fatal("new aim clicked")
	}
	h.handleGridModeKey("u")
	if len(clicks) != 4 {
		t.Fatal("unchanged new aim did not click")
	}
}
