package mocks

import (
	"github.com/y3owk1n/neru/internal/ports"
	"time"
)

var _ ports.GridPeekPort = (*MockGridPeekPort)(nil)

// MockGridPeekPort never touches desktop pixels or timers.
type MockGridPeekPort struct{ PeekGridFunc func(time.Duration) error }

func (m *MockGridPeekPort) PeekGrid(duration time.Duration) error {
	if m.PeekGridFunc != nil {
		return m.PeekGridFunc(duration)
	}
	return nil
}
