package handlers

import (
	"errors"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	tuierrors "github.com/zulfikawr/fm/internal/tui/errors"
)

func TestCommands_Helpers(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("SetMsg", func(t *testing.T) {
		SetMsg(m, "test message")
		if m.Message.Text != "test message" {
			t.Errorf("expected 'test message', got %q", m.Message.Text)
		}
	})

	t.Run("LogError", func(t *testing.T) {
		err := errors.New("test error")
		LogError(m, err, "Context")
		if m.Message.Error == nil {
			t.Fatal("expected message error to be set")
		}

		tuiErr, ok := m.Message.Error.(*tuierrors.Error)
		if !ok {
			t.Fatal("expected error to be *tuierrors.Error")
		}
		if tuiErr.Operation != "Context" {
			t.Errorf("expected operation Context, got %s", tuiErr.Operation)
		}
	})

	t.Run("LogPush/Update", func(t *testing.T) {
		id := LogPush(m, "Test", tuictx.LogInfo, tuictx.StatusRunning, "Start", "")
		if len(m.Logs.Entries) != 1 {
			t.Fatal("expected 1 log entry")
		}

		LogUpdate(m, id, tuictx.StatusSuccess, tuictx.LogSuccess, "Finished", "details")
		entry := m.Logs.Entries[0]
		if entry.Status != tuictx.StatusSuccess {
			t.Errorf("expected StatusSuccess, got %v", entry.Status)
		}
	})
}

func TestCommands_Watcher(t *testing.T) {
	t.Run("WatchRemoteDir", func(t *testing.T) {
		cmd := WatchRemoteDir()
		msg := cmd()
		if _, ok := msg.(RemotePollMsg); !ok {
			t.Errorf("expected RemotePollMsg, got %T", msg)
		}
	})

	t.Run("WatchDir Nil", func(t *testing.T) {
		cmd := WatchDir(nil)
		if cmd != nil {
			t.Error("expected nil cmd for nil watcher")
		}
	})
}
