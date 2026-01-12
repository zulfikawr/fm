package update

import (
	"fm/internal/files/sorting"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleCreateTab handles creating a new tab
func HandleCreateTab(m *state.Model) (tea.Cmd, bool) {
	if len(m.Tabs) >= 9 {
		return commands.SetMsg(m, "Tab limit reached (max 9 tabs)"), true
	}

	// Ensure we have at least one tab before saving state
	if len(m.Tabs) == 0 {
		m.Tabs = []state.Tab{{
			Path:          m.Navigation.Path,
			SortMode:      sorting.SortDefault,
			SelectedPaths: make(map[string]bool),
		}}
		m.ActiveTab = 0
	}

	// Save current tab state BEFORE switching
	actions.SaveTabState(m)
	newTab := state.Tab{
		Path:          m.Navigation.Path,
		SortMode:      sorting.SortDefault,
		SelectedPaths: make(map[string]bool),
	}
	m.Tabs = append(m.Tabs, newTab)
	m.ActiveTab = len(m.Tabs) - 1
	actions.SyncTabToModel(m)
	return actions.Reload(m), true
}

// HandleSwitchTab handles switching to a specific tab
func HandleSwitchTab(m *state.Model, tabNum int) (tea.Cmd, bool) {
	// Ensure we have at least one tab
	if len(m.Tabs) == 0 {
		return nil, false
	}

	// Only switch tabs if there are multiple tabs
	if len(m.Tabs) <= 1 {
		return nil, false
	}

	if tabNum > 0 && tabNum <= len(m.Tabs) {
		// Save current tab state BEFORE switching
		actions.SaveTabState(m)
		// Reset model cursor/offset so syncTabToModel can set new ones clearly
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
		// Switch to new tab
		m.ActiveTab = tabNum - 1
		actions.SyncTabToModel(m)
		return actions.Reload(m), true
	}
	return nil, false
}

// HandleCloseTab handles closing the current tab
func HandleCloseTab(m *state.Model) (tea.Cmd, bool) {
	// Ensure we have at least one tab
	if len(m.Tabs) == 0 {
		return nil, false
	}

	// Close current tab (only if more than one tab)
	if len(m.Tabs) <= 1 {
		return nil, false
	}

	// Validate active tab index before removing
	if m.ActiveTab < 0 || m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = 0
		return nil, false
	}

	// Remove current tab
	m.Tabs = append(m.Tabs[:m.ActiveTab], m.Tabs[m.ActiveTab+1:]...)
	// Adjust active tab index
	if m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = len(m.Tabs) - 1
	}
	actions.SyncTabToModel(m)
	return actions.Reload(m), true
}
