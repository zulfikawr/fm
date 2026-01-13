package actions

import (
	"errors"
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/tui/commands"
)

func TestFinalizeRemoteConnect(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Success
	msg := commands.RemoteConnectMsg{Path: "/remote", FS: m.FS}
	_, handled := FinalizeRemoteConnect(m, msg)
	if !handled {
		t.Error("Expected handled to be true")
	}
	if m.Navigation.Path != "/remote" {
		t.Errorf("Expected path /remote, got %s", m.Navigation.Path)
	}

	// Auth Error
	m.UI.RemoteAuth = false
	msgErr := commands.RemoteConnectMsg{Err: errors.New("auth failed")}
	_, handled = FinalizeRemoteConnect(m, msgErr)
	if !handled {
		t.Error("Expected handled to be true for auth prompt trigger")
	}
	if !m.UI.RemoteAuth {
		t.Error("Expected RemoteAuth to be true")
	}
}

func TestSetHostConfirmReq(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	msg := commands.HostConfirmMsg{Request: nil}

	SetHostConfirmReq(m, msg)
	if !m.UI.HostConfirm {
		t.Error("Expected HostConfirm to be true")
	}
}
