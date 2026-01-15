package ui

import (
	"fm/internal/tui/theme"
)

// Marker renders a selection checkbox marker.
func Marker(selected bool, isUp bool, isCursor bool, styles theme.Stylesheet) string {
	if isUp {
		return "    "
	}

	style := styles.DimCol.UnsetPadding().UnsetWidth()
	content := "[ ] "
	if selected {
		style = styles.KeyCol.UnsetPadding().UnsetWidth()
		content = "[x] "
	}

	if isCursor {
		style = style.Inherit(styles.SelectedItem.UnsetPadding().UnsetWidth())
	}

	return style.Render(content)
}

// ItemRow returns a styled selectable item row.
func ItemRow(content string, width int, isCursor bool, styles theme.Stylesheet) string {
	style := styles.Item
	if isCursor {
		style = styles.SelectedItem
	}
	return style.Width(width).Render(content)
}
