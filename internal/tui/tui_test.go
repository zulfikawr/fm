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
		{Name: "↑ ..", IsDir: true, IsUp: true},
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
		m = newModel.(*Model)
		if m.cursor != 1 {
			t.Errorf("Expected cursor at 1, got %d", m.cursor)
		}

		// Press up
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(*Model)
		if m.cursor != 0 {
			t.Errorf("Expected cursor at 0, got %d", m.cursor)
		}
	})

	t.Run("Toggle Selection", func(t *testing.T) {
		m.cursor = 1 // On dir1
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)

		if !m.filteredItems[1].Selected {
			t.Error("Expected item at cursor to be selected")
		}
	})

	t.Run("Search Toggle", func(t *testing.T) {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
		m = newModel.(*Model)
		if !m.searching {
			t.Error("Expected searching to be true")
		}
	})

	t.Run("Filtering", func(t *testing.T) {
		// Mock search input
		m.searching = true
		m.searchInput.SetValue("file")
		m.applyFilter()

		// Items: ↑ .., dir1, file1. Filter "file" should show ↑ .. and file1
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
		m = newModel.(*Model)
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
		m = newModel.(*Model)
		if m.width != 100 || m.height != 50 {
			t.Errorf("Expected 100x50, got %dx%d", m.width, m.height)
		}
	})

	t.Run("Cycle Sort", func(t *testing.T) {
		initialSort := m.sortMode
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
		m = newModel.(*Model)
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
			{Name: "↑ ..", IsDir: true, IsUp: true},
			{Name: "dir1", IsDir: true, Path: dirPath},
		}
		m.searchInput.SetValue("") // Ensure no search filter
		m.applyFilter()
		m.cursor = 1 // On dir1

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

	t.Run("Clipboard Copy", func(t *testing.T) {
		// Mock items and filteredItems
		m.items = []files.Item{{Name: "file1", Path: "/tmp/file1", Selected: true}}
		m.applyFilter()

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		m = newModel.(*Model)
		if len(m.clipboard) != 1 || m.clipboard[0] != "/tmp/file1" {
			t.Errorf("Expected /tmp/file1 in clipboard, got %v", m.clipboard)
		}
	})

	t.Run("Select Mode Logic", func(t *testing.T) {
		m.items = []files.Item{
			{Name: "file1", Path: "file1", Selected: false},
			{Name: "file2", Path: "file2", Selected: false},
		}
		m.selectMode = false
		m.applyFilter()
		m.width = 80 // Set width for rendering

		// Check row without selectMode
		row := m.renderRow(m.items[0], false)
		if strings.Contains(row, "[ ]") || strings.Contains(row, "[x]") || strings.HasPrefix(row, "    ") {
			t.Error("Markers and leading spaces should be hidden when selectMode is false")
		}

		// Press space on file1
		m.cursor = 0
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if !m.selectMode {
			t.Error("Expected selectMode to be true after selecting an item")
		}

		// Check row with selectMode
		row = m.renderRow(m.items[0], false)
		if !strings.Contains(row, "[x]") {
			t.Error("Expected [x] marker when item is selected and selectMode is true")
		}
		row = m.renderRow(m.items[1], false)
		if !strings.Contains(row, "[ ]") {
			t.Error("Expected [ ] marker when item is not selected but selectMode is true")
		}

		// Press space on file1 again (unselect)
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.selectMode {
			t.Error("Expected selectMode to be false after unselecting all items")
		}

		// Select again and then press ESC
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if !m.selectMode {
			t.Error("Expected selectMode to be true")
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(*Model)
		if m.selectMode {
			t.Error("Expected selectMode to be false after pressing ESC")
		}
		if m.items[0].Selected {
			t.Error("Expected selections to be cleared after ESC")
		}
	})

	t.Run("Quit", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
		if cmd == nil {
			t.Fatal("Expected quit command")
		}
	})
}

func TestSettingsLogic(t *testing.T) {
	m := NewModel("/")
	m.settingsOpen = true
	m.width = 80
	m.height = 24

	t.Run("Navigation skipping disabled settings", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.ShowDateModified = false
		m.settingsCursor = 5 // On "Enable Git Status" (oops, wait)

		// In my current order:
		// 0: Hidden
		// 1: CaseSensitive
		// 2: Confirmations
		// 3: Wrap
		// 4: ShowHeader
		// 5: EnableGit
		// 6: ShowSize
		// 7: SizeFormat (skip if 6 is off)
		// 8: ShowDate
		// 9: DateFormat (skip if 8 is off)
		// 10: Theme

		m.settingsCursor = 6 // On ShowSize (off)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		// Should skip 7 and go to 8
		if m.settingsCursor != 8 {
			t.Errorf("Expected cursor to skip 7 and go to 8, got %d", m.settingsCursor)
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		// Should skip 9 and go to 10
		if m.settingsCursor != 10 {
			t.Errorf("Expected cursor to skip 9 and go to 10, got %d", m.settingsCursor)
		}

		// Move up from 10
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(*Model)
		// Should skip 9 and go to 8
		if m.settingsCursor != 8 {
			t.Errorf("Expected cursor to skip 9 and go to 8 when moving up, got %d", m.settingsCursor)
		}
	})

	t.Run("Toggle logic for formats", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.SizeFormatIndex = 0
		m.settingsCursor = 7 // On SizeFormat

		// Toggle should NOT work if ShowSize is off
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.cfg.SizeFormatIndex != 0 {
			t.Error("SizeFormat should not toggle when ShowSize is off")
		}

		m.cfg.ShowSize = true
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.cfg.SizeFormatIndex == 0 {
			t.Error("SizeFormat should toggle when ShowSize is true")
		}
	})

	t.Run("Settings rendering", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.settingsOpen = true
		m.width = 80

		header := m.renderHeader()
		footer := m.renderFooter()
		content := m.renderSettingsList(header, footer)

		if !strings.Contains(content, "File Operations") {
			t.Error("Settings list should contain group headers")
		}

		// Size format should be present but its value might be dimmed (hard to test exact style without complex matches)
		if !strings.Contains(content, "Size Format") {
			t.Error("Settings list should contain Size Format even if disabled")
		}
	})
}
