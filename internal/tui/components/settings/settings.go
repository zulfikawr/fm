package settings

import (
	"strings"

	"fm/internal/tui/theme"
)

// SettingItem represents a single setting or keybinding row
type SettingItem struct {
	Label    string
	Value    string
	Inactive bool
}

// SettingGroup represents a categorized group of settings
type SettingGroup struct {
	Title    string
	Settings []SettingItem
}

// Props contains all data needed to render the settings list
type Props struct {
	Width  int
	Height int
	Cursor int
	Offset int
	Groups []SettingGroup
	Styles theme.Stylesheet
}

// Render renders the complete settings view
func Render(props Props) string {
	if props.Height < 0 {
		return ""
	}

	rows := renderGroups(props)

	// Apply scroll offset
	if props.Offset > 0 && props.Offset < len(rows) {

		rows = rows[props.Offset:]
	} else if props.Offset >= len(rows) {

		rows = []string{}
	}

	// Ensure we fill the viewport
	if len(rows) > props.Height {

		rows = rows[:props.Height]
	} else {
		for i := len(rows); i < props.Height; i++ {

			rows = append(rows, "")
		}
	}

	return strings.Join(rows, "\n")
}
