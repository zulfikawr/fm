package file

import (
	"strings"
	"testing"

	"fm/internal/files/core"
	"fm/internal/testutil"
	"fm/internal/tui/theme"
)

func TestList_Render(t *testing.T) {
	styles := theme.GetStylesheet(0)

	items := []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "file2.txt", Path: "/test/file2.txt"},
	}

	props := Props{
		Width:  80,
		Height: 10,
		Items:  items,
		Styles: styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)

	if !strings.Contains(stripped, "file1.txt") {
		t.Errorf("expected file1.txt in view, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "file2.txt") {
		t.Errorf("expected file2.txt in view, got:\n%s", stripped)
	}
}
