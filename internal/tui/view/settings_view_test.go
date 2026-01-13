package view

import (
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"
	"fm/internal/tui/theme"
)

func TestRenderSettingsView(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	s := &ViewState{
		Width:  80,
		Height: 40,
		UI:     &state.UIState{},
		Config: &config.Config{},
	}

	res := RenderSettingsView(s, "H", "F", styles)
	if !strings.Contains(res, "File Operations") {
		t.Error("Expected group title")
	}
}

func TestUpdateSettingsScroll(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Display.Width = 80
	m.Display.Height = 24
	m.Settings.Cursor = 20 // Keybindings group

	UpdateSettingsScroll(m)
	// Should not panic and should update offset
}
