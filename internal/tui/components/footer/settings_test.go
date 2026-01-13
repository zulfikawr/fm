package footer

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderSettingsFooter(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:          80,
		SettingsCursor: 0,
		Styles:         styles,
	}

	res := renderSettingsFooter(props)
	if !strings.Contains(res, "Navigate") || !strings.Contains(res, "Reset") {
		t.Error("Expected settings footer to contain hints")
	}
}
