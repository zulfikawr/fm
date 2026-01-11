package tui

import (
	"fm/internal/files"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateWindowSize(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	m.searching = false
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(*Model)
	if m.width != 100 || m.height != 50 {
		t.Errorf("Expected 100x50, got %dx%d", m.width, m.height)
	}
}

func TestUpdateQuit(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Expected quit command")
	}
}
