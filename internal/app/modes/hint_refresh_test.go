package modes

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/y3owk1n/neru/internal/app/components"
	hintscomponent "github.com/y3owk1n/neru/internal/app/components/hints"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/hint"
	"github.com/y3owk1n/neru/internal/domain/modecmd"
	"github.com/y3owk1n/neru/internal/domain/state"
)

// Supplied targets make attempted refreshes observable without querying AX.
func newHintRefreshFixture() (*Handler, *observer.ObservedLogs) {
	core, logs := observer.New(zap.DebugLevel)
	h := newHandlerWithState(handlerState{
		appState:    state.NewAppState(),
		logger:      zap.New(core),
		hints:       &components.HintsComponent{Context: &hintscomponent.Context{}},
		modeSession: 7,
	})
	h.appState.SetMode(domain.ModeHints)
	h.hints.Context.SetManager(hint.NewManager(zap.NewNop(), &h.mu))
	h.hints.Context.SetTargets([]modecmd.Target{{}})

	return h, logs
}

func TestHintsMode_Backspace_RemovesTokenWithoutRefresh(t *testing.T) {
	h, logs := newHintRefreshFixture()
	h.mu.Lock()
	h.hints.Context.Manager().SetCurrentInput("AB")
	NewHintsMode(&h.handlerState).Backspace()
	if got := h.hints.Context.Manager().CurrentInput(); got != "A" {
		t.Errorf("input = %q, want A", got)
	}
	h.mu.Unlock()
	if logs.Len() != 0 {
		t.Fatal("nonempty backspace attempted refresh")
	}
}

func TestHintsMode_Backspace_EmptyRefreshesAsynchronously(t *testing.T) {
	h, logs := newHintRefreshFixture()
	h.mu.Lock()
	NewHintsMode(&h.handlerState).Backspace()
	if logs.Len() != 0 {
		t.Fatal("refresh ran while the handler lock was held")
	}
	h.mu.Unlock()
	deadline := time.After(2 * time.Second)
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for logs.Len() == 0 {
		select {
		case <-deadline:
			t.Fatal("deferred refresh never ran")
		case <-tick.C:
		}
	}
}

func TestHandler_RefreshHintsAfterBackspace_GuardsDeferredWork(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Handler)
	}{
		{"stale session with reused manager", func(h *Handler) { h.modeSession++ }},
		{"left hints", func(h *Handler) { h.appState.SetMode(domain.ModeIdle) }},
		{"started typing", func(h *Handler) { h.hints.Context.Manager().SetCurrentInput("A") }},
		{"active search", func(h *Handler) { h.hints.Context.SetSearchActive(true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, logs := newHintRefreshFixture()
			tc.change(h)
			h.refreshHintsAfterBackspace(7)
			if logs.Len() != 0 {
				t.Fatal("guarded callback attempted refresh")
			}
		})
	}
}

func TestHandler_RefreshHintsIfEmpty_SkipsTypingAndSearch(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     string
		search    bool
		wantError bool
	}{
		{"typing", "AB", false, false},
		{"search even without query", "", true, false},
		{"empty refuses supplied targets", "", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := newHintRefreshFixture()
			h.hints.Context.Manager().SetCurrentInput(tc.input)
			h.hints.Context.SetSearchActive(tc.search)
			if err := h.RefreshHintsIfEmpty(); (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tc.wantError)
			}
			if h.hints.Context.Manager().CurrentInput() != tc.input ||
				h.hints.Context.SearchActive() != tc.search {
				t.Fatal("background refresh discarded input or search")
			}
		})
	}
}
