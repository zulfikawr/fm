package tui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearch(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-search-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(tmpDir)

	t.Run("Searching State", func(t *testing.T) {
		m.searching = true
		m.searchInput.Focus()
		m.searchInput.SetValue("test")

		// Test typing in searching
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		m = newModel.(*Model)
		if m.searchInput.Value() != "testa" {
			t.Errorf("Expected search value testa, got %s", m.searchInput.Value())
		}

		// Test enter in searching
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)
		if m.searching {
			t.Error("Expected searching to be false after enter")
		}
	})
}
