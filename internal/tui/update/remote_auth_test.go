package update

import (
	"testing"

	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleRemoteAuth(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputAuth
	m.Remote.Host = "host"
	m.Inputs.ActiveInput.SetValue("password")

	// Test enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_ = HandleRemoteAuth(msg, m)
	if m.UI.InputActive {
		t.Error("Expected auth input to close on enter")
	}
}
