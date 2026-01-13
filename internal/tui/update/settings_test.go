package update

import (
	"testing"

	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleSettingsUpdate(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.SettingsOpen = true
	m.Settings.Cursor = 0

	// Test down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	_ = HandleSettingsUpdate(msg, m)
	if m.Settings.Cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", m.Settings.Cursor)
	}

	// Test up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	_ = HandleSettingsUpdate(msg, m)
	if m.Settings.Cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", m.Settings.Cursor)
	}

	// Test exit settings
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")}
	_ = HandleSettingsUpdate(msg, m)
	if m.UI.SettingsOpen {
		t.Error("Expected settings to close")
	}

	// Test toggle (Space/Enter)
	m.UI.SettingsOpen = true
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}
	_ = HandleSettingsUpdate(msg, m)
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	_ = HandleSettingsUpdate(msg, m)

	// Test cycle (Left/Right)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	_ = HandleSettingsUpdate(msg, m)
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	_ = HandleSettingsUpdate(msg, m)
}
