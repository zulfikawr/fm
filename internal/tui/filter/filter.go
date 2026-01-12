package filter

import (
	"fm/internal/files"
	"fm/internal/tui/state"
	"strings"
)

// Apply filters the items based on the search query in the model.
func Apply(m *state.Model) {
	query := strings.ToLower(m.Inputs.ActiveInput.Value())
	if !m.UI.InputActive || m.Inputs.Mode != state.InputSearch || query == "" {
		m.Navigation.FilteredItems = make([]files.Item, len(m.Navigation.Items))
		copy(m.Navigation.FilteredItems, m.Navigation.Items)
		return
	}

	m.Navigation.FilteredItems = nil
	for _, item := range m.Navigation.Items {
		// Always show the "up" directory
		if item.IsUp {
			m.Navigation.FilteredItems = append(m.Navigation.FilteredItems, item)
			continue
		}

		if strings.Contains(strings.ToLower(item.Name), query) {
			m.Navigation.FilteredItems = append(m.Navigation.FilteredItems, item)
		}
	}
}
