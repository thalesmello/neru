//go:build darwin

package darwin

import "testing"

func TestClickSequenceCount_BoundsAtTripleAndRestarts(t *testing.T) {
	count := 0
	for _, want := range []int{1, 2, 3, 1, 2, 3} {
		count = clickSequenceCount(count, true)
		if count != want {
			t.Fatalf("got %d, want %d", count, want)
		}
	}
	if got := clickSequenceCount(2, false); got != 1 {
		t.Fatalf("new button/point/interval continued as %d", got)
	}
	if got := clickSequenceCount(-1, true); got != 1 {
		t.Fatalf("invalid previous count returned %d", got)
	}
}
