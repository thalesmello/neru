package ports

import "time"

// GridPeekPort is an optional visual mask for the active grid surface. It
// preserves mode, selection, rendering and keyboard capture. Another peek
// restarts the deadline; leaving the surface cancels restoration. It is not
// screen-share exclusion and never reloads configuration.
type GridPeekPort interface {
	PeekGrid(duration time.Duration) error
}
