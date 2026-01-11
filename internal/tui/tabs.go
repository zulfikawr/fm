package tui

import (
	"fm/internal/files"
)

// Tab represents a single tab with its own navigation state
type Tab struct {
	path          string
	items         []files.Item
	filteredItems []files.Item
	cursor        int
	offset        int
	sortMode      files.SortMode
	gitBranch     string
	gitRoot       string
	searching     bool
	searchQuery   string
	selectMode    bool
	selectedPaths map[string]bool
}

// saveTabState saves the current model state to the active tab
func (m *Model) saveTabState() {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		m.tabs[m.activeTab].path = m.path
		m.tabs[m.activeTab].items = m.items
		m.tabs[m.activeTab].filteredItems = m.filteredItems
		m.tabs[m.activeTab].cursor = m.cursor
		m.tabs[m.activeTab].offset = m.offset
		m.tabs[m.activeTab].sortMode = m.sortMode
		m.tabs[m.activeTab].gitBranch = m.gitBranch
		m.tabs[m.activeTab].gitRoot = m.gitRoot
		m.tabs[m.activeTab].searching = m.searching
		m.tabs[m.activeTab].searchQuery = m.searchInput.Value()
		m.tabs[m.activeTab].selectMode = m.selectMode
		m.tabs[m.activeTab].selectedPaths = make(map[string]bool)
		for k, v := range m.selectedPaths {
			m.tabs[m.activeTab].selectedPaths[k] = v
		}
	}
}

// syncTabToModel loads the active tab's state into the model
func (m *Model) syncTabToModel() {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		tab := m.tabs[m.activeTab]
		m.path = tab.path
		m.items = tab.items
		m.filteredItems = tab.filteredItems
		m.cursor = tab.cursor
		m.offset = tab.offset
		m.sortMode = tab.sortMode
		m.gitBranch = tab.gitBranch
		m.gitRoot = tab.gitRoot
		m.searching = tab.searching
		m.searchInput.SetValue(tab.searchQuery)
		if m.searching {
			m.searchInput.Focus()
		} else {
			m.searchInput.Blur()
		}
		m.selectMode = tab.selectMode
		m.selectedPaths = make(map[string]bool)
		for k, v := range tab.selectedPaths {
			m.selectedPaths[k] = v
		}
		m.pathGeneration++
	}
}
