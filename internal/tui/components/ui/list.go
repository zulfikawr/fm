package ui

import (
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// ListProps encapsulates common rendering data for list items
type ListProps struct {
	Selected bool
	IsUp     bool
	IsCursor bool
	Width    int
	Styles   theme.Stylesheet
}

// Marker renders a selection checkbox marker.
func Marker(props ListProps) string {
	if props.IsUp {
		return "    "
	}

	style := props.Styles.DimCol.UnsetPadding().UnsetWidth()
	content := "[ ] "
	if props.Selected {
		style = props.Styles.KeyCol.UnsetPadding().UnsetWidth()
		content = "[x] "
	}

	if props.IsCursor {
		style = style.Inherit(props.Styles.SelectedItem.UnsetPadding().UnsetWidth())
	}

	return style.Render(content)
}

// ItemRow returns a styled selectable item row.
func ItemRow(content string, props ListProps) string {
	style := props.Styles.Item
	if props.IsCursor {
		style = props.Styles.SelectedItem
	}
	return style.Width(props.Width).Render(content)
}
