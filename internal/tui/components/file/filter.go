package file

import (
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
)

// ApplyFilter filters the items based on the search query in the model
func ApplyFilter(m *tui_context.Model) {
	query := m.Navigation.FilterQuery
	if query == "" {
		m.Navigation.FilteredItems = make([]core.Item, len(m.Navigation.Items))
		copy(m.Navigation.FilteredItems, m.Navigation.Items)
		return
	}

	m.Navigation.FilteredItems = nil
	for _, item := range m.Navigation.Items {
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
