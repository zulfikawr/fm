package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderPermIndicator(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{Styles: styles}

	// Read-only
	item := core.Item{CanWrite: false}
	res := renderPermIndicator(props, item, false)
	if !strings.Contains(res, "!") {
		t.Errorf("Expected !, got %s", res)
	}

	// Writable
	item.CanWrite = true
	res = renderPermIndicator(props, item, false)
	if strings.Contains(res, "!") {
		t.Errorf("Expected no !, got %s", res)
	}
}
