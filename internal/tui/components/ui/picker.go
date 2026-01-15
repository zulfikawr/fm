package ui

import (
	"fm/internal/tui/theme"
	"fmt"
)

// Picker renders a cycling option selection indicator.
func Picker(value string, styles theme.Stylesheet) string {
	content := fmt.Sprintf("< %s >", value)
	return styles.KeyCol.Render(content)
}

// PickerLabeled renders a label followed by an option picker.
func PickerLabeled(label string, value string, width int, styles theme.Stylesheet) string {
	labelPart := label + ":"
	pickerPart := Picker(value, styles)

	return FlexRow(width, labelPart, " ", pickerPart)
}
