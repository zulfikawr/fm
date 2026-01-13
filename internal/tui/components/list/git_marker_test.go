package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderGitMarker(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{Styles: styles}

	item := core.Item{GitStatus: "M"}
	res := renderGitMarker(props, item, false)
	if !strings.Contains(res, "M") {
		t.Errorf("Expected M, got %s", res)
	}
}
