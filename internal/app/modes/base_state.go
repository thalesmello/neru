package modes

var _ baseStateReporter = (*GridMode)(nil)
var _ baseStateReporter = (*RecursiveGridMode)(nil)

// AtBase reports an untouched coarse grid. A root-level virtual selection,
// including one placed by nudge, is input that still needs to be reset.
func (m *GridMode) AtBase() bool {
	grid := m.handler.grid
	if grid == nil || grid.Manager == nil || grid.Context == nil {
		return false
	}

	_, selected := grid.Context.SelectionPoint()

	return grid.Manager.CurrentInput() == "" && !grid.Manager.InSubgrid() && !selected
}

// AtBase reports an untouched root recursive grid. Its initial center selection
// is normal; only typed input, zoom history or a final cell makes it dirty.
func (m *RecursiveGridMode) AtBase() bool {
	grid := m.handler.recursiveGrid
	if grid == nil || grid.Manager == nil || grid.Manager.CurrentGrid() == nil {
		return false
	}

	return grid.Manager.CurrentDepth() == 0 && grid.Manager.CurrentInput() == "" &&
		!grid.Manager.CurrentGrid().HasFinalCell()
}
