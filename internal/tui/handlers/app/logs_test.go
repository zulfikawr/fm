package app_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogs_Keys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.ActiveView = tuictx.ViewLogs
	m.Logs.Entries = []tuictx.LogEntry{
		{Message: "Log 1"},
		{Message: "Log 2"},
	}

	app.HandleLogs(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Logs.Cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.Logs.Cursor)
	}
}
