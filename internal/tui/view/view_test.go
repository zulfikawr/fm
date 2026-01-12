package view

import (
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/files"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestView(t *testing.T) {
	s := &ViewState{
		Width:     80,
		Height:    24,
		Config:    &config.Config{},
		UI:        &state.UIState{},
		Progress:  &state.ProgressState{},
		Clipboard: &state.ClipboardState{},
	}
	s.UI.PromptCache = make(map[string]string)
	styles := theme.GetStylesheet(0)

	t.Run("Basic Rendering", func(t *testing.T) {
		v := Render(s, styles)
		if v == "" {
			t.Error("View should not be empty")
		}
	})

	t.Run("Loading View", func(t *testing.T) {
		s.UI.Loading = true
		s.FilteredItems = []files.Item{}
		v := Render(s, styles)
		if !strings.Contains(v, "Loading...") {
			t.Error("Expected Loading... in view")
		}
		s.UI.Loading = false
	})
}
