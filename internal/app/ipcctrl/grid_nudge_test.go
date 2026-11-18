package ipcctrl

import (
	"context"
	"testing"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/app/modes"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
)

func TestNudgeGrid_ValidatesAndDelegates(t *testing.T) {
	h := NewOverlayHandler(state.NewAppState(), nil)
	calls := 0
	h.nudgeGrid = func(dir domain.Direction, nudges, count int) error {
		calls++
		if dir != domain.DirectionLeft || nudges != 4 || count != 0 {
			t.Fatalf("unexpected nudge %v %d %d", dir, nudges, count)
		}
		return nil
	}
	response := h.handleNudgeGrid(context.Background(), ipc.Command{Args: []string{"--direction", "left", "--nudges-per-cell", "4", "--count", "0"}})
	if !response.Success || calls != 1 {
		t.Fatalf("response %+v, calls %d", response, calls)
	}
	for _, args := range [][]string{
		nil, {"--direction", "diagonal"}, {"--direction", "left", "--count", "-1"},
		{"--direction", "left", "--nudges-per-cell", "0"}, {"--direction", "left", "--count", "bad"},
		{"--direction", "left", "--unknown"}, {"--direction", "left", "extra"},
	} {
		if response := h.handleNudgeGrid(context.Background(), ipc.Command{Args: args}); response.Success || response.Code != ipc.CodeInvalidInput {
			t.Fatalf("accepted %v: %+v", args, response)
		}
	}
	if calls != 1 {
		t.Fatal("invalid input reached nudge callback")
	}
}

func TestNudgeGrid_ControllerRoutesAndExplicitlyRefusesIdle(t *testing.T) {
	app := state.NewAppState()
	c := New(Deps{AppState: app, Modes: modes.NewHandler(modes.HandlerDeps{AppState: app})})
	response := c.HandleCommand(context.Background(), ipc.Command{Action: domain.CommandNudgeGrid, Args: []string{"--direction=up"}})
	if response.Success || response.Code == ipc.CodeUnknownCommand || response.Message == "" {
		t.Fatalf("idle refusal: %+v", response)
	}
}
