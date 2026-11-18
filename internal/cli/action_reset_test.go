package cli

import "testing"

func TestActionResetCommand_ParsesToggleFlag(t *testing.T) {
	cmd := buildResetCommand()
	flag := cmd.Flags().Lookup("toggle")
	if flag == nil || flag.Value.Type() != "bool" || flag.DefValue != "false" {
		t.Fatal("reset must register --toggle as an optional bool")
	}
	if err := cmd.ParseFlags([]string{"--toggle"}); err != nil {
		t.Fatal(err)
	}
	if enabled, err := cmd.Flags().GetBool("toggle"); err != nil || !enabled {
		t.Fatalf("toggle = %v, error = %v", enabled, err)
	}
}
