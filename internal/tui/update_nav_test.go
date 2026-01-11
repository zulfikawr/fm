package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNavigation(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-nav-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(&files.LocalFS{}, tmpDir)

	m.items = []files.Item{
		{Name: "↑ ..", IsDir: true, IsUp: true, CanRead: true},
		{Name: "dir1", IsDir: true, Path: filepath.Join(tmpDir, "dir1"), CanRead: true},
		{Name: "file1", IsDir: false, Path: filepath.Join(tmpDir, "file1"), CanRead: true},
	}
	m.applyFilter()

	t.Run("Move Cursor", func(t *testing.T) {
		m.cursor = 0
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		if m.cursor != 1 {
			t.Errorf("Expected cursor at 1, got %d", m.cursor)
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(*Model)
		if m.cursor != 0 {
			t.Errorf("Expected cursor at 0, got %d", m.cursor)
		}
	})

	t.Run("Wrap Navigation", func(t *testing.T) {
		m.cfg.WrapNavigation = true
		m.cursor = 0
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(*Model)
		if m.cursor != len(m.filteredItems)-1 {
			t.Errorf("Expected wrapped cursor at end, got %d", m.cursor)
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		if m.cursor != 0 {
			t.Errorf("Expected wrapped cursor at 0, got %d", m.cursor)
		}
	})

	t.Run("Navigation Enter Dir", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "dir1")
		os.Mkdir(dirPath, 0755)

		m.path = tmpDir
		m.cursor = 1 // dir1
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)
		if !strings.HasSuffix(m.path, "dir1") {
			t.Errorf("Expected path to end with dir1, got %s", m.path)
		}
	})

	t.Run("Navigation Back", func(t *testing.T) {
		oldPath := m.path
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(*Model)
		if m.path == oldPath {
			t.Error("Expected path to change after backspace")
		}
	})

	t.Run("Horizontal Navigation", func(t *testing.T) {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	})
}

func TestFileOpening(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-open-test")
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("hello"), 0644)

	m := NewModel(&files.LocalFS{}, tmpDir)
	m.items = []files.Item{{Name: "test.txt", Path: filePath, IsDir: false, CanRead: true}}
	m.applyFilter()
	m.cursor = 0

	t.Run("Open with editor", func(t *testing.T) {
		m.cfg.EditorIndex = 0 // vim (terminal)
		newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if cmd == nil {
			t.Error("Expected non-nil cmd for terminal editor")
		}
		m = newModel.(*Model)
	})

	t.Run("Open with non-terminal editor", func(t *testing.T) {
		m.cfg.EditorIndex = 4 // code (non-terminal)
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		// We can't easily verify the execution in test, but we ensure it doesn't panic
	})
}
