package list

import (
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderRow(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{Styles: styles, Width: 80}
	layout := Layout{NameWidth: 20}
	item := core.Item{Name: "f1"}

	res := renderRow(props, item, false, layout)
	if res == "" {
		t.Error("Expected non-empty row")
	}

	res = renderRow(props, item, true, layout)
	if res == "" {
		t.Error("Expected non-empty cursor row")
	}
}
