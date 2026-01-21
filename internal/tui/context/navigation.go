package context

import (
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"time"
)

// --- Navigation State ---

// NavigationState holds navigation-related state
type NavigationState struct {
	Path           string      // Current directory path
	PathGen        int         // Incremented on each navigation to detect stale messages
	Cursor         int         // Current cursor position
	Offset         int         // Scroll offset
	Items          []core.Item // Original items
	FilteredItems  []core.Item // Filtered items for display
	SelectedCount  int         // Number of selected items
	SelectedPaths  map[string]bool
	FilterTimer    *time.Timer     // Timer for debouncing filter
	FilterGen      int             // Generation counter for filter
	FilterQuery    string          // Current active filter query
	BackHistory    []string        // History for "Back" navigation
	ForwardHistory []string        // History for "Forward" navigation
	ParentFS       core.FileSystem // Previous FS before entering archive
	ParentPath     string          // Previous path before entering archive
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

// SelectAll selects all currently filtered items
func (n *NavigationState) SelectAll() {
	if n.SelectedPaths == nil {
		n.SelectedPaths = make(map[string]bool)
	}
	for i := range n.FilteredItems {
		item := &n.FilteredItems[i]
		if item.IsUp {
			continue
		}
		if !n.SelectedPaths[item.Path] {
			n.SelectedPaths[item.Path] = true
			n.SelectedCount++
		}
		item.Selected = true
	}
}

// ClearSelection clears all selections
func (n *NavigationState) ClearSelection() {
	n.SelectedPaths = make(map[string]bool)
	n.SelectedCount = 0
	for i := range n.Items {
		n.Items[i].Selected = false
	}
	for i := range n.FilteredItems {
		n.FilteredItems[i].Selected = false
	}
}

// IsSelected checks if a path is selected
func (n *NavigationState) IsSelected(path string) bool {
	if n.SelectedPaths == nil {
		return false
	}
	return n.SelectedPaths[path]
}

// --- Tab Management ---

// Tab represents a navigation context
type Tab struct {
	FS             core.FileSystem
	Path           string
	Items          []core.Item
	FilteredItems  []core.Item
	Cursor         int
	Offset         int
	SortMode       sorting.SortMode
	GitBranch      string
	GitRoot        string
	SearchQuery    string
	Searching      bool
	SelectMode     bool
	SelectedPaths  map[string]bool
	RemoteUser     string
	RemoteHost     string
	BackHistory    []string
	ForwardHistory []string
	ParentFS       core.FileSystem // Previous FS before entering archive
	ParentPath     string          // Previous path before entering archive
}

// NewTab creates a new tab for the given path
func NewTab(fs core.FileSystem, path string, sortMode sorting.SortMode, remoteUser, remoteHost string) Tab {
	return Tab{
		FS:            fs,
		Path:          path,
		SortMode:      sortMode,
		SelectedPaths: make(map[string]bool),
		RemoteUser:    remoteUser,
		RemoteHost:    remoteHost,
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
