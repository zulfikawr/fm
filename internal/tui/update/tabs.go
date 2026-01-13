package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleCreateTab handles creating a new tab
func HandleCreateTab(m *state.Model) (tea.Cmd, bool) {
	return actions.CreateTab(m)
}

// HandleSwitchTab handles switching to a specific tab
func HandleSwitchTab(m *state.Model, tabNum int) (tea.Cmd, bool) {
	return actions.SwitchTab(m, tabNum)
}

// HandleCloseTab handles closing the current tab
func HandleCloseTab(m *state.Model) (tea.Cmd, bool) {
	return actions.CloseTab(m)
}
