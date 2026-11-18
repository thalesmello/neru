package overlay

import (
	"math"

	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
)

var _ ports.GridAppearancePort = (*Adapter)(nil)

// SetGridOpacity changes only presentation; the backend updates its existing
// surface and retains the mode's selection, capture and configured borders.
func (a *Adapter) SetGridOpacity(opacity float64) error {
	if math.IsNaN(opacity) || math.IsInf(opacity, 0) || opacity < 0 || opacity > 1 {
		return derrors.New(derrors.CodeInvalidInput, "grid opacity must be a finite number between 0 and 1")
	}
	if a.released() {
		return derrors.New(derrors.CodeNotSupported, "the overlay has been destroyed")
	}
	appearance, ok := a.manager.(ports.GridAppearancePort)
	if !ok {
		return derrors.New(derrors.CodeNotSupported, "live grid opacity is not supported by this overlay backend")
	}
	return appearance.SetGridOpacity(opacity)
}
