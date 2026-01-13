package footer

import (
	"strings"
	"testing"

	"fm/internal/files/sorting"
	"fm/internal/tui/theme"
)

func TestRenderSortMode(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	res := renderSortMode(sorting.SortName, styles)
	if !strings.Contains(res, "Name") {
		t.Errorf("Expected Name in sort mode, got %q", res)
	}
}
