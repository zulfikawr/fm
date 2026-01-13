package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderHeaderRows(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:      80,
		ShowHeader: true,
		Items:      []core.Item{{Name: "f1"}},
		Styles:     styles,
	}
	layout := CalculateLayout(props)

	rows := renderHeaderRows(props, layout)
	if len(rows) != 3 {
		t.Errorf("Expected 3 header rows, got %d", len(rows))
	}
	if !strings.Contains(rows[1], "Name") {
		t.Error("Expected Name in header")
	}
}
