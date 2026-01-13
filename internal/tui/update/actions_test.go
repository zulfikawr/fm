package update

import (
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleAction(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.Items = []core.Item{
		{Name: "f1", Path: "/test/f1"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items
	m.Navigation.Cursor = 0

	// Test Space (Selection)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}
	_ = HandleAction(msg, m)
	if !m.Navigation.Items[0].Selected {
		t.Error("Expected item to be selected")
	}

	// Test 's' (Sort)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	_ = HandleAction(msg, m)

	// Test '/' (Search)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	_ = HandleAction(msg, m)
	if m.Inputs.Mode != state.InputSearch {
		t.Error("Expected InputSearch mode")
	}
	m.UI.InputActive = false

	// Test 'c' (Copy)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}
	_ = HandleAction(msg, m)
	if len(m.Operations.Clipboard.Paths) == 0 {
		t.Error("Expected items in clipboard")
	}

	// Test 'x' (Cut)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	_ = HandleAction(msg, m)
	if !m.Operations.Clipboard.IsCut {
		t.Error("Expected IsCut true")
	}

	// Test 'v' (Paste)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")}
	_ = HandleAction(msg, m)

	// Test 'r' (Rename)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	_ = HandleAction(msg, m)
	if m.Inputs.Mode != state.InputRename {
		t.Error("Expected InputRename mode")
	}
	m.UI.InputActive = false

	// Test 'd' (Delete)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	_ = HandleAction(msg, m)

	// Test '.' (Settings)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")}
	_ = HandleAction(msg, m)
	if !m.UI.SettingsOpen {
		t.Error("Expected SettingsOpen true")
	}
	m.UI.SettingsOpen = false

	// Test 'g' (Goto)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	_ = HandleAction(msg, m)
	if m.Inputs.Mode != state.InputGoto {
		t.Error("Expected InputGoto mode")
	}

	// Test ReadOnly
	m.Display.ReadOnly = true
	m.UI.InputActive = false
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	_ = HandleAction(msg, m)
	// Error msg should be returned
}
