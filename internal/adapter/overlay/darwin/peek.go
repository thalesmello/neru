//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#include "../../platform/darwin/overlay.h"
*/
import "C"

import (
	"github.com/y3owk1n/neru/internal/adapter/overlay/manager"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
	"math"
	"time"
)

var _ ports.GridPeekPort = (*Manager)(nil)

// PeekGrid schedules the mask and restoration on Cocoa's main queue. The
// native generation guard cancels it on surface teardown or replacement.
func (m *Manager) PeekGrid(duration time.Duration, opacity float64) error {
	if math.IsNaN(opacity) || math.IsInf(opacity, 0) || opacity < 0 || opacity > 1 {
		return derrors.New(derrors.CodeInvalidInput, "peek opacity must be between 0 and 1")
	}
	if duration < time.Millisecond || duration > 10*time.Second {
		return derrors.New(derrors.CodeInvalidInput, "grid peek duration must be between 1ms and 10s")
	}
	if m == nil {
		return derrors.New(derrors.CodeNotSupported, "grid peek requires a native overlay")
	}
	mode := m.Mode()
	if mode != manager.ModeGrid && mode != manager.ModeRecursiveGrid {
		return derrors.New(derrors.CodeInvalidInput, "grid peek requires an active grid")
	}
	m.shareMu.RLock()
	defer m.shareMu.RUnlock()
	if m.window == nil {
		return derrors.New(derrors.CodeNotSupported, "grid peek requires a native overlay")
	}
	C.NeruPeekGrid(m.window, C.double(duration.Seconds()), C.double(opacity))
	return nil
}
