package modes

import (
	"github.com/y3owk1n/neru/internal/app/components"
	gridcomponent "github.com/y3owk1n/neru/internal/app/components/grid"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
	"go.uber.org/zap"
	"image"
	"testing"
)

func TestSetGridRepeat_PreservesSelection(t *testing.T) {
	app := state.NewAppState()
	app.SetMode(domain.ModeGrid)
	ctx := &gridcomponent.Context{}
	point := image.Pt(200, 300)
	ctx.SetSelectionPoint(point)
	h := newHandlerWithState(handlerState{appState: app, logger: zap.NewNop(), grid: &components.GridComponent{Context: ctx}})
	for _, enabled := range []bool{true, false} {
		if !h.SetGridRepeat(enabled) {
			t.Fatal("active grid refused repeat update")
		}
		if ctx.Repeat() != enabled {
			t.Fatal("flag not updated")
		}
		got, ok := ctx.SelectionPoint()
		if !ok || got != point {
			t.Fatal("repeat update reset selection")
		}
		if app.CurrentMode() != domain.ModeGrid {
			t.Fatal("repeat update exited grid")
		}
	}
	app.SetMode(domain.ModeIdle)
	if h.SetGridRepeat(true) {
		t.Fatal("idle accepted grid repeat")
	}
}
