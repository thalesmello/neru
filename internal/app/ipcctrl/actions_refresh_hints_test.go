package ipcctrl

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/adapter/ipc"
)

func TestParseActionArgs_RefreshHintsIfInputEmpty(t *testing.T) {
	parsed, parseErr := parseActionArgs([]string{flagIfInputEmpty})
	if parseErr || !parsed.ifInputEmpty || !parsed.present[flagIfInputEmpty] {
		t.Fatal("--if-input-empty was not parsed and recorded")
	}
	for name := range actionFlagSupport {
		rejected := rejectUnsupportedFlags(name, parsed) != nil
		if rejected == (name == "refresh_hints") {
			t.Errorf("%s: rejection = %v", name, rejected)
		}
	}
}

func TestHandleAction_RefreshHintsIfInputEmpty_ValidatesFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code string
	}{
		{"accepted", []string{"refresh_hints", flagIfInputEmpty}, ipc.CodeActionFailed},
		{"wrong action", []string{"backspace", flagIfInputEmpty}, ipc.CodeInvalidInput},
		{"no value accepted", []string{"refresh_hints", flagIfInputEmpty + "=true"}, ipc.CodeInvalidInput},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &ActionsHandler{logger: zap.NewNop()}
			resp := h.handleAction(context.Background(), ipc.Command{Action: ActionCommand, Args: tc.args})
			if resp.Success || resp.Code != tc.code {
				t.Fatalf("response = %+v, want failure code %s", resp, tc.code)
			}
		})
	}
}
