package app

import (
	tui_context "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleHelp handles help-related messages
func HandleHelp(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.ActiveView == tui_context.ViewHelp {
			return handleHelpKeys(m, msg)
		}
	}
	return nil
}

func handleHelpKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	totalItems := 34 // Total items across all sections

	switch msg.String() {
	case "up", "k":
		if m.Help.Cursor > 0 {
			m.Help.Cursor--
		}
	case "down", "j":
		if m.Help.Cursor < totalItems-1 {
			m.Help.Cursor++
		}
	case "esc", "q":
		m.UI.ToggleHelp()
	default:
		// Check for custom help toggle key
		for i := range m.Config.Keybindings {
			kb := m.Config.Keybindings[i]
			if kb.Action == "help" {
				for j := range kb.Keys {
					if kb.Keys[j] == msg.String() {
						m.UI.ToggleHelp()
						return nil
					}
				}
			}
		}
	}

	m.Help.Offset = ScrollHelp(m)
	return nil
}

// ScrollHelp recalculates the help view offset
func ScrollHelp(m *tui_context.Model) int {
	cursor := m.Help.Cursor
	offset := m.Help.Offset
	height := m.Display.ViewportHeight

	// Roughly map cursor to rendered rows (including headers and spacers)
	// This is an approximation
	rowIdx := 0
	if cursor <= 6 { // Navigation (7 items)
		rowIdx = cursor + 2
	} else if cursor <= 10 { // Selection (4 items)
		rowIdx = cursor + 4
	} else if cursor <= 13 { // Tabs (3 items)
		rowIdx = cursor + 6
	} else if cursor <= 21 { // File Ops (8 items)
		rowIdx = cursor + 8
	} else if cursor <= 26 { // Search (5 items)
		rowIdx = cursor + 10
	} else { // Misc (7 items)
		rowIdx = cursor + 12
	}

	if rowIdx < offset {
		// When moving up, if we hit the first item of a section,
		// scroll up one more to show the header
		if cursor == 0 || cursor == 7 || cursor == 11 || cursor == 14 || cursor == 22 || cursor == 27 {
			return rowIdx - 1
		}
		return rowIdx
	}

	if rowIdx >= offset+height {
		return rowIdx - height + 1
	}

	return offset
}
