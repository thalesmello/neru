package mocks

import (
	"github.com/y3owk1n/neru/internal/ports"
	"time"
)

var _ ports.GridPeekPort = (*MockGridPeekPort)(nil)

// MockGridPeekPort never touches desktop pixels or timers.
type MockGridPeekPort struct {
	PeekGridFunc func(time.Duration, float64) error
}

func (m *MockGridPeekPort) PeekGrid(duration time.Duration, opacity float64) error {
	if m.PeekGridFunc != nil {
		return m.PeekGridFunc(duration, opacity)
	}
	return nil
}
