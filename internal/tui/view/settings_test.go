package view

import (
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestViewSettings(t *testing.T) {
	s := &ViewState{
		Width:     80,
		Height:    40,
		Separator: "/",
		Config:    &config.Config{},
		UI:        &state.UIState{},
	}
	s.Config.ShowSize = false
	s.UI.SettingsOpen = true
	styles := theme.GetStylesheet(s.Config.ThemeIndex)

	t.Run("Settings Rendering", func(t *testing.T) {
		headerProps := header.Props{
			Width:        s.Width,
			SettingsOpen: s.UI.SettingsOpen,
			Styles:       styles,
		}
		h := header.Render(headerProps)

		footerProps := footer.Props{
			Mode:   footer.ModeSettings,
			Width:  s.Width,
			Styles: styles,
		}
		f := footer.Render(footerProps)

		content := RenderSettingsList(s, h, f, styles)

		if !strings.Contains(content, "File Operations") {
			t.Error("Settings list should contain group headers")
		}
		if !strings.Contains(content, "Size Format") {
			t.Error("Settings list should contain Size Format")
		}
	})
}
