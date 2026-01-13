package settings

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestFormatBool(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	on := FormatBool(true, styles)
	if !strings.Contains(on, "ON") {
		t.Errorf("Expected ON, got %s", on)
	}

	off := FormatBool(false, styles)
	if !strings.Contains(off, "OFF") {
		t.Errorf("Expected OFF, got %s", off)
	}
}
