package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRenderNamePart(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{Styles: styles}
	layout := Layout{NameWidth: 10}

	// Directory
	item := core.Item{Name: "dir", IsDir: true}
	res := renderNamePart(props, item, false, layout)
	if !strings.Contains(res, "dir/") {
		t.Errorf("Expected dir/, got %s", res)
	}

	// Truncation
	item = core.Item{Name: "verylongfilename"}
	res = renderNamePart(props, item, false, layout)
	if !strings.Contains(res, "…") {
		t.Errorf("Expected truncation, got %s", res)
	}
}
