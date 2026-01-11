package tui

import (
	"sort"
	"strings"

	"fm/internal/files"
)

func (m *Model) applyFilter() {
	query := m.searchInput.Value()
	if !m.cfg.CaseSensitive {
		query = strings.ToLower(query)
	}

	var filtered []files.Item
	// Filter items
	if query == "" {
		filtered = make([]files.Item, len(m.items))
		copy(filtered, m.items)
	} else {
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
	}

	// Sort filtered items
	if len(filtered) > 1 {
		startIndex := 0
		if filtered[0].IsUp {
			startIndex = 1
		}

		toSort := filtered[startIndex:]
		sort.SliceStable(toSort, func(i, j int) bool {
			switch m.sortMode {
			case files.SortName:
				return strings.ToLower(toSort[i].Name) < strings.ToLower(toSort[j].Name)
			case files.SortNameDesc:
				return strings.ToLower(toSort[i].Name) > strings.ToLower(toSort[j].Name)
			case files.SortNewest:
				return toSort[i].MTime.After(toSort[j].MTime)
			case files.SortOldest:
				return toSort[i].MTime.Before(toSort[j].MTime)
			case files.SortSizeDesc:
				// Directories might still be counting (-1)
				return toSort[i].Size > toSort[j].Size
			case files.SortSizeAsc:
				return toSort[i].Size < toSort[j].Size
			default: // SortDefault
				if toSort[i].IsDir != toSort[j].IsDir {
					return toSort[i].IsDir
				}
				return strings.ToLower(toSort[i].Name) < strings.ToLower(toSort[j].Name)
			}
		})
	}

	m.filteredItems = filtered
}
