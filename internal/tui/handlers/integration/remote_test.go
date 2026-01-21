package integration

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestRemote_Connect(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	msg := messages.RemoteConnectMsg{
		FS:   fs,
		Path: "/home/user",
	}

	HandleRemote(m, msg)
	// Success case returns NavigateMsg
}

func TestRemote_HostConfirm(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	msg := messages.HostConfirmMsg{
		Request: nil,
	}

	HandleRemote(m, msg)
	if !m.UI.HostConfirm {
		t.Error("expected HostConfirm state")
	}
}
