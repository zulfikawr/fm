package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleGenericMsg handles generic utility messages
func HandleGenericMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.ClearMsg:
		commands.ClearMsgState(m)
		return nil

	case commands.ErrorMsg:
		// Clear all processing items on error for safety
		m.Operations.ProcessingItems = make(map[string]bool)
		return tea.Batch(actions.LogError(m, msg.Err, "Operation failed"), actions.Reload(m))
	}
	return nil
}
