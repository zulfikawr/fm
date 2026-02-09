package app

import (
	tui_context "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleLogs handles log-related messages
func HandleLogs(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.ActiveView == tui_context.ViewLogs {
			return handleLogKeys(m, msg)
		}
	}
	return nil
}

func handleLogKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "alt+l":
		m.UI.ToggleLogs()
		return nil

	case "up", "k":
		if m.Logs.Cursor > 0 {
			m.Logs.Cursor--
		}
		m.Logs.Offset = ScrollLogs(m.Logs.Cursor, m.Logs.Offset, m.Display.ViewportHeight)

	case "down", "j":
		if m.Logs.Cursor < len(m.Logs.Entries)-1 {
			m.Logs.Cursor++
		}
		m.Logs.Offset = ScrollLogs(m.Logs.Cursor, m.Logs.Offset, m.Display.ViewportHeight)
	}

	return nil
}

// ScrollLogs recalculates the logs view offset
func ScrollLogs(cursor, offset, viewportHeight int) int {
	if viewportHeight <= 0 {
		return 0
	}
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+viewportHeight {
		return cursor - viewportHeight + 1
	}
	return offset
}
