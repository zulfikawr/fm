package handlers

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
)

func TestClipboard_Handler(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.ClipboardOpen = true
	m.Operations.Clipboard.Paths = []string{"/test/f1.txt", "/test/f2.txt"}
	m.Display.ViewportHeight = 20

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	t.Run("Navigation", func(t *testing.T) {
		tm.Send(tea.KeyMsg{Type: tea.KeyDown})
		time.Sleep(10 * time.Millisecond)
		if m.Operations.Clipboard.Cursor != 1 {
			t.Errorf("expected clipboard cursor at 1, got %d", m.Operations.Clipboard.Cursor)
		}

		tm.Send(tea.KeyMsg{Type: tea.KeyUp})
		time.Sleep(10 * time.Millisecond)
		if m.Operations.Clipboard.Cursor != 0 {
			t.Errorf("expected clipboard cursor at 0, got %d", m.Operations.Clipboard.Cursor)
		}
	})

	t.Run("Remove item ('d')", func(t *testing.T) {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
		time.Sleep(10 * time.Millisecond)
		if len(m.Operations.Clipboard.Paths) != 1 {
			t.Errorf("expected 1 item in clipboard, got %d", len(m.Operations.Clipboard.Paths))
		}
	})

	t.Run("Close (Esc)", func(t *testing.T) {
		tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
		time.Sleep(10 * time.Millisecond)
		if m.UI.ClipboardOpen {
			t.Error("expected clipboard to be closed after esc")
		}
	})

	_ = tm.Quit()
}

func TestScrollClipboard(t *testing.T) {
	t.Run("Scroll Down", func(t *testing.T) {
		offset := scrollClipboard(5, 0, 3)
		if offset != 3 {
			t.Errorf("expected offset 3, got %d", offset)
		}
	})

	t.Run("Scroll Up", func(t *testing.T) {
		offset := scrollClipboard(1, 3, 3)
		if offset != 1 {
			t.Errorf("expected offset 1, got %d", offset)
		}
	})
}
