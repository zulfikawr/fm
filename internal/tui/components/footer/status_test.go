package footer

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderNormalFooter(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:         80,
		Cursor:        0,
		TotalItems:    10,
		SelectedCount: 2,
		Styles:        styles,
	}

	res := renderNormalFooter(props)
	if !strings.Contains(res, "1/10") || !strings.Contains(res, "2 selected") {
		t.Errorf("Unexpected normal footer output: %q", res)
	}
}

func TestAssembleFooterContent(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	parts := []string{"Part1", "Part2"}

	res := assembleFooterContent(parts, 80, styles)
	if !strings.Contains(res, "Part1") || !strings.Contains(res, "Part2") {
		t.Error("Expected parts in assembled content")
	}

	// Test truncation
	res = assembleFooterContent([]string{"Very Long Content That Should Be Truncated"}, 10, styles)
	if !strings.Contains(res, "...") {
		t.Error("Expected truncation ellipsis")
	}
}

func TestAddReadOnlyIndicator(t *testing.T) {
	// wait, this is in header breadcrumb.go usually. Let me check footer.go
}
