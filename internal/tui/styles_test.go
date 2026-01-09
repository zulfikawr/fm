package tui

import (
	"testing"
)

func TestThemeApplication(t *testing.T) {
	for i, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			styles := NewStylesheet(theme)
			// Check some basic style properties
			if styles.Header.GetForeground() != theme.Dir {
				t.Errorf("Expected header foreground %v, got %v", theme.Dir, styles.Header.GetForeground())
			}

			m := NewModel("/")
			m.cfg.ThemeIndex = i
			m.styles = styles
			m.width = 80

			header := m.renderHeader()
			if header == "" {
				t.Error("Header should not be empty")
			}
		})
	}
}
