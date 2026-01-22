package ui

import (
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// MenuProps encapsulates common data for rendering menu rows
type MenuProps struct {
	Label    string
	Value    string
	Width    int
	IsCursor bool
	Styles   theme.Stylesheet
}

// SelectableRow renders a row that highlights when it is the active cursor position.
func SelectableRow(content string, props MenuProps) string {
	style := props.Styles.Item
	if props.IsCursor {
		style = props.Styles.SelectedItem
	}
	return style.Width(props.Width).Render(content)
}

// MenuRow renders a standardized label-value pair for menu items.
func MenuRow(props MenuProps) string {
	content := props.Label + ": " + props.Value
	return SelectableRow(content, props)
}
