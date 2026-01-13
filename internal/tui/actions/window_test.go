package actions

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResizeWindow(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	ResizeWindow(m, msg)

	if m.Display.Width != 100 {
		t.Errorf("Expected width 100, got %d", m.Display.Width)
	}
	if m.Display.Height != 50 {
		t.Errorf("Expected height 50, got %d", m.Display.Height)
	}
	if m.Inputs.ActiveInput.Width <= 0 {
		t.Error("Expected active input width to be updated")
	}
}
