package app

import (
	"errors"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
)

func TestCommands_Basic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("SetMsg", func(t *testing.T) {
		utils.SetMsg(m, "test message")
		if m.Message.Text != "test message" {
			t.Errorf("expected test message, got %s", m.Message.Text)
		}
	})

	t.Run("LogError", func(t *testing.T) {
		err := errors.New("test error")
		utils.LogError(m, err, "test context")
		if m.Message.Error == nil {
			t.Error("expected error to be set in model")
		}
	})
}

func TestCommands_Logs(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("LogPush and LogUpdate", func(t *testing.T) {
		id := utils.LogPush(m, "Test", tuictx.LogInfo, tuictx.StatusPending, "msg", "details")
		if id == "" {
			t.Fatal("expected non-empty log ID")
		}

		utils.LogUpdate(m, id, tuictx.StatusSuccess, tuictx.LogSuccess, "updated msg", "updated details")
		entry := m.Logs.Entries[0]
		if entry.Status != tuictx.StatusSuccess {
			t.Errorf("expected status Success, got %v", entry.Status)
		}
	})
}

func TestCommands_Watcher(t *testing.T) {
	t.Run("WatchRemoteDir", func(t *testing.T) {
		cmd := utils.WatchRemoteDir()
		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})

	t.Run("WatchDir", func(t *testing.T) {
		cmd := utils.WatchDir(nil)
		if cmd != nil {
			t.Error("expected nil command for nil watcher")
		}
	})
}
