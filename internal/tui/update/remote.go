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
		cmd, handled := actions.FinalizeRemoteConnect(m, msg)
		if !handled && msg.Err != nil {
			err := tuierrors.TransientError("remote connection", msg.Err.Error(), 3)
			return actions.LogError(m, err, "Remote connection failed")
		}
		return cmd

	case commands.HostConfirmMsg:
		actions.SetHostConfirmReq(m, msg)
		return nil
	}
	return nil
}
