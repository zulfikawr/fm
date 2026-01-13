package update

import (
	"errors"
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
)

func TestHandleRemoteMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test HostConfirmMsg
	hostMsg := commands.HostConfirmMsg{Request: nil}
	_ = HandleRemoteMsg(m, hostMsg)
	if !m.UI.HostConfirm {
		t.Error("Expected HostConfirm to be true")
	}

	// Test RemoteConnectMsg Success
	m.UI.Loading = true
	connectMsg := commands.RemoteConnectMsg{
		FS:   m.FS,
		Path: "/remote",
	}
	// Call FinalizeRemoteConnect directly as it's definitely where state changes
	_, _ = actions.FinalizeRemoteConnect(m, connectMsg)
	if m.Navigation.Path != "/remote" {
		t.Errorf("Expected path /remote, got %s", m.Navigation.Path)
	}
	// Reload() sets Loading to true, so it should be true now
	if !m.UI.Loading {
		t.Error("Expected loading to be true after Reload")
	}

	// Test RemoteConnectMsg Auth Error
	m.UI.RemoteAuth = false
	connectErr := commands.RemoteConnectMsg{Err: errors.New("auth failed")}
	_, _ = actions.FinalizeRemoteConnect(m, connectErr)
	if !m.UI.RemoteAuth {
		t.Error("Expected RemoteAuth to be true after auth failure")
	}
}
