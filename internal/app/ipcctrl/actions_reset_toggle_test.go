package ipcctrl

import "testing"

func TestResetToggleFlag_IsAcceptedOnlyBySupportedActions(t *testing.T) {
	parsed, bad := parseActionArgs([]string{"--toggle"})
	if bad || !parsed.useToggle {
		t.Fatal("reset toggle flag was not parsed")
	}
	if response := rejectUnsupportedFlags("reset", parsed); response != nil {
		t.Fatalf("reset rejected toggle: %+v", response)
	}
	if response := rejectUnsupportedFlags("backspace", parsed); response == nil {
		t.Fatal("backspace unexpectedly accepted reset toggle")
	}
}
