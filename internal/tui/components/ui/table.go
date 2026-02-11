package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Column defines a table column's properties
type Column struct {
	Title string
	Width int
	Align string // "left", "right"
}

// HeaderProps encapsulates data for rendering table headers
type HeaderProps struct {
	Width   int
	Columns []Column
	Gap     int
	Styles  theme.Stylesheet
}

// RenderHeader renders a standardized table header with separators
func RenderHeader(props HeaderProps) string {
	var parts []string
	for i := range props.Columns {
		col := props.Columns[i]
		title := col.Title
		if len(title) > col.Width {
			title = Truncate(title, col.Width)
		}

		var part string
		if col.Align == "right" {
			part = fmt.Sprintf("%*s", col.Width, title)
		} else {
			part = fmt.Sprintf("%-*s", col.Width, title)
		}
		parts = append(parts, part)
	}

	headerContent := strings.Join(parts, strings.Repeat(" ", props.Gap))

	// Ensure it starts with a space if there's room
	if props.Width > lipgloss.Width(headerContent) {
		headerContent = " " + headerContent
	}

	// Top and bottom separators
	sep := props.Styles.Separator.Width(props.Width).Render(strings.Repeat("-", props.Width))
	headerText := props.Styles.ListHeader.Width(props.Width).Render(headerContent)

	return lipgloss.JoinVertical(lipgloss.Left, sep, headerText, sep)
}
