package mocks

import "github.com/y3owk1n/neru/internal/ports"

var _ ports.GridAppearancePort = (*MockGridAppearancePort)(nil)

// MockGridAppearancePort is a display-independent runtime-opacity capability.
type MockGridAppearancePort struct {
	SetGridOpacityFunc func(float64) error
}

func (m *MockGridAppearancePort) SetGridOpacity(opacity float64) error {
	if m.SetGridOpacityFunc != nil {
		return m.SetGridOpacityFunc(opacity)
	}
	return nil
}
