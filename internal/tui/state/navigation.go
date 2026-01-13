package state

import "fm/internal/files/core"

// NavigationState holds navigation-related state
type NavigationState struct {
	Path          string      // Current directory path
	PathGen       int         // Incremented on each navigation to detect stale messages
	Cursor        int         // Current cursor position
	Offset        int         // Scroll offset
	Items         []core.Item // Original items
	FilteredItems []core.Item // Filtered items for display
	SelectedCount int         // Number of selected items
	SelectedPaths map[string]bool
}

// Select adds a path to the selection
func (n *NavigationState) Select(path string) {
	if n.SelectedPaths == nil {
		n.SelectedPaths = make(map[string]bool)
	}
	if !n.SelectedPaths[path] {
		n.SelectedPaths[path] = true
		n.SelectedCount++
	}
}

// Deselect removes a path from the selection
func (n *NavigationState) Deselect(path string) {
	if n.SelectedPaths != nil && n.SelectedPaths[path] {
		delete(n.SelectedPaths, path)
		n.SelectedCount--
	}
}

// ToggleSelection toggles the selection state of a path
func (n *NavigationState) ToggleSelection(path string) {
	if n.IsSelected(path) {
		n.Deselect(path)
	} else {
		n.Select(path)
	}
}

// ClearSelection clears all selections
func (n *NavigationState) ClearSelection() {
	n.SelectedPaths = make(map[string]bool)
	n.SelectedCount = 0
	for i := range n.Items {
		n.Items[i].Selected = false
	}
}

// IsSelected checks if a path is selected
func (n *NavigationState) IsSelected(path string) bool {
	if n.SelectedPaths == nil {
		return false
	}
	return n.SelectedPaths[path]
}

// ClearSelection clears selection state across navigation and UI
func (m *Model) ClearSelection() {
	m.Navigation.ClearSelection()
	m.UI.SelectMode = false
}
