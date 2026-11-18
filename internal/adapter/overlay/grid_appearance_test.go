package overlay_test

import (
	"errors"
	"math"
	"testing"

	"github.com/y3owk1n/neru/internal/adapter/overlay"
	"github.com/y3owk1n/neru/internal/derrors"
)

type appearanceManager struct {
	headlessManager
	opacity                              float64
	calls, clears, hides, shows, resizes int
	err                                  error
}

func (m *appearanceManager) SetGridOpacity(value float64) error {
	m.opacity = value
	m.calls++
	return m.err
}
func (m *appearanceManager) Clear()                { m.clears++ }
func (m *appearanceManager) Hide()                 { m.hides++ }
func (m *appearanceManager) Show()                 { m.shows++ }
func (m *appearanceManager) ResizeToActiveScreen() { m.resizes++ }

func TestAdapter_SetGridOpacity_DoesNotRecreateOrClearSurface(t *testing.T) {
	m := &appearanceManager{}
	a := overlay.NewAdapter(m, nil, nil)
	for _, value := range []float64{0, 0.05, 0.5, 1} {
		if err := a.SetGridOpacity(value); err != nil {
			t.Fatal(err)
		}
		if m.opacity != value {
			t.Fatalf("opacity = %v, want %v", m.opacity, value)
		}
	}
	if m.calls != 4 || m.clears != 0 || m.hides != 0 || m.shows != 0 || m.resizes != 0 {
		t.Fatalf("appearance update changed lifecycle: %+v", m)
	}
}

func TestAdapter_SetGridOpacity_RefusesUnsupportedBackend(t *testing.T) {
	a := overlay.NewAdapter(&headlessManager{}, nil, nil)
	if err := a.SetGridOpacity(0.5); !derrors.IsNotSupported(err) {
		t.Fatalf("got %v", err)
	}
}

func TestAdapter_SetGridOpacity_RejectsInvalidValues(t *testing.T) {
	m := &appearanceManager{}
	a := overlay.NewAdapter(m, nil, nil)
	for _, value := range []float64{-0.01, 1.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := a.SetGridOpacity(value); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
	if m.calls != 0 {
		t.Fatal("invalid values reached backend")
	}
}

func TestAdapter_SetGridOpacity_PropagatesBackendFailure(t *testing.T) {
	want := errors.New("backend unavailable")
	a := overlay.NewAdapter(&appearanceManager{err: want}, nil, nil)
	if got := a.SetGridOpacity(0.5); !errors.Is(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestAdapter_SetGridOpacity_DestroyedSurfaceIsNotTouched(t *testing.T) {
	m := &appearanceManager{}
	a := overlay.NewAdapter(m, nil, nil)
	a.Destroy()
	if err := a.SetGridOpacity(0.5); !derrors.IsNotSupported(err) {
		t.Fatalf("got %v", err)
	}
	if m.calls != 0 {
		t.Fatal("reached destroyed backend")
	}
}
