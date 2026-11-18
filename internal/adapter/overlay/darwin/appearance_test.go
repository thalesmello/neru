//go:build darwin

package darwin_test

import (
	"testing"

	"github.com/y3owk1n/neru/internal/adapter/overlay/darwin"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/ports"
)

// Compile-time: darwin.Manager satisfies GridAppearancePort.
var _ ports.GridAppearancePort = (*darwin.Manager)(nil)

// TestManager_SetGridOpacity_RangeValidation_RejectedBeforeHeadlessCheck pins
// that an out-of-range opacity returns CodeInvalidInput regardless of whether a
// window exists. Range validation runs first so callers receive a consistent
// code without needing a live backend.
func TestManager_SetGridOpacity_RangeValidation_RejectedBeforeHeadlessCheck(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		opacity float64
		wantErr bool
	}{
		{"zero is valid", 0.0, false},
		{"one is valid", 1.0, false},
		{"midpoint is valid", 0.5, false},
		{"negative rejected", -0.001, true},
		{"above one rejected", 1.001, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Zero-value Manager is headless (no window). The range check runs
			// first, so an invalid value still returns CodeInvalidInput.
			mgr := &darwin.Manager{}
			err := mgr.SetGridOpacity(tc.opacity)

			if !tc.wantErr {
				// Valid range hits the headless check next — expect CodeNotSupported.
				if err == nil {
					t.Fatal("SetGridOpacity() = nil on headless manager, want CodeNotSupported")
				}
				if !derrors.IsNotSupported(err) {
					t.Errorf("SetGridOpacity(%v) on headless = %v (code %q), want CodeNotSupported",
						tc.opacity, err, derrors.GetCode(err))
				}
				return
			}

			// Invalid range must come back as CodeInvalidInput before the
			// headless check would fire CodeNotSupported.
			if err == nil {
				t.Fatalf("SetGridOpacity(%v) = nil, want CodeInvalidInput", tc.opacity)
			}
			if derrors.IsNotSupported(err) {
				t.Errorf("SetGridOpacity(%v) = CodeNotSupported, want CodeInvalidInput (range runs first)",
					tc.opacity)
			}
			if !derrors.IsCode(err, derrors.CodeInvalidInput) {
				t.Errorf("SetGridOpacity(%v) code = %q, want CodeInvalidInput", tc.opacity, derrors.GetCode(err))
			}
		})
	}
}

// TestManager_SetGridOpacity_HeadlessReturnsNotSupported pins that a manager
// with no window (zero value or nil) always returns CodeNotSupported for
// valid opacity values, so callers can distinguish "not implemented" from a
// draw failure.
func TestManager_SetGridOpacity_HeadlessReturnsNotSupported(t *testing.T) {
	t.Parallel()

	managers := map[string]*darwin.Manager{
		"zero value": {},
		"nil":        nil,
	}

	for name, mgr := range managers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := mgr.SetGridOpacity(0.5)
			if err == nil {
				t.Fatal("SetGridOpacity() = nil, want CodeNotSupported")
			}
			if !derrors.IsNotSupported(err) {
				t.Errorf("SetGridOpacity() = %v (code %q), want CodeNotSupported",
					err, derrors.GetCode(err))
			}
		})
	}
}
