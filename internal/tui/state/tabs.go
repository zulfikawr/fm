package state

import (
	"fm/internal/files/core"
	"fm/internal/files/sorting"
)

// Tab represents a navigation context
type Tab struct {
	FS            core.FileSystem
	Path          string
	Items         []core.Item
	FilteredItems []core.Item
	Cursor        int
	Offset        int
	SortMode      sorting.SortMode
	GitBranch     string
	GitRoot       string
	SearchQuery   string
	Searching     bool
	SelectMode    bool
	SelectedPaths map[string]bool
	RemoteUser    string
	RemoteHost    string
}

// NewTab creates a new tab for the given path
func NewTab(fs core.FileSystem, path string, sortMode sorting.SortMode) Tab {
	return Tab{
		FS:            fs,
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

// AddTab creates and appends a new tab to the model
func (m *Model) AddTab(path string) {
	if len(m.Tabs) >= 9 {
		return
	}
	newTab := NewTab(m.FS, path, sorting.SortDefault)
	newTab.RemoteUser = m.Remote.User
	newTab.RemoteHost = m.Remote.Host
	m.Tabs = append(m.Tabs, newTab)
	m.ActiveTab = len(m.Tabs) - 1
}

// CloseActiveTab removes the current tab and adjusts the active index
func (m *Model) CloseActiveTab() bool {
	if len(m.Tabs) <= 1 {
		return false
	}
	if m.ActiveTab < 0 || m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = 0
		return false
	}

	m.Tabs = append(m.Tabs[:m.ActiveTab], m.Tabs[m.ActiveTab+1:]...)
	if m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = len(m.Tabs) - 1
	}
	return true
}

// SwitchTab switches to the specified tab number (1-based)
func (m *Model) SwitchTab(tabNum int) bool {
	if tabNum > 0 && tabNum <= len(m.Tabs) {
		m.ActiveTab = tabNum - 1
		return true
	}
	return false
}
