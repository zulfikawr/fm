package update

import (
	"errors"
	"testing"

	"fm/internal/tui/commands"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Test WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_ = Update(m, msg)
	if m.Display.Width != 100 {
		t.Errorf("Expected width 100, got %d", m.Display.Width)
	}

	// Test KeyMsg
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	_ = Update(m, keyMsg)

	// Test Generic Msg (Error)
	err := errors.New("test error")
	errMsg := commands.ErrorMsg{Err: err}
	_ = Update(m, errMsg)
	if m.Message.Error == nil {
		t.Error("Expected error to be handled and set in model")
	}
}
