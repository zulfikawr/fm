package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderMetaPart(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		ShowSize:         true,
		ShowDateModified: true,
		Styles:           styles,
	}
	layout := Layout{ColumnGap: 2, SizeWidth: 10, DateWidth: 10}

	item := core.Item{FormattedSize: "1KB", FormattedDate: "2026-01-12"}
	res := renderMetaPart(props, item, false, layout)
	if !strings.Contains(res, "1KB") || !strings.Contains(res, "2026-01-12") {
		t.Errorf("Expected size and date, got %s", res)
	}
}
