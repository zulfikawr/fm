package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleRemoteAuth handles remote authentication input events
func HandleRemoteAuth(msg tea.Msg, m *state.Model) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			actions.ClosePrompt(m)
			m.UI.RemoteAuth = false
			return nil
		case "enter":
			input := m.Inputs.ActiveInput.Value()
			actions.ClosePrompt(m)
			m.UI.RemoteAuth = false
			m.UI.Loading = true

			password := ""
			keyPath := ""

			// Check if input is an existing file path
			if _, err := os.Stat(input); err == nil {
				keyPath = input
			} else {
				password = input
			}

			return tea.Batch(
				commands.ConnectRemote(m.Remote.Host, m.Remote.User, password, keyPath, m.Remote.HostConfirmChan),
				commands.ListenForHostConfirmation(m.Remote.HostConfirmChan),
			)
		}
	}

	m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
	return cmd
}
