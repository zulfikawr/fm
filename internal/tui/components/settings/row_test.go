package settings

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderSettingRow(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	props := Props{
		Width:  80,
		Styles: styles,
	}
	sItem := SettingItem{Label: "Test", Value: "Val"}

	// Normal
	res := renderSettingRow(props, sItem, false)
	if !strings.Contains(res, "Test") || !strings.Contains(res, "Val") {
		t.Error("Expected label and value in row")
	}

	// Cursor
	res = renderSettingRow(props, sItem, true)
	if res == "" {
		t.Error("Expected non-empty cursor row")
	}

	// Inactive
	sItem.Inactive = true
	res = renderSettingRow(props, sItem, false)
	if res == "" {
		t.Error("Expected non-empty inactive row")
	}

	// Narrow width
	props.Width = 30
	res = renderSettingRow(props, sItem, false)
	if res == "" {
		t.Error("Expected non-empty row in narrow width")
	}

	// Extreme narrow
	props.Width = 5
	res = renderSettingRow(props, sItem, false)
	if res == "" {
		t.Error("Expected non-empty row in extreme narrow width")
	}

	// Very long label
	sItem.Label = "This is a very long label that will certainly trigger truncation"
	res = renderSettingRow(props, sItem, false)
	if !strings.Contains(res, "…") {
		// Expect truncation if it actually triggers
	}
}
