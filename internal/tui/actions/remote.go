package actions

import (
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// FinalizeRemoteConnect handles remote connection result logic
func FinalizeRemoteConnect(m *state.Model, msg commands.RemoteConnectMsg) (tea.Cmd, bool) {
	m.UI.Loading = false
	if msg.Err != nil {
		// If it failed due to authentication, prompt for password/key
		if !m.UI.RemoteAuth {
			m.UI.RemoteAuth = true
			return OpenPrompt(m, state.InputAuth, ""), true
		}
		// Return false so caller can log error
		return nil, false
	}

	// Success! Overwrite current tab
	m.FS = msg.FS
	m.Navigation.Path = msg.Path
	m.Navigation.PathGen++
	m.Navigation.Cursor = 0
	m.Navigation.Offset = 0
	m.Navigation.Items = nil
	m.Navigation.FilteredItems = nil
	return Reload(m), true
}

// SetHostConfirmReq updates the model with a host confirmation request
func SetHostConfirmReq(m *state.Model, msg commands.HostConfirmMsg) {
	m.UI.Loading = false
	m.UI.HostConfirm = true
	m.Remote.HostConfirmReq = msg.Request
}
