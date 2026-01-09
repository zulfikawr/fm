package tui

import (
	"strings"
	"testing"
)

func TestViewSettings(t *testing.T) {
	m := NewModel("/")
	m.cfg.ShowSize = false
	m.settingsOpen = true
	m.width = 80

	t.Run("Settings Rendering", func(t *testing.T) {
		header := m.renderHeader()
		footer := m.renderFooter()
		content := m.renderSettingsList(header, footer)

		if !strings.Contains(content, "File Operations") {
			t.Error("Settings list should contain group headers")
		}
		if !strings.Contains(content, "Size Format") {
			t.Error("Settings list should contain Size Format")
		}
	})
}
