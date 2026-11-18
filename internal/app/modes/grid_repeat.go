package modes

// SetGridRepeat changes the active grid's confirmation disposition in place.
// No selection, frame, capture, or recursion state is reset.
func (h *Handler) SetGridRepeat(enabled bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	setter, ok := activeModeEffect[repeatSetter](&h.handlerState, extensionRepeatControl)
	return ok && setter.ApplyRepeat(enabled)
}

var _ repeatSetter = (*GridMode)(nil)
var _ repeatSetter = (*RecursiveGridMode)(nil)

func (m *GridMode) ApplyRepeat(enabled bool) bool {
	if m.handler.grid == nil || m.handler.grid.Context == nil {
		return false
	}
	m.handler.grid.Context.SetRepeat(enabled)
	return true
}
func (m *RecursiveGridMode) ApplyRepeat(enabled bool) bool {
	if m.handler.recursiveGrid == nil || m.handler.recursiveGrid.Context == nil {
		return false
	}
	m.handler.recursiveGrid.Context.SetRepeat(enabled)
	return true
}
