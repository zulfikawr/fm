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
		case "tab":
			m.Inputs.AltMode = !m.Inputs.AltMode
			if m.Inputs.AltMode {
				m.Inputs.ActiveInput.Placeholder = "path or [user@]host[:port]"
			} else {
				m.Inputs.ActiveInput.Placeholder = "path"
			}
			return nil

		case "enter":
			input := m.Inputs.ActiveInput.Value()

			// If we are currently on a remote filesystem
			if !m.FS.IsLocal() {
				if m.Inputs.AltMode { // AltMode true means Local mode when on Remote FS
					actions.ClosePrompt(m)
					return actions.SwitchToLocal(m, input)
				}

				// Check if they want to navigate the current remote or connect to a new one.
				// A path starts with /, ., ~, or is empty. A connection string contains @.
				isPath := strings.HasPrefix(input, "/") || strings.HasPrefix(input, ".") || strings.HasPrefix(input, "~") || input == ""
				isConnection := strings.Contains(input, "@")

				if isPath && !isConnection {
					actions.ClosePrompt(m)
					return actions.NavigateToPath(m, input)
				}

				return handleRemoteGoto(input, m)
			}

			// Currently on local filesystem
			isRemote := m.Inputs.AltMode
			if !isRemote {
				// Auto-detect remote connection string
				isRemote = strings.Contains(input, "@") || (!strings.HasPrefix(input, "/") && !strings.HasPrefix(input, "./") && !strings.HasPrefix(input, "../") && !strings.HasPrefix(input, "~") && strings.Contains(input, "."))
			}

			if isRemote {
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
