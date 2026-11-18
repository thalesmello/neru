package ipcctrl

import (
	"context"
	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/domain/state"
	"testing"
)

func TestSetGridRepeat_DelegatesWithoutReactivation(t *testing.T) {
	h := NewOverlayHandler(state.NewAppState(), nil)
	called := false
	h.setGridRepeat = func(enabled bool) bool { called = enabled; return true }
	r := h.handleSetGridRepeat(context.Background(), ipc.Command{Args: []string{"--state", "on"}})
	if !r.Success || !called {
		t.Fatalf("got %+v", r)
	}
	for _, args := range [][]string{nil, {"--state", "bad"}} {
		if h.handleSetGridRepeat(context.Background(), ipc.Command{Args: args}).Success {
			t.Fatalf("accepted %v", args)
		}
	}
}
