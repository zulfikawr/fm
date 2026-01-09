package tui

import (
	"strings"

	"filemanager/internal/files"
)

func (m *Model) applyFilter() {
	query := m.searchInput.Value()
	if !m.cfg.CaseSensitive {
		query = strings.ToLower(query)
	}

	if query == "" {
		m.filteredItems = m.items
		return
	}

	var filtered []files.Item
	if len(m.items) > 0 && m.items[0].IsUp {
		filtered = append(filtered, m.items[0])
	}

	for _, item := range m.items {
		if item.IsUp {
			continue
		}

		name := item.Name
		if !m.cfg.CaseSensitive {
			name = strings.ToLower(name)
		}

		if strings.Contains(name, query) {
			filtered = append(filtered, item)
		}
	}
	m.filteredItems = filtered
}
