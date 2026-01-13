package filter

import (
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
)

func TestApply(t *testing.T) {
	items := []core.Item{
		{Name: "↑ ..", IsUp: true},
		{Name: "file1.txt"},
		{Name: "document.pdf"},
		{Name: "image.png"},
	}

	m := &state.Model{
		Navigation: state.NavigationState{
			Items: items,
		},
		Inputs: state.InputState{
			ActiveInput: textinput.New(),
			Mode:        state.InputSearch,
		},
	}

	t.Run("No Search", func(t *testing.T) {
		m.UI.InputActive = false
		Apply(m)
		if len(m.Navigation.FilteredItems) != 4 {
			t.Errorf("Expected 4 items, got %d", len(m.Navigation.FilteredItems))
		}
	})

	t.Run("Search Match", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.ActiveInput.SetValue("file")
		Apply(m)
		// Up item + match
		if len(m.Navigation.FilteredItems) != 2 {
			t.Errorf("Expected 2 items, got %d", len(m.Navigation.FilteredItems))
		}
	})

	t.Run("Search Case Insensitive", func(t *testing.T) {
		m.UI.InputActive = true
		m.Inputs.ActiveInput.SetValue("DOCUMENT")
		Apply(m)
		if len(m.Navigation.FilteredItems) != 2 {
			t.Errorf("Expected 2 items, got %d", len(m.Navigation.FilteredItems))
		}
	})
}
