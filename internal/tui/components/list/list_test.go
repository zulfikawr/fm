package list

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/theme"
)

func TestRender(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	items := []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt", IsDir: false},
		{Name: "dir1", Path: "/test/dir1", IsDir: true},
	}

	props := Props{
		Width:      80,
		Height:     10,
		Cursor:     0,
		Offset:     0,
		Items:      items,
		ShowHeader: true,
		Styles:     styles,
	}

	result := Render(props)

	if !strings.Contains(result, "file1.txt") {
		t.Error("Result should contain file1.txt")
	}
	if !strings.Contains(result, "dir1") {
		t.Error("Result should contain dir1")
	}
	if !strings.Contains(result, "Name") {
		t.Error("Result should contain Name header")
	}
}
