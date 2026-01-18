package handlers

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

func TestSettings_ToggleLogic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.SettingsOpen = true
	m.Config.ShowHidden = false

	HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !m.Config.ShowHidden {
		t.Error("Toggle failed in direct call")
	}
}

func TestSettings_Keys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Config.ShowHidden = false
	m.Config.CaseSensitive = false
	m.UI.SettingsOpen = true
	m.Display.ViewportHeight = 20

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// 1. Test Navigation (Down) - should move to CaseSensitive
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})

	// 2. Test Toggle (Space) - should toggle CaseSensitive
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// 3. Test Close (Esc)
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	_ = tm.Quit()

	finalModel := tm.FinalModel(t).(*testModelWrapper).m
	if !finalModel.Config.CaseSensitive {
		t.Error("expected CaseSensitive to be true in final model")
	}
	if finalModel.UI.SettingsOpen {
		t.Error("expected settings to be closed in final model")
	}
}
