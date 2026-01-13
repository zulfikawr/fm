package update

import (
	"testing"

	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleSearching(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputSearch
	m.Inputs.ActiveInput.SetValue("test")

	// Test enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_ = HandleSearching(msg, m)
	if m.UI.InputActive {
		t.Error("Expected search input to close on enter")
	}

	// Test esc
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputSearch
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	_ = HandleSearching(msg, m)
	if m.UI.InputActive {
		t.Error("Expected search input to close on esc")
	}
}
