package ipcctrl

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/parity"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/ports"
)

// OverlayHandler handles the overlay commands, which hide the overlay from
// screen sharing and adjust grid appearance.
type OverlayHandler struct {
	appState *state.AppState
	logger   *zap.Logger
	// platform names the platform the handler answers for.
	platform func() (parity.Platform, bool)
	// overlay is the optional overlay port used for grid-appearance operations.
	overlay ports.OverlayPort
}

// NewOverlayHandler creates a new overlay command handler.
//
// An optional ports.OverlayPort may be supplied as overlayPort[0]; when
// present it is type-asserted to ports.GridAppearancePort to serve
// set-grid-opacity. Callers that omit it, or pass nil, receive
// CodeNotSupported for that command.
func NewOverlayHandler(appState *state.AppState, logger *zap.Logger, overlayPort ...ports.OverlayPort) *OverlayHandler {
	var overlay ports.OverlayPort
	if len(overlayPort) > 0 {
		overlay = overlayPort[0]
	}

	return &OverlayHandler{
		appState: appState,
		logger:   logger,
		platform: parity.Current,
		overlay:  overlay,
	}
}

// RegisterHandlers registers overlay command handlers.
func (h *OverlayHandler) RegisterHandlers(
	handlers map[string]func(context.Context, ipc.Command) ipc.Response,
) {
	handlers[domain.CommandToggleScreenShare] = h.handleToggleScreenShare
	handlers[domain.CommandSetGridOpacity] = h.handleSetGridOpacity
	handlers[domain.CommandPeekGrid] = h.handlePeekGrid
}

func (h *OverlayHandler) handleToggleScreenShare(
	_ context.Context,
	cmd ipc.Command,
) ipc.Response {
	refusal := h.screenShareRefusal()
	if refusal != nil {
		return *refusal
	}

	desired, errResponse := parseToggleState(domain.CommandToggleScreenShare, cmd.Args)
	if errResponse != nil {
		return *errResponse
	}

	// --state names the reported state, so "on" is hidden. Naming it after the
	// visibility instead would invert between the flag and the status field.
	newState := applyToggleState(
		desired,
		h.appState.ToggleHiddenForScreenShare,
		h.appState.SetHiddenForScreenShare,
	)

	status := "visible"
	if newState {
		status = "hidden"
	}

	return ipc.Response{
		Success: true,
		Message: "screen share visibility: " + status,
		Code:    ipc.CodeOK,
		Data:    ipc.ScreenShareData{Hidden: newState},
	}
}

// screenShareRefusal refuses toggle-screen-share where the option it switches
// does nothing, rather than reporting a hidden overlay that nothing hid.
func (h *OverlayHandler) screenShareRefusal() *ipc.Response {
	platform, known := h.platform()
	if known && config.HidesFromScreenShare(platform) {
		return nil
	}

	where := string(platform)
	if !known {
		where = "this platform"
	}

	return &ipc.Response{
		Success: false,
		Message: domain.CommandToggleScreenShare + " does nothing on " + where + ": " +
			config.ScreenShareHideNote,
		Code: ipc.CodeNotSupported,
	}
}

// gridAppearanceCapability returns the GridAppearancePort from the overlay, or
// nil when the overlay does not implement it.
func (h *OverlayHandler) gridAppearanceCapability() ports.GridAppearancePort {
	if h.overlay == nil {
		return nil
	}

	ga, _ := h.overlay.(ports.GridAppearancePort)

	return ga
}

// handleSetGridOpacity handles the set-grid-opacity IPC command.
//
// The command overrides the cell-fill alpha for the lifetime of the daemon
// session. It is allowed in idle as well as in grid modes so that a binding
// can pre-set the opacity before activating a grid. It never writes config,
// reloads config, or resets mode state.
func (h *OverlayHandler) handleSetGridOpacity(
	_ context.Context,
	cmd ipc.Command,
) ipc.Response {
	opacity, errResp := parseSetGridOpacityArgs(cmd.Args)
	if errResp != nil {
		return *errResp
	}

	ga := h.gridAppearanceCapability()
	if ga == nil {
		return ipc.Response{
			Success: false,
			Message: domain.CommandSetGridOpacity + " is not supported by the active overlay backend",
			Code:    ipc.CodeNotSupported,
		}
	}

	if err := ga.SetGridOpacity(opacity); err != nil {
		code := ipc.CodeNotSupported
		if !derrors.IsNotSupported(err) {
			code = ipc.CodeActionFailed
		}

		return ipc.Response{
			Success: false,
			Message: fmt.Sprintf("%s: %v", domain.CommandSetGridOpacity, err),
			Code:    code,
		}
	}

	return ipc.Response{
		Success: true,
		Message: fmt.Sprintf("grid opacity set to %.4g", opacity),
		Code:    ipc.CodeOK,
	}
}

// --- argument parser ----------------------------------------------------------

const flagOpacityValue = "--value"

// parseSetGridOpacityArgs parses --value <float> from set-grid-opacity args.
// It rejects NaN, Inf, and values outside [0, 1]. The flag is required.
func parseSetGridOpacityArgs(args []string) (float64, *ipc.Response) {
	var opacity float64

	found := false

	for idx := 0; idx < len(args); idx++ {
		arg := strings.TrimSpace(args[idx])
		if arg == "" {
			continue
		}

		name, val, hasInline := strings.Cut(arg, "=")
		if name != flagOpacityValue {
			return 0, &ipc.Response{
				Success: false,
				Message: fmt.Sprintf(
					"%s does not accept %q (only %s <float>)",
					domain.CommandSetGridOpacity, arg, flagOpacityValue,
				),
				Code: ipc.CodeInvalidInput,
			}
		}

		if !hasInline {
			if idx+1 >= len(args) || strings.HasPrefix(args[idx+1], "--") {
				return 0, opacityValueRequired()
			}

			idx++
			val = strings.TrimSpace(args[idx])
		}

		f, parseErr := strconv.ParseFloat(val, 64)
		if parseErr != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, &ipc.Response{
				Success: false,
				Message: fmt.Sprintf(
					"%s --value %q is not a finite number",
					domain.CommandSetGridOpacity, val,
				),
				Code: ipc.CodeInvalidInput,
			}
		}

		if f < 0 || f > 1 {
			return 0, &ipc.Response{
				Success: false,
				Message: fmt.Sprintf(
					"%s --value %s is out of range [0, 1]",
					domain.CommandSetGridOpacity, val,
				),
				Code: ipc.CodeInvalidInput,
			}
		}

		opacity = f
		found = true
	}

	if !found {
		return 0, opacityValueRequired()
	}

	return opacity, nil
}

func opacityValueRequired() *ipc.Response {
	return &ipc.Response{
		Success: false,
		Message: domain.CommandSetGridOpacity + " requires --value <float> in [0, 1]",
		Code:    ipc.CodeInvalidInput,
	}
}
