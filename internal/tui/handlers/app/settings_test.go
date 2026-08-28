package app_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettings_ToggleLogic(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	config.SetConfigPath(filepath.Join(tmpDir, "config.json"))
	defer config.SetConfigPath("")

	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.ActiveView = tuictx.ViewSettings
	m.Config.UI.ShowHidden = false

	app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	if !m.Config.UI.ShowHidden {
		t.Error("Toggle failed in direct call")
	}

	t.Run("Navigation Keys", func(t *testing.T) {
		m.Settings.Cursor = 1
		app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if m.Settings.Cursor != 0 {
			t.Error("expected cursor up")
		}

		app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.Settings.Cursor != 1 {
			t.Error("expected cursor down")
		}
	})

	t.Run("Toggle Prev", func(t *testing.T) {
		m.Settings.Cursor = 4 // Editor
		initial := m.Config.External.EditorIndex
		app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if m.Config.External.EditorIndex == initial && len(constants.Editors) > 1 {
			t.Error("expected editor index change")
		}
	})

	t.Run("Toggle All Cases", func(t *testing.T) {
		for i := 0; i <= 13; i++ {
			m.Settings.Cursor = i
			app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		}
	})

	t.Run("Toggle Prev All Cases", func(t *testing.T) {
		idxs := []int{4, 9, 11, 13}
		for j := range idxs {
			m.Settings.Cursor = idxs[j]
			app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		}
	})

	t.Run("Reset Action", func(t *testing.T) {
		app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		if m.Operations.ActionType != constants.ActionResetSettings {
			t.Error("expected reset-settings action type")
		}
		if !m.UI.Confirming {
			t.Error("expected confirming state")
		}
	})
}

func TestConfirmSettingsReset(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Config.UI.ShowHidden = false

	app.ConfirmSettingsReset(m)
	// Default is true in DefaultConfig()
	if !m.Config.UI.ShowHidden {
		t.Error("expected ShowHidden to be reset to default (true)")
	}
}

func TestSettings_KeybindingCursorEditsHighlightedAction(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	config.SetConfigPath(filepath.Join(tmpDir, "config.json"))
	defer config.SetConfigPath("")

	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.UI.ActiveView = tuictx.ViewSettings

	// The first keybinding row is the navigation action "open".
	m.Settings.Cursor = 17
	app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.UI.InputActive || m.Inputs.Mode != tuictx.InputKeybinding {
		t.Fatal("expected keybinding input mode")
	}
	if got, want := m.Inputs.ActiveInput.Prompt, "Bind Open / Enter: "; got != want {
		t.Fatalf("highlighted row opened the wrong keybinding: got prompt %q, want %q", got, want)
	}
	if got, want := m.Inputs.ActiveInput.Value(), "enter, l, right"; got != want {
		t.Fatalf("highlighted row loaded the wrong shortcuts: got %q, want %q", got, want)
	}

	m.Inputs.ActiveInput.SetValue("ctrl+o")
	app.HandleSettings(m, tea.KeyMsg{Type: tea.KeyEnter})

	if got := config.GetKeybindingForAction("open", m.Config.Keybindings); !slices.Equal(got, []string{"ctrl+o"}) {
		t.Fatalf("expected highlighted open action to be updated, got %v", got)
	}
	if got := config.GetKeybindingForAction("quit", m.Config.Keybindings); !slices.Equal(got, []string{"ctrl+c"}) {
		t.Fatalf("unselected quit action was unexpectedly changed: %v", got)
	}
}
