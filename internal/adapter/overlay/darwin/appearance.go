//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#include "../../platform/darwin/overlay.h"
*/
import "C"

import (
	"math"

	"github.com/y3owk1n/neru/internal/adapter/overlay/manager"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
)

// Ensure Manager declares GridAppearancePort.
var _ ports.GridAppearancePort = (*Manager)(nil)

// SwitchTo scopes the live fill override to the two grid surfaces. This is
// a transition-only main-queue update, never extra work on the typing path.
func (m *Manager) SwitchTo(next manager.Mode) {
	if m == nil {
		return
	}
	m.Base.SwitchTo(next)
	m.shareMu.RLock()
	defer m.shareMu.RUnlock()
	if m.window != nil {
		enabled := next == manager.ModeGrid || next == manager.ModeRecursiveGrid
		C.NeruSetGridCellOpacityEnabled(m.window, C.int(boolToInt(enabled)))
	}
}

// SetGridOpacity overrides the cell background alpha in Grid and RecursiveGrid,
// including the fine-adjustment subgrid. Borders, text, badges and other modes
// are unaffected. opacity must be in [0, 1].
//
// Range is validated before headless is checked so callers receive
// CodeInvalidInput regardless of whether a window exists.
//
// shareMu is held read-only to prevent a concurrent Destroy from freeing the
// window handle while the C side copies it. The C function dispatches
// asynchronously to the Cocoa main queue and never calls back into Go.
func (m *Manager) SetGridOpacity(opacity float64) error {
	if math.IsNaN(opacity) || math.IsInf(opacity, 0) || opacity < 0 || opacity > 1 {
		return derrors.New(derrors.CodeInvalidInput, "grid opacity must be in [0, 1]")
	}
	if m == nil {
		return derrors.New(derrors.CodeNotSupported, "grid opacity: headless overlay has no native window")
	}

	m.shareMu.RLock()
	defer m.shareMu.RUnlock()
	window := m.window
	if window == nil {
		return derrors.New(derrors.CodeNotSupported, "grid opacity: overlay window was destroyed")
	}

	C.NeruSetGridCellOpacity(window, C.double(opacity))

	return nil
}
