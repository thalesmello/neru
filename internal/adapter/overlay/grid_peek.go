package overlay

import (
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
	"time"
)

var _ ports.GridPeekPort = (*Adapter)(nil)

// PeekGrid delegates a state-preserving visual mask, not a mode transition.
func (a *Adapter) PeekGrid(duration time.Duration) error {
	if duration < time.Millisecond || duration > 10*time.Second {
		return derrors.New(derrors.CodeInvalidInput, "grid peek duration must be between 1ms and 10s")
	}
	if a.released() {
		return derrors.New(derrors.CodeNotSupported, "the overlay has been destroyed")
	}
	peek, ok := a.manager.(ports.GridPeekPort)
	if !ok {
		return derrors.New(derrors.CodeNotSupported, "grid peek is not supported by this overlay backend")
	}
	return peek.PeekGrid(duration)
}
