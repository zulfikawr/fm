package ui

import (
	"fm/internal/tui/theme"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// Column defines a table column's properties
type Column struct {
	Title string
	Width int
	Align string // "left", "right"
}

// RenderHeader renders a standardized table header with separators
func RenderHeader(width int, columns []Column, gap int, styles theme.Stylesheet) string {
	var parts []string
	for _, col := range columns {
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

	headerContent := strings.Join(parts, strings.Repeat(" ", gap))

	// Ensure it starts with a space if there's room
	if width > lipgloss.Width(headerContent) {
		headerContent = " " + headerContent
	}

	// Top and bottom separators
	sep := styles.Separator.Width(width).Render(strings.Repeat("-", width))
	headerText := styles.ListHeader.Width(width).Render(headerContent)

	return lipgloss.JoinVertical(lipgloss.Left, sep, headerText, sep)
}
