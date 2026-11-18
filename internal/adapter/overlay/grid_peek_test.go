package overlay_test

import (
	"github.com/y3owk1n/neru/internal/adapter/overlay"
	"github.com/y3owk1n/neru/internal/derrors"
	"testing"
	"time"
)

type peekManager struct {
	appearanceManager
	peeks int
}

func (m *peekManager) PeekGrid(time.Duration, float64) error { m.peeks++; return nil }
func TestAdapter_PeekGrid_DoesNotChangeSurfaceLifecycle(t *testing.T) {
	m := &peekManager{}
	a := overlay.NewAdapter(m, nil, nil)
	if err := a.PeekGrid(time.Second, 0.25); err != nil {
		t.Fatal(err)
	}
	if m.peeks != 1 || m.shows != 0 || m.hides != 0 || m.clears != 0 || m.resizes != 0 {
		t.Fatal("peek changed lifecycle")
	}
	for _, d := range []time.Duration{0, -time.Second, 11 * time.Second} {
		if err := a.PeekGrid(d, 0.25); err == nil {
			t.Fatal("invalid duration accepted")
		}
	}
	a.Destroy()
	if err := a.PeekGrid(time.Second, 0.25); !derrors.IsNotSupported(err) {
		t.Fatal("destroyed surface accepted peek")
	}
	if m.peeks != 1 {
		t.Fatal("peek reached destroyed surface")
	}
}
