package ui

import (
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// SelectableRow renders a row that highlights when it is the active cursor position.
func SelectableRow(content string, width int, isCursor bool, styles theme.Stylesheet) string {
	style := styles.Item
	if isCursor {
		style = styles.SelectedItem
	}
	return style.Width(width).Render(content)
}

// MenuRow renders a standardized label-value pair for menu items.
func MenuRow(label, value string, width int, isCursor bool, styles theme.Stylesheet) string {
	content := label + ": " + value
	return SelectableRow(content, width, isCursor, styles)
}
