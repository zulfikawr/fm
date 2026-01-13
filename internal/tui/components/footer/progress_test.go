package footer

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderProgressBar(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Normal", func(t *testing.T) {
		res := renderProgressBar("Label", 0.5, 80, styles)
		if !strings.Contains(res, "50%") || !strings.Contains(res, "#") {
			t.Errorf("Unexpected progress bar output: %q", res)
		}
	})

	t.Run("Narrow", func(t *testing.T) {
		res := renderProgressBar("Very Long Label That Should Be Truncated", 0.5, 15, styles)
		if !strings.Contains(res, "50%") {
			t.Errorf("Unexpected progress bar output: %q", res)
		}
	})

	t.Run("Extremely Narrow", func(t *testing.T) {
		res := renderProgressBar("Label", 0.5, 5, styles)
		if res == "" {
			t.Error("Expected non-empty output even if extremely narrow")
		}
	})

	t.Run("Negative Percent", func(t *testing.T) {
		res := renderProgressBar("Label", -1.0, 80, styles)
		if !strings.Contains(res, "0%") {
			t.Error("Expected 0% for negative percentage")
		}
	})

	t.Run("Over 100 Percent", func(t *testing.T) {
		res := renderProgressBar("Label", 2.0, 80, styles)
		if !strings.Contains(res, "100%") {
			t.Error("Expected 100% for over 100 percentage")
		}
	})
}
