package state

import (
	"fm/internal/files"
	"fm/internal/files/sorting"
)

// Tab represents a navigation context
type Tab struct {
	Path          string
	Items         []files.Item
	FilteredItems []files.Item
	Cursor        int
	Offset        int
	SortMode      sorting.SortMode
	GitBranch     string
	GitRoot       string
	SearchQuery   string
	Searching     bool
	SelectMode    bool
	SelectedPaths map[string]bool
}

// NewTab creates a new tab for the given path
func NewTab(path string, sortMode sorting.SortMode) Tab {
	return Tab{
		Path:          path,
		SortMode:      sortMode,
		SelectedPaths: make(map[string]bool),
	}
}

// SelectedCount returns the number of selected items
func (t *Tab) SelectedCount() int {
	return len(t.SelectedPaths)
}

// IsSelected returns true if the given path is selected
func (t *Tab) IsSelected(path string) bool {
	if t.SelectedPaths == nil {
		return false
	}
	return t.SelectedPaths[path]
}
