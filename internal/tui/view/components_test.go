package view

import (
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestFormatBool(t *testing.T) {
	s := &ViewState{
		Config: &config.Config{},
	}

	t.Run("ON", func(t *testing.T) {
		styles := theme.GetStylesheet(s.Config.ThemeIndex)
		res := FormatBool(s, true, styles)
		if !strings.Contains(res, "ON") {
			t.Error("Expected ON")
		}
	})

	t.Run("OFF", func(t *testing.T) {
		styles := theme.GetStylesheet(s.Config.ThemeIndex)
		res := FormatBool(s, false, styles)
		if !strings.Contains(res, "OFF") {
			t.Error("Expected OFF")
		}
	})
}

func TestDetermineMode(t *testing.T) {
	s := &ViewState{
		UI:       &state.UIState{},
		Progress: &state.ProgressState{},
	}

	t.Run("Progress priority", func(t *testing.T) {
		s.Progress.Visible = true
		s.UI.InputActive = true
		if DetermineMode(s) != 1 { // footer.ModeProgress = 1
			t.Error("Expected ModeProgress")
		}
	})

	t.Run("Search mode", func(t *testing.T) {
		s.Progress.Visible = false
		s.UI.InputActive = true
		s.InputMode = state.InputSearch
		if DetermineMode(s) != 2 { // footer.ModeSearching = 2
			t.Error("Expected ModeSearching")
		}
	})

	t.Run("Rename mode", func(t *testing.T) {
		s.Progress.Visible = false
		s.UI.InputActive = true
		s.InputMode = state.InputRename
		if DetermineMode(s) != 3 { // footer.ModeRenaming = 3
			t.Error("Expected ModeRenaming")
		}
	})
}
