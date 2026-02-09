package nav

import (
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func CreateTab(m *tui_context.Model) tea.Cmd {
	if len(m.Tabs) >= 9 {
		return func() tea.Msg { return messages.TabLimitMsg{} }
	}
	SaveTabState(m)
	m.AddTab(m.Navigation.Path)
	m.ActiveTab = len(m.Tabs) - 1
	cmd := SyncTabToModel(m)
	return tea.Batch(cmd, Reload(m, false))
}

func SwitchTab(m *tui_context.Model, tabNum int) tea.Cmd {
	if tabNum > 0 && tabNum <= len(m.Tabs) {
		SaveTabState(m)
		m.ActiveTab = tabNum - 1
		cmd := SyncTabToModel(m)
		return tea.Batch(cmd, Reload(m, false))
	}
	return nil
}

func CloseTab(m *tui_context.Model) tea.Cmd {
	if m.CloseActiveTab() {
		cmd := SyncTabToModel(m)
		return tea.Batch(cmd, Reload(m, false))
	}
	return nil
}

func SaveTabState(m *tui_context.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		m.Navigation.FilterActive = m.UI.InputActive && m.Inputs.Mode == tui_context.InputSearch
		m.Tabs[m.ActiveTab].NavigationState = m.Navigation
	}
}

func SyncTabToModel(m *tui_context.Model) tea.Cmd {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		m.Navigation = m.Tabs[m.ActiveTab].NavigationState
		m.FS = m.Navigation.FS

		var cmd tea.Cmd
		if m.Navigation.FilterActive {
			m.UI.InputActive = true
			m.Inputs.Mode = tui_context.InputSearch
			cmd = m.Inputs.ActiveInput.FocusCmd()
		} else {
			m.UI.InputActive = false
			m.Inputs.Mode = tui_context.InputNone
			m.Inputs.ActiveInput.Blur()
		}

		m.Inputs.ActiveInput.SetValue(m.Navigation.FilterQuery)
		m.Navigation.PathGen++
		return cmd
	}
	return nil
}
