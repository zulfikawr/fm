package app_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestHandleUpdateMessages(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("UpdateAvailableMsg", func(t *testing.T) {
		msg := messages.UpdateAvailableMsg{Version: "v1.0.0"}
		app.HandleUpdateMessages(m, msg)
		if !m.UI.UpdateAvailable {
			t.Error("expected UpdateAvailable to be true")
		}
	})

	t.Run("UpdateProgressMsg", func(t *testing.T) {
		msg := messages.UpdateProgressMsg(0.5)
		app.HandleUpdateMessages(m, msg)
		if m.Operations.Progress.Percent != 0.5 {
			t.Errorf("expected 0.5 progress, got %f", m.Operations.Progress.Percent)
		}
	})

	t.Run("UpdateFinishedMsg", func(t *testing.T) {
		msg := messages.UpdateFinishedMsg{Err: nil}
		app.HandleUpdateMessages(m, msg)
		if m.UI.UpdateAvailable {
			t.Error("expected UpdateAvailable to be false after finish")
		}
	})
}
