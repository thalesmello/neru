package ipcctrl

import (
	"context"
	"strings"
	"time"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/ports"
)

func (h *OverlayHandler) handlePeekGrid(_ context.Context, cmd ipc.Command) ipc.Response {
	duration, err := parsePeekDuration(cmd.Args)
	if err != nil {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: err.Error()}
	}
	if h.appState == nil || (h.appState.CurrentMode() != domain.ModeGrid && h.appState.CurrentMode() != domain.ModeRecursiveGrid) {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: "peek-grid requires an active grid or recursive_grid"}
	}
	peek, ok := h.overlay.(ports.GridPeekPort)
	if !ok {
		return ipc.Response{Code: ipc.CodeNotSupported, Message: "grid peek is not supported by this overlay backend"}
	}
	if err := peek.PeekGrid(duration); err != nil {
		code := ipc.CodeActionFailed
		if derrors.IsNotSupported(err) {
			code = ipc.CodeNotSupported
		}
		if derrors.IsCode(err, derrors.CodeInvalidInput) {
			code = ipc.CodeInvalidInput
		}
		return ipc.Response{Code: code, Message: err.Error()}
	}
	return ipc.Response{Success: true, Code: ipc.CodeOK, Message: "grid peek started"}
}

func parsePeekDuration(args []string) (time.Duration, error) {
	duration := time.Second
	for i := 0; i < len(args); i++ {
		name, value, inline := strings.Cut(args[i], "=")
		if name != "--duration" {
			return 0, derrors.New(derrors.CodeInvalidInput, "peek-grid accepts only --duration")
		}
		if !inline {
			i++
			if i >= len(args) {
				return 0, derrors.New(derrors.CodeInvalidInput, "--duration requires a value")
			}
			value = args[i]
		}
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed < time.Millisecond || parsed > 10*time.Second {
			return 0, derrors.New(derrors.CodeInvalidInput, "peek duration must be between 1ms and 10s")
		}
		duration = parsed
	}
	return duration, nil
}
