//go:build darwin

package darwin

/*
#include "mouse_event_wait.h"
*/
import "C"

import "time"

// waitForMouseEvents exercises the same wait used by the native click bridge.
// It posts no input events and requires no desktop permissions.
func waitForMouseEvents(duration time.Duration) {
	C.NeruWaitForMouseEvents(C.double(duration.Seconds()))
}
