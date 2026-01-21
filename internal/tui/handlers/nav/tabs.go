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
		t := &m.Tabs[m.ActiveTab]
		t.FS = m.FS
		t.Path = m.Navigation.Path
		t.Items = m.Navigation.Items
		t.FilteredItems = m.Navigation.FilteredItems
		t.Cursor = m.Navigation.Cursor
		t.Offset = m.Navigation.Offset
		t.SortMode = m.Display.SortMode
		t.GitBranch = m.Git.Branch
		t.GitRoot = m.Git.Root
		t.Searching = m.UI.InputActive && m.Inputs.Mode == tui_context.InputSearch
		t.SearchQuery = m.Inputs.ActiveInput.Value()
		t.SelectMode = m.UI.SelectMode
		t.RemoteUser = m.Remote.User
		t.RemoteHost = m.Remote.Host
		t.BackHistory = make([]string, len(m.Navigation.BackHistory))
		copy(t.BackHistory, m.Navigation.BackHistory)
		t.ForwardHistory = make([]string, len(m.Navigation.ForwardHistory))
		copy(t.ForwardHistory, m.Navigation.ForwardHistory)
		t.SelectedPaths = make(map[string]bool)
		for k, v := range m.Navigation.SelectedPaths {
			t.SelectedPaths[k] = v
		}
		t.ParentFS = m.Navigation.ParentFS
		t.ParentPath = m.Navigation.ParentPath
	}
}

func SyncTabToModel(m *tui_context.Model) tea.Cmd {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		t := m.Tabs[m.ActiveTab]
		m.FS = t.FS
		m.Navigation.Path = t.Path
		m.Navigation.Items = t.Items
		m.Navigation.FilteredItems = t.FilteredItems
		m.Navigation.Cursor = t.Cursor
		m.Navigation.Offset = t.Offset
		m.Display.SortMode = t.SortMode
		m.Git.Branch = t.GitBranch
		m.Git.Root = t.GitRoot
		m.Remote.User = t.RemoteUser
		m.Remote.Host = t.RemoteHost
		m.Navigation.BackHistory = make([]string, len(t.BackHistory))
		copy(m.Navigation.BackHistory, t.BackHistory)
		m.Navigation.ForwardHistory = make([]string, len(t.ForwardHistory))
		copy(m.Navigation.ForwardHistory, t.ForwardHistory)
		m.Navigation.ParentFS = t.ParentFS
		m.Navigation.ParentPath = t.ParentPath

		var cmd tea.Cmd
		if t.Searching {
			m.UI.InputActive = true
			m.Inputs.Mode = tui_context.InputSearch
			cmd = m.Inputs.ActiveInput.FocusCmd()
		} else {
			m.UI.InputActive = false
			m.Inputs.Mode = tui_context.InputNone
			m.Inputs.ActiveInput.Blur()
		}

		m.Inputs.ActiveInput.SetValue(t.SearchQuery)
		m.UI.SelectMode = t.SelectMode
		m.Navigation.SelectedPaths = make(map[string]bool)
		for k, v := range t.SelectedPaths {
			m.Navigation.SelectedPaths[k] = v
		}
		m.Navigation.PathGen++
		return cmd
	}
	return nil
}
