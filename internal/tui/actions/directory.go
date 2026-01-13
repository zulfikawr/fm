package actions

import (
	"fm/internal/files/core"
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// FinalizeDirectoryLoad handles the logic after directory items are loaded
func FinalizeDirectoryLoad(m *state.Model, msg commands.LoadedItemsMsg) (tea.Cmd, bool) {
	m.UI.Loading = false
	if msg.Generation != m.Navigation.PathGen {
		return nil, false
	}

	if msg.Err != nil {
		// Wrap error with context for better error handling
		err := tuierrors.SystemError("load directory", msg.Err).
			WithContext("path", msg.Path).
			WithContext("generation", msg.Generation)

		cmd := LogError(m, err, "failed to load directory")
		m.Navigation.Items = []core.Item{}

		// If we failed to load the current path, try going back
		if m.Navigation.Path == msg.Path {
			parent := m.FS.Dir(m.Navigation.Path)
			if parent != m.Navigation.Path {
				m.Navigation.Path = parent
				m.Navigation.PathGen++
				return tea.Batch(cmd, Reload(m)), true
			}
		}
		return cmd, true
	}

	m.Message.Error = nil
	m.Navigation.Items = msg.Items
	m.Git.Branch = msg.GitBranch
	m.Git.Root = msg.GitRoot
	m.Display.ReadOnly = msg.IsReadOnly

	// Pre-calculate formatted strings
	for i := range m.Navigation.Items {
		m.Navigation.Items[i].UpdateFormatting(m.Config.SizeFormatIndex, m.Config.DateFormatIndex)
	}

	// Restore selection
	m.Navigation.SelectedCount = 0
	if len(m.Navigation.SelectedPaths) > 0 {
		for i := range m.Navigation.Items {
			if m.Navigation.IsSelected(m.Navigation.Items[i].Path) {
				m.Navigation.Items[i].Selected = true
				m.Navigation.SelectedCount++
			}
		}
	}

	m.UI.SelectMode = m.Navigation.SelectedCount > 0

	// Restore cursor/offset from tab or memory
	if m.ActiveTab < len(m.Tabs) {
		tab := m.Tabs[m.ActiveTab]
		if tab.Path == m.Navigation.Path {
			m.Navigation.Cursor = tab.Cursor
			m.Navigation.Offset = tab.Offset
		}
	}

	if m.Navigation.Cursor == 0 {
		if val, ok := m.Cache.CursorMemory.Get(m.Navigation.Path); ok {
			m.Navigation.Cursor = val
		}
	}
	if m.Navigation.Offset == 0 {
		if val, ok := m.Cache.OffsetMemory.Get(m.Navigation.Path); ok {
			m.Navigation.Offset = val
		}
	}

	// Bounds check cursor and offset after loading items
	if len(m.Navigation.Items) > 0 {
		if m.Navigation.Cursor >= len(m.Navigation.Items) {
			m.Navigation.Cursor = len(m.Navigation.Items) - 1
		}
		if m.Navigation.Cursor < 0 {
			m.Navigation.Cursor = 0
		}
		// Ensure offset doesn't push view below visible items
		if m.Navigation.Offset >= len(m.Navigation.Items) {
			m.Navigation.Offset = 0
		}
		if m.Navigation.Offset < 0 {
			m.Navigation.Offset = 0
		}
	} else {
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
	}

	return nil, false
}
