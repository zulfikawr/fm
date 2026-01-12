package update

import (
	"fm/internal/files/ops"
	"fm/internal/tui/actions"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleSearching handles search input events
func HandleSearching(msg tea.Msg, m *state.Model) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			actions.ClosePrompt(m)
			return nil
		}
	}

	m.Inputs.ActiveInput, cmd = m.Inputs.ActiveInput.Update(msg)

	// Validate query
	if err := ops.ValidateSearchQuery(m.Inputs.ActiveInput.Value()); err != nil {
		// If query is invalid, strip the last character or clear it
		val := m.Inputs.ActiveInput.Value()
		if len(val) > 0 {
			m.Inputs.ActiveInput.SetValue(val[:len(val)-1])
		}
	}

	filter.Apply(m)
	if m.Navigation.Cursor >= len(m.Navigation.FilteredItems) {
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
	}
	return cmd
}
