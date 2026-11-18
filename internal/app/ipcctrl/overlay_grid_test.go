package ipcctrl

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
	portmocks "github.com/y3owk1n/neru/internal/ports/mocks"
)

// gridAppearanceMock wraps MockOverlayPort and adds SetGridOpacity / PeekGrid
// so that the type assertion to ports.GridAppearancePort succeeds.
type gridAppearanceMock struct {
	*portmocks.MockOverlayPort
	setOpacityErr error
	lastOpacity   float64
	opacityCalls  int
}

func (m *gridAppearanceMock) SetGridOpacity(opacity float64) error {
	m.opacityCalls++
	m.lastOpacity = opacity

	return m.setOpacityErr
}

func (m *gridAppearanceMock) PeekGrid(_ time.Duration) error { return nil }

func newGridAppearanceMock() *gridAppearanceMock {
	return &gridAppearanceMock{MockOverlayPort: &portmocks.MockOverlayPort{}}
}

// TestParseSetGridOpacityArgs_AcceptsValidValues pins valid values and both
// flag spellings (--value=X and --value X).
func TestParseSetGridOpacityArgs_AcceptsValidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want float64
	}{
		{"zero inline", []string{"--value=0"}, 0},
		{"one inline", []string{"--value=1"}, 1},
		{"mid inline", []string{"--value=0.45"}, 0.45},
		{"mid separate", []string{"--value", "0.45"}, 0.45},
		{"last wins", []string{"--value=0.2", "--value=0.8"}, 0.8},
		{"blank skipped", []string{"", "  ", "--value=0.5"}, 0.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, errResp := parseSetGridOpacityArgs(tc.args)
			if errResp != nil {
				t.Fatalf("parseSetGridOpacityArgs(%v) = error %+v, want %.4g",
					tc.args, errResp, tc.want)
			}

			if got != tc.want {
				t.Fatalf("parseSetGridOpacityArgs(%v) = %g, want %g", tc.args, got, tc.want)
			}
		})
	}
}

// TestParseSetGridOpacityArgs_RejectsInvalidValues pins NaN, Inf, out-of-range,
// unknown flags, and a missing value.
func TestParseSetGridOpacityArgs_RejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"missing flag", nil},
		{"NaN", []string{"--value=NaN"}},
		{"pos Inf", []string{"--value=+Inf"}},
		{"neg Inf", []string{"--value=-Inf"}},
		{"below zero", []string{"--value=-0.1"}},
		{"above one", []string{"--value=1.001"}},
		{"unknown flag", []string{"--opacity=0.5"}},
		{"missing value at end", []string{"--value"}},
		{"flag as value", []string{"--value", "--value=0.5"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, errResp := parseSetGridOpacityArgs(tc.args)
			if errResp == nil {
				t.Fatalf("parseSetGridOpacityArgs(%v) = %g, want an error", tc.args, got)
			}

			if errResp.Code != ipc.CodeInvalidInput {
				t.Fatalf("parseSetGridOpacityArgs(%v) code = %s, want %s",
					tc.args, errResp.Code, ipc.CodeInvalidInput)
			}
		})
	}
}

// TestHandleSetGridOpacity_DelegatesOpacityToPort pins that a valid value
// reaches the GridAppearancePort and that the response is successful.
func TestHandleSetGridOpacity_DelegatesOpacityToPort(t *testing.T) {
	t.Parallel()

	mock := newGridAppearanceMock()
	h := NewOverlayHandler(state.NewAppState(), zap.NewNop(), mock)

	resp := h.handleSetGridOpacity(context.Background(),
		ipc.Command{Args: []string{"--value=0.45"}})

	if !resp.Success {
		t.Fatalf("handleSetGridOpacity() = %+v, want success", resp)
	}

	if resp.Code != ipc.CodeOK {
		t.Fatalf("code = %s, want %s", resp.Code, ipc.CodeOK)
	}

	if mock.opacityCalls != 1 {
		t.Fatalf("SetGridOpacity called %d times, want 1", mock.opacityCalls)
	}

	if mock.lastOpacity != 0.45 {
		t.Fatalf("opacity = %g, want 0.45", mock.lastOpacity)
	}
}

// TestHandleSetGridOpacity_ReturnsNotSupportedWhenNoPort pins that a nil or
// absent overlay yields CodeNotSupported.
func TestHandleSetGridOpacity_ReturnsNotSupportedWhenNoPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler *OverlayHandler
	}{
		{"no overlay arg", NewOverlayHandler(state.NewAppState(), zap.NewNop())},
		{"nil variadic", NewOverlayHandler(state.NewAppState(), zap.NewNop(), nil)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := tc.handler.handleSetGridOpacity(context.Background(),
				ipc.Command{Args: []string{"--value=0.5"}})

			if resp.Success || resp.Code != ipc.CodeNotSupported {
				t.Fatalf("handleSetGridOpacity() = %+v, want %s", resp, ipc.CodeNotSupported)
			}
		})
	}
}

// TestHandleSetGridOpacity_ReturnsNotSupportedForPlainOverlay pins that an
// overlay that does NOT implement GridAppearancePort yields CodeNotSupported.
func TestHandleSetGridOpacity_ReturnsNotSupportedForPlainOverlay(t *testing.T) {
	t.Parallel()

	// MockOverlayPort does not embed GridAppearancePort, so the type assertion
	// in gridAppearanceCapability will return nil.
	plain := &portmocks.MockOverlayPort{}
	h := NewOverlayHandler(state.NewAppState(), zap.NewNop(), plain)

	resp := h.handleSetGridOpacity(context.Background(),
		ipc.Command{Args: []string{"--value=0.3"}})

	if resp.Success || resp.Code != ipc.CodeNotSupported {
		t.Fatalf("handleSetGridOpacity() = %+v, want %s", resp, ipc.CodeNotSupported)
	}
}

// TestHandleSetGridOpacity_MapsPortErrorCodes pins that a port returning a
// derrors.CodeNotSupported error maps to ipc.CodeNotSupported and any other
// error maps to ipc.CodeActionFailed.
func TestHandleSetGridOpacity_MapsPortErrorCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		portErr  error
		wantCode string
	}{
		{
			"derrors not supported",
			derrors.New(derrors.CodeNotSupported, "not implemented on this backend"),
			ipc.CodeNotSupported,
		},
		{
			"generic error",
			errors.New("native call failed"),
			ipc.CodeActionFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := newGridAppearanceMock()
			mock.setOpacityErr = tc.portErr

			h := NewOverlayHandler(state.NewAppState(), zap.NewNop(), mock)

			resp := h.handleSetGridOpacity(context.Background(),
				ipc.Command{Args: []string{"--value=0.5"}})

			if resp.Success {
				t.Fatalf("handleSetGridOpacity() succeeded unexpectedly")
			}

			if resp.Code != tc.wantCode {
				t.Fatalf("code = %s, want %s", resp.Code, tc.wantCode)
			}
		})
	}
}

// TestHandleSetGridOpacity_RejectsInvalidArgs pins that bad args return
// CodeInvalidInput without touching the port.
func TestHandleSetGridOpacity_RejectsInvalidArgs(t *testing.T) {
	t.Parallel()

	mock := newGridAppearanceMock()
	h := NewOverlayHandler(state.NewAppState(), zap.NewNop(), mock)

	badCases := [][]string{
		nil,
		{"--value=NaN"},
		{"--value=1.5"},
		{"--opacity=0.5"},
	}

	for _, args := range badCases {
		resp := h.handleSetGridOpacity(context.Background(), ipc.Command{Args: args})
		if resp.Success || resp.Code != ipc.CodeInvalidInput {
			t.Errorf("args %v: got %+v, want %s", args, resp, ipc.CodeInvalidInput)
		}

		if mock.opacityCalls != 0 {
			t.Errorf("args %v: port called on invalid args", args)
		}
	}
}

// TestCommandSetGridOpacity_IsRegistered pins that the constant matches the
// handler's registration key.
func TestCommandSetGridOpacity_IsRegistered(t *testing.T) {
	t.Parallel()

	h := NewOverlayHandler(state.NewAppState(), zap.NewNop())
	handlers := make(map[string]func(context.Context, ipc.Command) ipc.Response)
	h.RegisterHandlers(handlers)

	if _, ok := handlers[domain.CommandSetGridOpacity]; !ok {
		t.Fatalf("RegisterHandlers: %q not registered", domain.CommandSetGridOpacity)
	}
}
