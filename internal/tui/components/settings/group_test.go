package settings

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderGroups(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:  80,
		Styles: styles,
		Groups: []SettingGroup{
			{Title: "Title", Settings: []SettingItem{{Label: "L", Value: "V"}}},
		},
	}

	rows := renderGroups(props)
	if len(rows) < 3 {
		t.Errorf("Expected at least 3 rows (empty, title, setting), got %d", len(rows))
	}
	if rows[0] != "" {
		t.Error("Expected leading empty row")
	}
	if !strings.Contains(rows[1], "Title") {
		t.Errorf("Expected Title in rows[1], got %q", rows[1])
	}
}
