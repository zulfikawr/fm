package actions

import (
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// CreateTab handles creating a new tab
func CreateTab(m *state.Model) (tea.Cmd, bool) {
	if len(m.Tabs) >= 9 {
		return commands.SetMsg(m, "Tab limit reached (max 9 tabs)"), true
	}

	// Save current tab state BEFORE switching
	SaveTabState(m)
	m.AddTab(m.Navigation.Path)
	SyncTabToModel(m)
	return Reload(m), true
}

// SwitchTab handles switching to a specific tab
func SwitchTab(m *state.Model, tabNum int) (tea.Cmd, bool) {
	// Save current tab state BEFORE switching
	SaveTabState(m)
	if m.SwitchTab(tabNum) {
		SyncTabToModel(m)
		return Reload(m), true
	}
	return nil, false
}

// CloseTab handles closing the current tab
func CloseTab(m *state.Model) (tea.Cmd, bool) {
	if m.CloseActiveTab() {
		SyncTabToModel(m)
		return Reload(m), true
	}
	return nil, false
}

// SaveTabState saves the current model state to the active tab
func SaveTabState(m *state.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		m.Tabs[m.ActiveTab].FS = m.FS
		m.Tabs[m.ActiveTab].Path = m.Navigation.Path
		m.Tabs[m.ActiveTab].Items = m.Navigation.Items
		m.Tabs[m.ActiveTab].FilteredItems = m.Navigation.FilteredItems
		m.Tabs[m.ActiveTab].Cursor = m.Navigation.Cursor
		m.Tabs[m.ActiveTab].Offset = m.Navigation.Offset
		m.Tabs[m.ActiveTab].SortMode = m.Display.SortMode
		m.Tabs[m.ActiveTab].GitBranch = m.Git.Branch
		m.Tabs[m.ActiveTab].GitRoot = m.Git.Root
		m.Tabs[m.ActiveTab].Searching = m.UI.InputActive && m.Inputs.Mode == state.InputSearch
		m.Tabs[m.ActiveTab].SearchQuery = m.Inputs.ActiveInput.Value()
		m.Tabs[m.ActiveTab].SelectMode = m.UI.SelectMode
		m.Tabs[m.ActiveTab].RemoteUser = m.Remote.User
		m.Tabs[m.ActiveTab].RemoteHost = m.Remote.Host
		m.Tabs[m.ActiveTab].SelectedPaths = make(map[string]bool)
		for k, v := range m.Navigation.SelectedPaths {
			m.Tabs[m.ActiveTab].SelectedPaths[k] = v
		}
	}
}

// SyncTabToModel loads the active tab's state into the model
func SyncTabToModel(m *state.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		tab := m.Tabs[m.ActiveTab]
		m.FS = tab.FS
		m.Navigation.Path = tab.Path
		m.Navigation.Items = tab.Items
		m.Navigation.FilteredItems = tab.FilteredItems
		m.Navigation.Cursor = tab.Cursor
		m.Navigation.Offset = tab.Offset
		m.Display.SortMode = tab.SortMode
		m.Git.Branch = tab.GitBranch
		m.Git.Root = tab.GitRoot
		m.Remote.User = tab.RemoteUser
		m.Remote.Host = tab.RemoteHost

		if tab.Searching {
			m.UI.InputActive = true
			m.Inputs.Mode = state.InputSearch
			m.Inputs.ActiveInput.Focus()
			m.Inputs.ActiveInput.Prompt = "/ "
		} else {
			m.UI.InputActive = false
			m.Inputs.Mode = state.InputNone
			m.Inputs.ActiveInput.Blur()
		}

		m.Inputs.ActiveInput.SetValue(tab.SearchQuery)
		m.UI.SelectMode = tab.SelectMode
		m.Navigation.SelectedPaths = make(map[string]bool)
		for k, v := range tab.SelectedPaths {
			m.Navigation.SelectedPaths[k] = v
		}
		m.Navigation.PathGen++
	}
}
