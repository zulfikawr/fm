package tui

import (
	"testing"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabStateSync(t *testing.T) {
	fs := &files.LocalFS{}
	m := NewModel(fs, "/tmp")

	// Simulate initial load for tab 1
	items1 := make([]files.Item, 10)
	for i := range items1 {
		items1[i] = files.Item{Name: "file", Path: "/tmp/file"}
	}
	m.Update(LoadedItemsMsg{Generation: m.pathGeneration, Path: "/tmp", Items: items1})

	// Setup state in Tab 1
	m.cursor = 5
	m.offset = 2
	m.saveTabState()

	// Create second tab (alt+t)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true}
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	if len(m.tabs) != 2 {
		t.Fatalf("Expected 2 tabs, got %d", len(m.tabs))
	}

	// Tab 2 should start at /tmp but with cursor 0
	if m.cursor != 0 {
		t.Errorf("Expected cursor 0 in new tab, got %d", m.cursor)
	}

	// Navigate Tab 2 to /home
	m.path = "/home"
	m.pathGeneration++
	m.cursor = 3
	m.offset = 1
	m.saveTabState()

	// Switch back to Tab 1 (alt+1)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1"), Alt: true}
	newModel, _ = m.Update(msg)
	m = newModel.(*Model)

	if m.activeTab != 0 {
		t.Errorf("Expected active tab 0, got %d", m.activeTab)
	}
	if m.path != "/tmp" {
		t.Errorf("Expected path /tmp, got %s", m.path)
	}

	// Immediately after switch (before reload), cursor should be 5
	if m.cursor != 5 {
		t.Errorf("Expected cursor 5 in tab 1 immediately after switch, got %d", m.cursor)
	}

	// Simulate reload for Tab 1
	items1New := make([]files.Item, 20)
	for i := range items1New {
		items1New[i] = files.Item{Name: "file", Path: "/tmp/file"}
	}
	loadMsg := LoadedItemsMsg{
		Generation: m.pathGeneration,
		Path:       "/tmp",
		Items:      items1New,
	}
	newModel, _ = m.Update(loadMsg)
	m = newModel.(*Model)

	if m.cursor != 5 {
		t.Errorf("Expected cursor 5 in tab 1 after reload, got %d", m.cursor)
	}

	// Switch back to Tab 2 (alt+2)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2"), Alt: true}
	newModel, _ = m.Update(msg)
	m = newModel.(*Model)

	if m.activeTab != 1 {
		t.Errorf("Expected active tab 1, got %d", m.activeTab)
	}
	if m.cursor != 3 {
		t.Errorf("Expected cursor 3 in tab 2, got %d", m.cursor)
	}
}
