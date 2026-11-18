//go:build darwin

package darwin

import "github.com/y3owk1n/neru/internal/adapter/overlay/manager"

var _ manager.GridSurfaceReplacer = (*Manager)(nil)

// PrepareGridReplacement forgets the diff cache without putting an empty
// frame on screen. DrawGrid replaces the fine-grid cells in one transaction.
func (m *Manager) PrepareGridReplacement() {
	if m != nil && m.GridOverlay() != nil {
		m.GridOverlay().InvalidateGridState()
	}
}
