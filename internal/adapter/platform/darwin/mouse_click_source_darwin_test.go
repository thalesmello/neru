//go:build darwin

package darwin

import (
	"os"
	"strings"
	"testing"
)

func TestNativeClick_ButtonPairCannotPumpRunLoopOrInterleavePointerEvents(t *testing.T) {
	data, err := os.ReadFile("accessibility_mouse_darwin.m")
	if err != nil {
		t.Fatal(err)
	}

	source := string(data)

	start := strings.Index(source, "static int performClickAtPosition(")
	if start < 0 {
		t.Fatal("native click implementation not found")
	}

	end := strings.Index(source[start:], "/// State tracking for click detection")
	if end < 0 {
		t.Fatal("native click implementation end not found")
	}

	click := source[start : start+end]
	steps := []string{
		"os_unfair_lock_lock(&mouseMoveLock);",
		"postPositionedEvent(down, pos);",
		"NeruWaitForMouseEvents(kNeruMouseClickDownUpDelay);",
		"postPositionedEvent(up, pos);",
		"os_unfair_lock_unlock(&mouseMoveLock);",
	}

	previous := -1
	for _, step := range steps {
		index := strings.Index(click, step)
		if index <= previous {
			t.Fatalf("click pair missing/out of order: %s", step)
		}

		previous = index
	}

	pairStart := strings.Index(click, steps[0])

	pairEnd := strings.Index(click, steps[len(steps)-1])
	if strings.Contains(click[pairStart:pairEnd], "CFRunLoopRunInMode") {
		t.Fatal("mouse-down interval re-enters the run loop")
	}
}
