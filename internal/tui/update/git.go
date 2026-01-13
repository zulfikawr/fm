package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleGitMsg delegates git-related messages to specialized handlers
func HandleGitMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.GitStatusMsg:
		actions.ApplyGitStatus(m, msg)
		filter.Apply(m)
		return nil
	}
	return nil
}
