package update

import (
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles the main application update logic, delegating to specialized handlers
func Update(m *state.Model, msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update spinner
	var spinnerCmd tea.Cmd
	m.Display.LoadingSpinner, spinnerCmd = m.Display.LoadingSpinner.Update(msg)
	if spinnerCmd != nil {
		cmds = append(cmds, spinnerCmd)
	}

	// Delegate based on message type
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds = append(cmds, HandleWindowMsg(m, msg))
	case tea.KeyMsg:
		cmds = append(cmds, HandleKeyMsg(m, msg))
	default:
		// Try domain-specific routers
		if cmd := HandleFileSystemMsg(m, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := HandleGitMsg(m, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := HandleRemoteMsg(m, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := HandleOperationsMsg(m, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if cmd := HandleGenericMsg(m, msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}
