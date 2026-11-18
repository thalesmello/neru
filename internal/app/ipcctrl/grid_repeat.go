package ipcctrl

import (
	"context"
	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/domain"
)

func (h *OverlayHandler) handleSetGridRepeat(_ context.Context, cmd ipc.Command) ipc.Response {
	desired, err := parseToggleState(domain.CommandSetGridRepeat, cmd.Args)
	if err != nil {
		return *err
	}
	if desired == nil {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: "set-grid-repeat requires --state on or off"}
	}
	if h.setGridRepeat == nil || !h.setGridRepeat(*desired) {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: "set-grid-repeat requires an active grid"}
	}
	return ipc.Response{Success: true, Code: ipc.CodeOK, Message: "grid repeat updated"}
}
