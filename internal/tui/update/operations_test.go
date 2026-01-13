package update

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/constants"
	"fm/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleConfirming(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.Confirming = true
	m.Operations.ActionType = constants.ActionDelete

	// Test 'y'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_ = HandleConfirming(msg, m)
	if m.UI.Confirming {
		t.Error("Expected confirming false after 'y'")
	}

	// Test 'n'
	m.UI.Confirming = true
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	_ = HandleConfirming(msg, m)
	if m.UI.Confirming {
		t.Error("Expected confirming false after 'n'")
	}
}

func TestHandleOperationsMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// ConflictMsg
	msg := commands.ConflictMsg{Src: "s", Dst: "d"}
	_ = HandleOperationsMsg(m, msg)
	if !m.UI.Confirming {
		t.Error("Expected confirming true after ConflictMsg")
	}

	// ProgressMsg
	progMsg := commands.ProgressMsg{Label: "test", Percent: 0.5}
	_ = HandleOperationsMsg(m, progMsg)
	if !m.Operations.Progress.Visible {
		t.Error("Expected progress visible")
	}
}
