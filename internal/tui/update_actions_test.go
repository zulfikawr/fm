package tui

import (
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActions(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-actions-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(&files.LocalFS{}, tmpDir)

	m.items = []files.Item{
		{Name: "↑ ..", IsUp: true, IsDir: true, CanRead: true, CanWrite: true},
		{Name: "f1", Path: filepath.Join(tmpDir, "f1"), IsDir: false, CanRead: true, CanWrite: true},
	}
	m.applyFilter()

	t.Run("Toggle Selection", func(t *testing.T) {
		m.cursor = 1 // On f1
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if !m.items[1].Selected {
			t.Error("Expected item to be selected")
		}
		if !m.selectMode {
			t.Error("Expected selectMode to be true")
		}
	})

	t.Run("Copy Action", func(t *testing.T) {
		m.cursor = 1
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if len(m.clipboard) != 1 {
			t.Error("Expected 1 item in clipboard")
		}
	})

	t.Run("Cut Action", func(t *testing.T) {
		m.cursor = 1
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		if len(m.clipboard) != 1 {
			t.Error("Expected 1 item in clipboard")
		}
		if !m.clipboardCut {
			t.Error("Expected clipboardCut to be true")
		}
		if m.actionType != "cut" {
			t.Errorf("Expected actionType 'cut', got %s", m.actionType)
		}

		// Verify Copy resets it
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		if m.clipboardCut {
			t.Error("Expected clipboardCut to be false after copy")
		}
	})

	t.Run("Sort Action", func(t *testing.T) {
		initialSort := m.sortMode
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		if m.sortMode == initialSort {
			t.Error("Expected sort mode to change")
		}
	})

	t.Run("Search Toggle", func(t *testing.T) {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		if !m.searching {
			t.Error("Expected searching to be true")
		}
		m.searching = false
	})

	t.Run("Settings Toggle", func(t *testing.T) {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
		if !m.settingsOpen {
			t.Error("Expected settings to be open")
		}
		m.settingsOpen = false
	})

	t.Run("Clear Selection (Esc)", func(t *testing.T) {
		m.items[1].Selected = true
		m.selectMode = true
		m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if m.selectMode {
			t.Error("Expected selectMode to be false after Esc")
		}
		if m.items[1].Selected {
			t.Error("Expected selection to be cleared after Esc")
		}
	})

	t.Run("Paste Action", func(t *testing.T) {
		m.clipboard = []string{"somefile"}
		m.cfg.ConfirmOperations = false
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		if !m.loading {
			t.Error("Expected loading to be true after paste action")
		}
	})

	t.Run("Paste Action with Confirmation", func(t *testing.T) {
		m.clipboard = []string{"somefile"}
		m.cfg.ConfirmOperations = true
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		if !m.confirming {
			t.Error("Expected confirming to be true when ConfirmOperations is enabled")
		}
		if m.actionType != "paste" {
			t.Errorf("Expected actionType 'paste', got %s", m.actionType)
		}
		m.confirming = false
	})

	t.Run("Rename Action", func(t *testing.T) {
		m.cursor = 1
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		if !m.renaming {
			t.Error("Expected renaming to be true after 'r'")
		}
		m.renaming = false
	})

	t.Run("Delete Action", func(t *testing.T) {
		m.cursor = 1
		m.cfg.ConfirmOperations = false
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		if !m.loading {
			t.Error("Expected loading to be true after 'd'")
		}
	})
}
