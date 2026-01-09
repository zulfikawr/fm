package tui

import (
	"os"
	"path/filepath"
	"testing"

	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActions(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-actions-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(tmpDir)

	m.items = []files.Item{
		{Name: "↑ ..", IsUp: true, IsDir: true},
		{Name: "f1", Path: filepath.Join(tmpDir, "f1"), IsDir: false},
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
}
