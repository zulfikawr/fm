package handlers

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
)

func TestFileOps_Clipboard(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	m.Navigation.Items = []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	time.Sleep(10 * time.Millisecond)
	if len(m.Operations.Clipboard.Paths) != 1 {
		t.Errorf("expected 1 item in clipboard, got %d", len(m.Operations.Clipboard.Paths))
	}

	m.Operations.Clipboard.Clear()
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	time.Sleep(10 * time.Millisecond)
	if !m.Operations.Clipboard.IsCut {
		t.Error("expected clipboard to be in Cut mode")
	}

	_ = tm.Quit()
}

func TestFileOps_DeleteConfirm(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Config.ConfirmOperations = true

	m.Navigation.Items = []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	time.Sleep(10 * time.Millisecond)
	if !m.UI.Confirming {
		t.Error("expected UI to be in confirming mode")
	}

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	time.Sleep(10 * time.Millisecond)
	if m.UI.Confirming {
		t.Error("expected UI to stop confirming after 'n'")
	}

	_ = tm.Quit()
}

func TestFileOps_Rename(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	m.Navigation.Items = []core.Item{
		{Name: "old.txt", Path: "/test/old.txt"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Press 'r' to start rename
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	time.Sleep(10 * time.Millisecond)
	if m.Inputs.Mode != tuictx.InputRename {
		t.Errorf("expected InputRename mode, got %v", m.Inputs.Mode)
	}
	if m.Inputs.ActiveInput.Value() != "old.txt" {
		t.Errorf("expected input to be old.txt, got %q", m.Inputs.ActiveInput.Value())
	}

	_ = tm.Quit()
}
