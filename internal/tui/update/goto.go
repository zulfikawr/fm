package update

import (
	"strings"

	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleGoto handles goto input events
func HandleGoto(msg tea.Msg, m *state.Model) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			actions.ClosePrompt(m)
			return nil
		case "enter":
			input := m.Inputs.ActiveInput.Value()
			if input == "" {
				actions.ClosePrompt(m)
				return nil
			}

			// Smart detection: if it contains @ or doesn't look like a local path and contains a dot/hostname
			if strings.Contains(input, "@") || (!strings.HasPrefix(input, "/") && !strings.HasPrefix(input, "./") && !strings.HasPrefix(input, "../") && !strings.HasPrefix(input, "~") && strings.Contains(input, ".")) {
				return handleRemoteGoto(input, m)
			}

			actions.ClosePrompt(m)
			return actions.NavigateToPath(m, input)
		}
	}

	m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)
	return cmd
}

func handleRemoteGoto(input string, m *state.Model) tea.Cmd {
	var user, host string
	if strings.Contains(input, "@") {
		parts := strings.SplitN(input, "@", 2)
		user = parts[0]
		host = parts[1]
	} else {
		host = input
	}

	m.Remote.Host = host
	m.Remote.User = user
	m.UI.Loading = true
	actions.ClosePrompt(m)

	return tea.Batch(
		commands.ConnectRemote(host, user, "", "", m.Remote.HostConfirmChan),
		commands.ListenForHostConfirmation(m.Remote.HostConfirmChan),
	)
}
