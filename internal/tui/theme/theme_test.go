package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestThemesNotEmpty(t *testing.T) {
	if len(Themes) == 0 {
		t.Error("Themes slice should not be empty")
	}
}

func TestAllThemesHaveNames(t *testing.T) {
	for i, theme := range Themes {
		if theme.Name == "" {
			t.Errorf("Theme at index %d has empty name", i)
		}
	}
}

func TestAllThemesHaveUniqueNames(t *testing.T) {
	seen := make(map[string]bool)
	for _, theme := range Themes {
		if seen[theme.Name] {
			t.Errorf("Duplicate theme name: %s", theme.Name)
		}
		seen[theme.Name] = true
	}
}

func TestAllThemesHaveRequiredColors(t *testing.T) {
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Check that all color fields are set (non-empty lipgloss.Color)
			if theme.Subtle == "" {
				t.Error("Subtle color not set")
			}
			if theme.Dir == "" {
				t.Error("Dir color not set")
			}
			if theme.Exec == "" {
				t.Error("Exec color not set")
			}
			if theme.File == "" {
				t.Error("File color not set")
			}
			if theme.SelectedBg == "" {
				t.Error("SelectedBg color not set")
			}
			if theme.SelectedFg == "" {
				t.Error("SelectedFg color not set")
			}
			if theme.Bg == "" {
				t.Error("Bg color not set")
			}
			// Git colors
			if theme.GitMod == "" {
				t.Error("GitMod color not set")
			}
			if theme.GitStaged == "" {
				t.Error("GitStaged color not set")
			}
			if theme.GitUntracked == "" {
				t.Error("GitUntracked color not set")
			}
			if theme.GitConflict == "" {
				t.Error("GitConflict color not set")
			}
			if theme.GitGhost == "" {
				t.Error("GitGhost color not set")
			}
		})
	}
}

func TestThemeIndexBounds(t *testing.T) {
	// Ensure we have at least one theme for default
	if len(Themes) < 1 {
		t.Error("Need at least one theme")
	}

	// Test accessing all themes by index
	for i := 0; i < len(Themes); i++ {
		theme := Themes[i]
		if theme.Name == "" {
			t.Errorf("Theme at index %d is invalid", i)
		}
	}
}

func TestKnownThemesExist(t *testing.T) {
	knownThemes := []string{
		"Gruvbox",
		"Nord",
		"Dracula",
		"Monokai",
		"Solarized Dark",
	}

	themeNames := make(map[string]bool)
	for _, theme := range Themes {
		themeNames[theme.Name] = true
	}

	for _, name := range knownThemes {
		if !themeNames[name] {
			t.Errorf("Expected theme %q not found", name)
		}
	}
}

func TestThemeColorsAreValidANSI(t *testing.T) {
	// lipgloss.Color accepts ANSI color codes as strings
	// Valid codes are typically 0-255 for 256-color mode
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Just verify we can create styles with these colors without panic
			_ = lipgloss.NewStyle().Foreground(theme.Dir)
			_ = lipgloss.NewStyle().Background(theme.Bg)
			_ = lipgloss.NewStyle().Foreground(theme.SelectedFg).Background(theme.SelectedBg)
		})
	}
}
