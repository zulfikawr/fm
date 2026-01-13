package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleWindowMsg delegates window-related messages to specialized handlers
func HandleWindowMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return actions.ResizeWindow(m, msg)
	}
	return nil
}
