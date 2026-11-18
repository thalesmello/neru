package modes

import (
	"go.uber.org/zap"

	"github.com/y3owk1n/neru/internal/domain"
)

// RefreshHintsIfEmpty refreshes without discarding a token or active search.
// Like RefreshHints, it requires an open, scannable hints session.
func (h *Handler) RefreshHintsIfEmpty() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.appState.CurrentMode() == domain.ModeHints && h.hints != nil && h.hints.Context != nil {
		ctx := h.hints.Context
		if ctx.SearchActive() || (ctx.Manager() != nil && ctx.Manager().CurrentInput() != "") {
			return nil
		}
	}

	return h.refreshHints()
}

// refreshHintsAfterBackspace is reached only asynchronously from locked mode
// context. Managers survive sessions, so pointer identity is not a guard.
func (h *Handler) refreshHintsAfterBackspace(session uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.canRefreshHintsAfterBackspace(session) {
		return
	}

	if err := h.refreshHints(); err != nil {
		h.logger.Debug("Hint backspace refresh failed", zap.Error(err))
	}
}

// canRefreshHintsAfterBackspace checks deferred work under h.mu.
func (h *handlerState) canRefreshHintsAfterBackspace(session uint64) bool {
	return h.modeSession == session && h.appState.CurrentMode() == domain.ModeHints &&
		h.hints != nil && h.hints.Context != nil && h.hints.Context.Manager() != nil &&
		h.hints.Context.Manager().CurrentInput() == "" && !h.hints.Context.SearchActive()
}
