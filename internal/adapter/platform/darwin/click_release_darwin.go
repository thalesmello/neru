//go:build darwin

package darwin

import (
	"time"

	"github.com/y3owk1n/neru/internal/derrors"
)

// clickReleaseAttempts bounds recovery if the window server loses an up event.
const (
	clickReleaseAttempts    = 3
	clickReleaseSettleDelay = 10 * time.Millisecond
)

type clickReleaseOps struct {
	isDown       func() bool
	physicalDown func() bool
	postUp       func() error
	pause        func()
	remember     func()
}

// recoverClickRelease never adds another down. A physical button press belongs
// to the user, not to this recovery. Failed synthetic releases remain recorded
// so pause/mode-exit cleanup can retry rather than forgetting a held button.
func recoverClickRelease(ops clickReleaseOps) error {
	for range clickReleaseAttempts {
		if !ops.isDown() || ops.physicalDown() {
			return nil
		}

		err := ops.postUp()
		if err != nil {
			ops.remember()
			return err
		}

		ops.pause()
	}

	if !ops.isDown() || ops.physicalDown() {
		return nil
	}

	ops.remember()

	return derrors.New(
		derrors.CodeActionFailed,
		"mouse button remained down after click release retries",
	)
}
