package update

import (
	"testing"

	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleWindowMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_ = HandleWindowMsg(m, msg)
	if m.Display.Width != 100 {
		t.Errorf("Expected width 100, got %d", m.Display.Width)
	}
}
