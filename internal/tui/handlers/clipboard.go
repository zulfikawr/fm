package handlers

import (
	tui_context "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleClipboard handles clipboard-related messages
func HandleClipboard(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.ClipboardOpen {
			return handleClipboardKeys(m, msg)
		}
	}
	return nil
}

func handleClipboardKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "alt+c":
		m.UI.ToggleClipboard()
		return nil

	case "up", "k":
		if m.Operations.Clipboard.Cursor > 0 {
			m.Operations.Clipboard.Cursor--
		}
		m.Operations.Clipboard.Offset = scrollClipboard(m.Operations.Clipboard.Cursor, m.Operations.Clipboard.Offset, m.Display.ViewportHeight)

	case "down", "j":
		if m.Operations.Clipboard.Cursor < len(m.Operations.Clipboard.Paths)-1 {
			m.Operations.Clipboard.Cursor++
		}
		m.Operations.Clipboard.Offset = scrollClipboard(m.Operations.Clipboard.Cursor, m.Operations.Clipboard.Offset, m.Display.ViewportHeight)

	case "backspace", "d", "x":
		// Remove item from clipboard
		if len(m.Operations.Clipboard.Paths) > 0 {
			idx := m.Operations.Clipboard.Cursor
			m.Operations.Clipboard.Paths = append(m.Operations.Clipboard.Paths[:idx], m.Operations.Clipboard.Paths[idx+1:]...)
			if m.Operations.Clipboard.Cursor >= len(m.Operations.Clipboard.Paths) && m.Operations.Clipboard.Cursor > 0 {
				m.Operations.Clipboard.Cursor--
			}
		}
		return nil

	case "v":
		// Trigger paste from clipboard view
		if len(m.Operations.Clipboard.Paths) > 0 {
			m.UI.ClipboardOpen = false
			return performPaste(m)
		}
	}

	return nil
}

func scrollClipboard(cursor, offset, viewportHeight int) int {
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
