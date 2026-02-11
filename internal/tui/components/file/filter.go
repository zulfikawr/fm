package file

import (
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

// ApplyFilter filters the items based on the search query in the model
func ApplyFilter(m *tuictx.Model) {
	query := m.Navigation.FilterQuery
	if query == "" {
		m.Navigation.FilteredItems = make([]core.Item, len(m.Navigation.Items))
		copy(m.Navigation.FilteredItems, m.Navigation.Items)
		return
	}

	m.Navigation.FilteredItems = nil
	for i := range m.Navigation.Items {
		item := m.Navigation.Items[i]
		if item.State.IsUp {
			m.Navigation.FilteredItems = append(m.Navigation.FilteredItems, item)
			continue
		}

		if strings.Contains(item.State.SearchKey, query) {
			m.Navigation.FilteredItems = append(m.Navigation.FilteredItems, item)
		}
	}

	// Ensure cursor is within bounds after filtering
	if m.Navigation.Cursor >= len(m.Navigation.FilteredItems) {
		if len(m.Navigation.FilteredItems) > 0 {
			m.Navigation.Cursor = len(m.Navigation.FilteredItems) - 1
		} else {
			m.Navigation.Cursor = 0
		}
	}
}
