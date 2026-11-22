//go:build darwin

package darwin

import (
	"testing"
	"time"
)

func TestMouseEventWait_HonorsRequestedDuration(t *testing.T) {
	for _, duration := range []time.Duration{8 * time.Millisecond, 10 * time.Millisecond} {
		t.Run(duration.String(), func(t *testing.T) {
			started := time.Now()

			waitForMouseEvents(duration)

			if elapsed := time.Since(started); elapsed < duration {
				t.Fatalf(
					"mouse event wait returned after %v, before requested %v",
					elapsed,
					duration,
				)
			}
		})
	}
}
