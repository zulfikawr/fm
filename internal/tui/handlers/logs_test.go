package handlers

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
)

func TestLogs_Handler(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.LogOpen = true
	m.Logs.Entries = []tuictx.LogEntry{
		{ID: "1", Message: "Log 1"},
		{ID: "2", Message: "Log 2"},
	}
	m.Display.ViewportHeight = 20

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Move down
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(10 * time.Millisecond)
	if m.Logs.Cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.Logs.Cursor)
	}

	// Toggle via alt+l
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l"), Alt: true})
	time.Sleep(10 * time.Millisecond)
	if m.UI.LogOpen {
		t.Error("expected logs closed")
	}

	_ = tm.Quit()
}
