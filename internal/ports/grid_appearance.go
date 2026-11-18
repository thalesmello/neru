package ports

// GridAppearancePort is an optional overlay capability for changing only the
// presentation of an active grid. This operation never exits a mode, resets its
// selection, releases keyboard capture, or reloads configuration. Backends that
// do not implement it are refused with CodeNotSupported by the overlay adapter.
//
// The macOS implementation schedules work on the Cocoa main queue and never
// calls back into the mode handler. It is independent of screen-share hiding.
type GridAppearancePort interface {
	// SetGridOpacity overrides the cell background alpha in Grid and Recursive
	// Grid, including Grid's fine-adjustment layer. It leaves label badges,
	// text, separators, and other modes unchanged. opacity must be in [0, 1].
	// The override lasts until daemon restart; configuration is not written.
	SetGridOpacity(opacity float64) error
}
