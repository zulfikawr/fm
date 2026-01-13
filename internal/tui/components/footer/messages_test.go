package footer

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderMessage(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:   80,
		Message: "Test Message",
		Styles:  styles,
	}

	res := renderMessage(props)
	if !strings.Contains(res, "Test Message") {
		t.Errorf("Expected message 'Test Message', got %q", res)
	}
}
