package update

import (
	"fm/internal/constants"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleOperationsMsg delegates operation-related messages to specialized handlers
func HandleOperationsMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.ProgressMsg:
		actions.UpdateProgress(m, msg.Label, msg.Percent)
		return commands.ListenToProgress(msg.Channel)

	case commands.OperationFinishedMsg:
		actions.FinalizeOperation(m, msg.Paths)
		return actions.Reload(m)

	case commands.ConflictMsg:
		HandleConflict(m, msg)
		return nil
	}
	return nil
}

// HandleRenaming handles renaming events
func HandleRenaming(msg tea.Msg, m *state.Model) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			actions.ClosePrompt(m)
			return nil
		case "enter":
			newName := m.Inputs.ActiveInput.Value()
			cmd = actions.PerformRename(m, newName)
			actions.ClosePrompt(m)
			return cmd
		}
	}
	m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
	return cmd
}

// HandleConfirming handles confirmation events
func HandleConfirming(msg tea.Msg, m *state.Model) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			var cmds []tea.Cmd
			switch m.Operations.ActionType {
			case constants.ActionDelete:
				cmds = append(cmds, actions.PerformDelete(m)...)
			case constants.ActionPaste:
				cmds = append(cmds, actions.PerformPaste(m)...)
			case constants.ActionConflict:
				return actions.ResolveConflict("overwrite", m)
			case constants.ActionCancel:
				if m.Operations.CancelFunc != nil {
					m.Operations.CancelFunc()
					m.Operations.CancelFunc = nil
				}
				m.Operations.Progress.Hide()
			}
			m.UI.Confirming = false
			return tea.Batch(cmds...)
		case "n", "N", "esc":
			if m.Operations.ActionType == constants.ActionConflict {
				return actions.ResolveConflict("skip", m)
			}
			m.UI.Confirming = false
			return nil
		case "r", "R":
			if m.Operations.ActionType == constants.ActionConflict {
				return actions.ResolveConflict("rename", m)
			}
		}
	}
	return nil
}
