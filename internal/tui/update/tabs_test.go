package update

import (
	"testing"

	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleTabKeys(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{FS: m.FS, Path: "/t1", SelectedPaths: make(map[string]bool)},
		{FS: m.FS, Path: "/t2", SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0

	// Test 'alt+2' to switch tab
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2"), Alt: true}
	_ = HandleKeyMsg(m, msg)
	if m.ActiveTab != 1 {
		t.Errorf("Expected active tab 1, got %d", m.ActiveTab)
	}

	// Test 'alt+t' to create tab
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true}
	_ = HandleKeyMsg(m, msg)
	if len(m.Tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(m.Tabs))
	}

	// Test 'alt+w' to close tab
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w"), Alt: true}
	_ = HandleKeyMsg(m, msg)
	if len(m.Tabs) != 2 {
		t.Errorf("Expected 2 tabs after close, got %d", len(m.Tabs))
	}

	// Test switch to non-existent
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9"), Alt: true}
	_ = HandleKeyMsg(m, msg)
	// ActiveTab should not be 8

	// Test close last tab (should not happen)
	m.Tabs = []state.Tab{{Path: "/t1"}}
	m.ActiveTab = 0
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w"), Alt: true}
	_ = HandleKeyMsg(m, msg)
	if len(m.Tabs) != 1 {
		t.Error("Should not close last tab")
	}

	// Test close with invalid active tab
	m.Tabs = []state.Tab{{Path: "/t1"}, {Path: "/t2"}}
	m.ActiveTab = 5
	_, _ = HandleCloseTab(m)
}
