package ipcctrl

import (
	"context"
	"flag"
	"io"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
)

func (h *OverlayHandler) handleNudgeGrid(_ context.Context, cmd ipc.Command) ipc.Response {
	flags := flag.NewFlagSet(domain.CommandNudgeGrid, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	direction := flags.String("direction", "", "Cardinal direction")
	nudges := flags.Int("nudges-per-cell", 4, "Steps per active cell")
	count := flags.Int("count", 1, "Number of steps; zero initializes selection")
	if err := flags.Parse(cmd.Args); err != nil {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: err.Error()}
	}
	dir, err := domain.ParseDirection(*direction)
	if err != nil || *nudges <= 0 || *count < 0 || flags.NArg() != 0 {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: "nudge-grid requires --direction up|down|left|right, positive --nudges-per-cell, and nonnegative --count"}
	}
	if h.nudgeGrid == nil {
		return ipc.Response{Code: ipc.CodeInvalidInput, Message: "nudge-grid requires an active native Grid"}
	}
	if err := h.nudgeGrid(dir, *nudges, *count); err != nil {
		code := ipc.CodeActionFailed
		if derrors.IsCode(err, derrors.CodeInvalidInput) {
			code = ipc.CodeInvalidInput
		}
		return ipc.Response{Code: code, Message: err.Error()}
	}
	return ipc.Response{Success: true, Code: ipc.CodeOK, Message: "grid selection nudged"}
}
