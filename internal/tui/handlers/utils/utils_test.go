package utils

import (
	"errors"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
)

func TestUtils(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tui_context.NewModel(fs, "/")

	t.Run("SetMsg and SetErrMsg", func(t *testing.T) {
		SetMsg(m, "test msg")
		if m.Message.Text != "test msg" {
			t.Errorf("SetMsg failed, got %q", m.Message.Text)
		}
		SetErrMsg(m, "error msg")
		if m.Message.Text != "error msg" {
			t.Errorf("SetErrMsg failed, got %q", m.Message.Text)
		}
	})

	t.Run("Log Operations", func(t *testing.T) {
		id := LogPush(m, "test", tui_context.LogInfo, tui_context.StatusSuccess, "msg", "details")
		if id == "" {
			t.Error("LogPush failed")
		}
		LogUpdate(m, id, tui_context.StatusError, tui_context.LogError, "new msg", "new details")
	})

	t.Run("LogError", func(t *testing.T) {
		LogError(m, errors.New("system error"), "context")
		if m.Message.Error == nil {
			t.Error("LogError failed")
		}
		LogError(m, nil, "") // should do nothing
	})

	t.Run("WatchDir nil", func(t *testing.T) {
		cmd := WatchDir(nil)
		if cmd != nil {
			t.Error("WatchDir(nil) should return nil")
		}
	})

	t.Run("WatchRemoteDir", func(t *testing.T) {
		cmd := WatchRemoteDir()
		if cmd == nil {
			t.Error("WatchRemoteDir should return cmd")
		}
	})

	t.Run("RestartWatcherAction", func(t *testing.T) {
		cmd := RestartWatcherAction(m)
		if cmd == nil {
			t.Error("RestartWatcherAction should return cmd")
		}
	})
}
