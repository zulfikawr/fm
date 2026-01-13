package update

import (
	"context"
	"os"
	"testing"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleNavigation(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.ViewportHeight = 10
	m.Navigation.FilteredItems = []core.Item{
		{Name: "f1", Path: "/f1"},
		{Name: "f2", Path: "/f2"},
		{Name: "f3", Path: "/f3"},
	}
	m.Navigation.Cursor = 1

	// Test down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	_ = HandleNavigation(msg, m)
	if m.Navigation.Cursor != 2 {
		t.Errorf("Expected cursor 2, got %d", m.Navigation.Cursor)
	}

	// Test up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	_ = HandleNavigation(msg, m)
	if m.Navigation.Cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", m.Navigation.Cursor)
	}

	// Test 'enter' on directory
	m.Navigation.FilteredItems[1].IsDir = true
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	m.FS.(*testutil.MockFileSystem).StatFunc = func(ctx context.Context, p string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}
	cmds := HandleNavigation(msg, m)
	if len(cmds) == 0 {
		t.Error("Expected navigation cmd for enter on dir")
	}

	// Test backspace
	msg = tea.KeyMsg{Type: tea.KeyBackspace}
	cmds = HandleNavigation(msg, m)
	if len(cmds) == 0 {
		t.Error("Expected navigation cmd for backspace")
	}

	// Test wrap up
	m.Config.WrapNavigation = true
	m.Navigation.FilteredItems = []core.Item{{Name: "f1"}, {Name: "f2"}, {Name: "f3"}}
	m.Navigation.Cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	_ = HandleNavigation(msg, m)
	if m.Navigation.Cursor != 2 {
		t.Errorf("Expected cursor wrapped to 2, got %d", m.Navigation.Cursor)
	}

	// Test remote enter on file
	m.Navigation.FilteredItems = []core.Item{{Name: "f1", IsDir: false}}
	m.FS.(*testutil.MockFileSystem).IsLocalFunc = func() bool { return false }
	m.Navigation.Cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	cmds = HandleNavigation(msg, m)
	if len(cmds) == 0 {
		t.Error("Expected message cmd for remote enter")
	}
}

func TestScroll(t *testing.T) {
	// scroll(cursor, offset, viewportHeight)
	if scroll(5, 0, 10) != 0 {
		t.Error("Should not scroll if within viewport")
	}
	if scroll(15, 0, 10) != 6 {
		t.Errorf("Expected 6, got %d", scroll(15, 0, 10))
	}
	if scroll(2, 5, 10) != 2 {
		t.Errorf("Expected 2, got %d", scroll(2, 5, 10))
	}
}
