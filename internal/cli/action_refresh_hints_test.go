package cli

import "testing"

func TestActionRefreshHintsCommand_RegistersIdleOnlyFlag(t *testing.T) {
	cmd := buildRefreshHintsCommand()
	flag := cmd.Flags().Lookup("if-input-empty")
	if flag == nil || flag.Value.Type() != "bool" || flag.DefValue != "false" {
		t.Fatal("refresh_hints must register --if-input-empty as an optional bool")
	}
	if err := cmd.ParseFlags([]string{"--if-input-empty"}); err != nil {
		t.Fatal(err)
	}
	if enabled, err := cmd.Flags().GetBool("if-input-empty"); err != nil || !enabled {
		t.Fatalf("if-input-empty = %v, error = %v", enabled, err)
	}
}
