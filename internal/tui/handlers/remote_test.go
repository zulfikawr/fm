package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRemote_ConnectMsg(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Success", func(t *testing.T) {
		remoteFS := testutil.NewMockFileSystem()
		msg := RemoteConnectMsg{
			FS:   remoteFS,
			Path: "/home/user",
		}
		HandleRemote(m, msg)
		if m.FS != remoteFS {
			t.Error("expected FS to be updated to remoteFS")
		}
	})

	t.Run("Auth Required", func(t *testing.T) {
		msg := RemoteConnectMsg{
			Err: errors.New("ssh: handshake failed: ssh: unable to authenticate"),
		}
		HandleRemote(m, msg)
		if !m.UI.RemoteAuth {
			t.Error("expected UI.RemoteAuth to be true")
		}
		if m.Inputs.Mode != tuictx.InputAuth {
			t.Errorf("expected InputAuth mode, got %v", m.Inputs.Mode)
		}
	})

	t.Run("Host Down Error", func(t *testing.T) {
		m.UI.RemoteAuth = false
		msg := RemoteConnectMsg{
			Err: errors.New("dial tcp 1.2.3.4:22: connect: connection refused"),
		}
		HandleRemote(m, msg)
		if m.UI.RemoteAuth {
			t.Error("expected UI.RemoteAuth to be false for connection refused")
		}
		if m.Message.Text == "" {
			t.Error("expected error message to be set")
		}
	})
}

func TestRemote_HostConfirm(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Handle HostConfirmMsg", func(t *testing.T) {
		resolve := make(chan bool, 1)
		req := &ssh.HostConfirmRequest{
			Hostname: "example.com",
			Resolve:  resolve,
		}
		HandleRemote(m, HostConfirmMsg{Request: req})

		if !m.UI.HostConfirm {
			t.Error("expected UI.HostConfirm to be true")
		}
		if m.Remote.HostConfirmReq != req {
			t.Error("expected HostConfirmReq to be set")
		}
	})

	t.Run("Accept Host", func(t *testing.T) {
		resolve := make(chan bool, 1)
		m.Remote.HostConfirmReq = &ssh.HostConfirmRequest{
			Hostname: "example.com",
			Resolve:  resolve,
		}
		m.UI.HostConfirm = true

		HandleRemote(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

		if m.UI.HostConfirm {
			t.Error("expected UI.HostConfirm to be false after 'y'")
		}

		select {
		case res := <-resolve:
			if !res {
				t.Error("expected resolve to be true")
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for resolve")
		}
	})
}

func TestRemote_ListenForHostConfirmation(t *testing.T) {
	askChan := make(chan *ssh.HostConfirmRequest, 1)
	req := &ssh.HostConfirmRequest{Hostname: "test-host"}
	askChan <- req

	cmd := listenForHostConfirmation(askChan)
	msg := cmd()

	confirmMsg, ok := msg.(HostConfirmMsg)
	if !ok {
		t.Fatalf("expected HostConfirmMsg, got %T", msg)
	}
	if confirmMsg.Request != req {
		t.Errorf("expected request %v, got %v", req, confirmMsg.Request)
	}

	close(askChan)
	msg = listenForHostConfirmation(askChan)()
	if msg != nil {
		t.Errorf("expected nil msg from closed channel, got %v", msg)
	}
}
