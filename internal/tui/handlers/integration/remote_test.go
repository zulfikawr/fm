package integration_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRemote_Handlers(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("HandleRemoteConnectMsg", func(t *testing.T) {
		msg := messages.RemoteConnectMsg{FS: fs, Path: "/home"}
		integration.HandleRemote(m, msg)
		if m.FS != fs {
			t.Error("expected FS to be updated")
		}
	})

	t.Run("HandleRemoteGoto", func(t *testing.T) {
		integration.HandleRemoteGoto(m, "user@host")
	})

	t.Run("HandleAuthFinalize", func(t *testing.T) {
		integration.HandleAuthFinalize(m, "password")
	})

	t.Run("HandleHostConfirm", func(t *testing.T) {
		integration.HandleRemote(m, messages.HostConfirmMsg{})
		if !m.UI.HostConfirm {
			t.Error("expected HostConfirm UI state")
		}
	})

	t.Run("HandleHostConfirmKeys", func(t *testing.T) {
		m.UI.HostConfirm = true
		m.Remote.HostConfirmReq = &ssh.HostConfirmRequest{
			Resolve: make(chan bool, 1),
		}
		integration.HandleRemote(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
		if m.UI.HostConfirm {
			t.Error("expected HostConfirm to be false after 'y'")
		}
	})
}
