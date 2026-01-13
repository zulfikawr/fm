package settings

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRender(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	groups := []SettingGroup{
		{
			Title: "Group 1",
			Settings: []SettingItem{
				{Label: "Setting 1", Value: "Value 1"},
			},
		},
	}

	props := Props{
		Width:  80,
		Height: 10,
		Cursor: 0,
		Offset: 0,
		Groups: groups,
		Styles: styles,
	}

	result := Render(props)

	if !strings.Contains(result, "Group 1") {
		t.Error("Result should contain Group 1")
	}
	if !strings.Contains(result, "Setting 1") {
		t.Error("Result should contain Setting 1")
	}
	if !strings.Contains(result, "Value 1") {
		t.Error("Result should contain Value 1")
	}

	// Test negative height
	props.Height = -1
	if Render(props) != "" {
		t.Error("Expected empty string for negative height")
	}

	// Test offset
	props.Height = 10
	props.Offset = 2
	result = Render(props)
	if result == "" {
		t.Error("Expected non-empty result with offset")
	}

	// Test offset out of bounds
	props.Offset = 100
	result = Render(props)
	if result == "" {
		// it might be empty lines but should not panic
	}
}
