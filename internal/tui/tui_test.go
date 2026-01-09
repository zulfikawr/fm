package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelNavigation(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-tui-test")
	defer os.RemoveAll(tmpDir)

	m := NewModel(tmpDir)

	// Mock some items
	m.items = []files.Item{
		{Name: "..", IsDir: true, IsUp: true},
		{Name: "dir1", IsDir: true, Path: "dir1"},
		{Name: "file1", IsDir: false, Path: "file1"},
	}
	m.applyFilter()

	t.Run("Move Cursor", func(t *testing.T) {
		// Start at 0
		if m.cursor != 0 {
			t.Errorf("Expected cursor at 0, got %d", m.cursor)
		}

		// Press down
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(Model)
		if m.cursor != 1 {
			t.Errorf("Expected cursor at 1, got %d", m.cursor)
		}

		// Press up
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(Model)
		if m.cursor != 0 {
			t.Errorf("Expected cursor at 0, got %d", m.cursor)
		}
	})

	t.Run("Toggle Selection", func(t *testing.T) {
		m.cursor = 1 // On dir1
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(Model)

		if !m.filteredItems[1].Selected {
			t.Error("Expected item at cursor to be selected")
		}
	})

	t.Run("Search Toggle", func(t *testing.T) {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
		m = newModel.(Model)
		if !m.searching {
			t.Error("Expected searching to be true")
		}
	})

	t.Run("Filtering", func(t *testing.T) {
		// Mock search input
		m.searching = true
		m.searchInput.SetValue("file")
		m.applyFilter()

		// Items: .., dir1, file1. Filter "file" should show .. and file1
		if len(m.filteredItems) != 2 {
			t.Errorf("Expected 2 filtered items, got %d", len(m.filteredItems))
		}
		if m.filteredItems[1].Name != "file1" {
			t.Errorf("Expected filtered item to be file1, got %s", m.filteredItems[1].Name)
		}
	})

	t.Run("LoadedItemsMsg", func(t *testing.T) {
		items := []files.Item{{Name: "test.txt", IsDir: false}}
		msg := LoadedItemsMsg{
			Path:  m.path, // Must match model's path
			Items: items,
		}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)
		if len(m.items) != 1 || m.items[0].Name != "test.txt" {
			t.Errorf("Expected 1 item 'test.txt', got %d items", len(m.items))
		}
		if m.loading {
			t.Error("Expected loading to be false after LoadedItemsMsg")
		}
	})

	t.Run("WindowSize", func(t *testing.T) {
		m.searching = false // Ensure we are not in searching mode
		newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
		m = newModel.(Model)
		if m.width != 100 || m.height != 50 {
			t.Errorf("Expected 100x50, got %dx%d", m.width, m.height)
		}
	})

	t.Run("Cycle Sort", func(t *testing.T) {
		initialSort := m.sortMode
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
		m = newModel.(Model)
		if m.sortMode == initialSort {
			t.Error("Expected sort mode to change")
		}
	})

	t.Run("Navigation Enter Dir", func(t *testing.T) {
		// Create actual dir on disk
		dirPath := filepath.Join(tmpDir, "dir1")
		os.Mkdir(dirPath, 0755)

		m.path = tmpDir
		m.items = []files.Item{
			{Name: "..", IsDir: true, IsUp: true},
			{Name: "dir1", IsDir: true, Path: dirPath},
		}
		m.searchInput.SetValue("") // Ensure no search filter
		m.applyFilter()
		m.cursor = 1 // On dir1

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(Model)
		if !strings.HasSuffix(m.path, "dir1") {
			t.Errorf("Expected path to end with dir1, got %s", m.path)
		}
	})

	t.Run("Navigation Back", func(t *testing.T) {
		oldPath := m.path
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		m = newModel.(Model)
		if m.path == oldPath {
			t.Error("Expected path to change after backspace")
		}
	})

	t.Run("Clipboard Copy", func(t *testing.T) {
		// Mock items and filteredItems
		m.items = []files.Item{{Name: "file1", Path: "/tmp/file1", Selected: true}}
		m.applyFilter()

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = newModel.(Model)
		if len(m.clipboard) != 1 || m.clipboard[0] != "/tmp/file1" {
			t.Errorf("Expected /tmp/file1 in clipboard, got %v", m.clipboard)
		}
	})

	t.Run("Quit", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
		if cmd == nil {
			t.Fatal("Expected quit command")
		}
		// In bubbletea, tea.Quit returns a special command.
		// We can't easily check if it's tea.Quit because it's a function pointer.
	})
}
