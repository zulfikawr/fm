package ui

import (
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Toggle renders a boolean ON/OFF switch with theme awareness.
func Toggle(value bool, styles theme.Stylesheet) string {
	if value {
		return styles.KeyCol.Render("[ON]")
	}
	return styles.DimCol.Render("[OFF]")
}

// ToggleProps encapsulates data for rendering a labeled toggle
type ToggleProps struct {
	Label  string
	Value  bool
	Width  int
	Styles theme.Stylesheet
}

// ToggleLabeled renders a label followed by a toggle switch.
func ToggleLabeled(props ToggleProps) string {
	labelPart := props.Label + ":"
	togglePart := Toggle(props.Value, props.Styles)

	// Use FlexRow to handle the layout and padding
	return FlexRow(props.Width, labelPart, " ", togglePart)
}
