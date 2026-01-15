package ui

import (
	"fm/internal/tui/theme"
)

// Toggle renders a boolean ON/OFF switch with theme awareness.
func Toggle(value bool, styles theme.Stylesheet) string {
	if value {
		return styles.KeyCol.Render("[ON]")
	}
	return styles.DimCol.Render("[OFF]")
}

// ToggleLabeled renders a label followed by a toggle switch.
func ToggleLabeled(label string, value bool, width int, styles theme.Stylesheet) string {
	labelPart := label + ":"
	togglePart := Toggle(value, styles)

	// Use FlexRow to handle the layout and padding
	return FlexRow(width, labelPart, " ", togglePart)
}
