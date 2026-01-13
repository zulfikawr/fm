package footer

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderPaginationInfo(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Empty", func(t *testing.T) {
		res := renderPaginationInfo(PaginationInfo{Total: 0}, styles)
		if !strings.Contains(res, "-/0") {
			t.Errorf("Expected -/0, got %q", res)
		}
	})

	t.Run("Normal", func(t *testing.T) {
		res := renderPaginationInfo(PaginationInfo{Current: 2, Total: 10}, styles)
		if !strings.Contains(res, "3/10") {
			t.Errorf("Expected 3/10, got %q", res)
		}
	})

	t.Run("Up Dir", func(t *testing.T) {
		res := renderPaginationInfo(PaginationInfo{Current: -1, Total: 10}, styles)
		if !strings.Contains(res, "-/10") {
			t.Errorf("Expected -/10, got %q", res)
		}
	})
}
