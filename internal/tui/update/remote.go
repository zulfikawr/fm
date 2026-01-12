package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleRemoteMsg delegates remote-related messages to specialized handlers
func HandleRemoteMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.RemoteConnectMsg:
		cmd := HandleRemoteConnect(m, msg)
		if cmd == nil && msg.Err != nil {
			err := tuierrors.TransientError("remote connection", msg.Err.Error(), 3)
			return actions.LogError(m, err, "Remote connection failed")
		}
		return cmd

	case commands.HostConfirmMsg:
		HandleHostConfirm(m, msg)
		return nil
	}
	return nil
}

// HandleRemoteConnect handles remote connection result messages
func HandleRemoteConnect(m *state.Model, msg commands.RemoteConnectMsg) tea.Cmd {
	m.UI.Loading = false
	if msg.Err != nil {
		// If it failed due to authentication, prompt for password/key
		if !m.UI.RemoteAuth {
			m.UI.RemoteAuth = true
			return actions.OpenPrompt(m, state.InputAuth, "")
		}
		// Return error log command - will be handled by caller
		return nil
	}

	// Success! Overwrite current tab
	m.FS = msg.FS
	m.Navigation.Path = msg.Path
	m.Navigation.PathGen++
	m.Navigation.Cursor = 0
	m.Navigation.Offset = 0
	m.Navigation.Items = nil
	m.Navigation.FilteredItems = nil
	return actions.Reload(m)
}

// HandleHostConfirm handles SSH host key confirmation messages
func HandleHostConfirm(m *state.Model, msg commands.HostConfirmMsg) {
	m.UI.Loading = false
	m.UI.HostConfirm = true
	m.Remote.HostConfirmReq = msg.Request
}
