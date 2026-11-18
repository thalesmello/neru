//go:build darwin

package darwin

/*
#include "click_sequence.h"
*/
import "C"

// clickSequenceCount exercises the exact helper used by native mouse events.
// It posts no events and needs no desktop permissions.
func clickSequenceCount(previous int, continuation bool) int {
	return int(C.NeruClickSequenceCount(C.int(previous), C.int(boolToClickInt(continuation))))
}
func boolToClickInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
