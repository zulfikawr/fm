package tui

import (
	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleSearching(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		}
	}

	m.searchInput, cmd = m.searchInput.Update(msg)

	// Validate query
	if err := files.ValidateSearchQuery(m.searchInput.Value()); err != nil {
		// If query is invalid, strip the last character or clear it
		val := m.searchInput.Value()
		if len(val) > 0 {
			m.searchInput.SetValue(val[:len(val)-1])
		}
	}

	m.applyFilter()
	if m.cursor >= len(m.filteredItems) {
		m.cursor = 0
		m.offset = 0
	}
	return m, cmd
}
