package ipcctrl

import (
	"context"
	"testing"
	"time"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports/mocks"
)

type peekOverlayMock struct {
	mocks.MockOverlayPort
	calls    int
	duration time.Duration
}

func (m *peekOverlayMock) PeekGrid(d time.Duration, opacity float64) error {
	m.calls++
	m.duration = d
	return nil
}

func TestPeekGrid_ValidatesDuration(t *testing.T) {
	for _, args := range [][]string{{"--duration=0"}, {"--duration=11s"}, {"--duration=NaN"}, {"--other"}, {"--duration"}} {
		if _, err := parsePeekDuration(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	for _, args := range [][]string{nil, {"--duration=1s"}, {"--duration", "500ms"}} {
		if _, err := parsePeekDuration(args); err != nil {
			t.Fatal(err)
		}
	}
}
func TestPeekGrid_ValidatesOpacity(t *testing.T) {
	for _, value := range []string{"NaN", "Inf", "-0.1", "1.1"} {
		if _, _, err := parsePeekOptions([]string{"--opacity", value}); err == nil {
			t.Fatalf("accepted %s", value)
		}
	}
	_, opacity, err := parsePeekOptions([]string{"--duration", "1s", "--opacity", "0.25"})
	if err != nil || opacity != 0.25 {
		t.Fatalf("opacity=%v err=%v", opacity, err)
	}
}

func TestOverlayHandler_PeekGrid_RequiresActiveGrid(t *testing.T) {
	app := state.NewAppState()
	mock := &peekOverlayMock{}
	h := NewOverlayHandler(app, nil, mock)
	resp := h.handlePeekGrid(context.Background(), ipc.Command{})
	if resp.Success || mock.calls != 0 {
		t.Fatal("idle peek reached overlay")
	}
	for _, mode := range []domain.Mode{domain.ModeGrid, domain.ModeRecursiveGrid} {
		app.SetMode(mode)
		resp = h.handlePeekGrid(context.Background(), ipc.Command{})
		if !resp.Success || mock.duration != time.Second {
			t.Fatalf("got %+v", resp)
		}
	}
}
func TestOverlayHandler_PeekGrid_Unsupported(t *testing.T) {
	app := state.NewAppState()
	app.SetMode(domain.ModeGrid)
	h := NewOverlayHandler(app, nil, &mocks.MockOverlayPort{})
	resp := h.handlePeekGrid(context.Background(), ipc.Command{})
	if resp.Code != ipc.CodeNotSupported {
		t.Fatalf("got %+v", resp)
	}
}
