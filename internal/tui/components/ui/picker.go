package ui

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Picker renders a cycling option selection indicator.
func Picker(value string, styles theme.Stylesheet) string {
	content := fmt.Sprintf("< %s >", value)
	return styles.SecondaryCol.Render(content)
}

// PickerProps encapsulates data for rendering a labeled picker
type PickerProps struct {
	Label  string
	Value  string
	Width  int
	Styles theme.Stylesheet
}

// PickerLabeled renders a label followed by an option picker.
func PickerLabeled(props PickerProps) string {
	labelPart := props.Label + ":"
	pickerPart := Picker(props.Value, props.Styles)

	return FlexRow(props.Width, labelPart, " ", pickerPart)
}
