package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FlexRow joins multiple items into a single line, ensuring the total width is met.
func FlexRow(width int, items ...string) string {
	if len(items) == 0 {
		return strings.Repeat(" ", width)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Left, items...)
	return lipgloss.NewStyle().Width(width).Render(content)
}

// Spacer returns a string of spaces of the given width
func Spacer(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(" ", width)
}

// JoinWithGaps joins items with a fixed gap between them
func JoinWithGaps(gap int, items ...string) string {
	sep := strings.Repeat(" ", gap)
	return strings.Join(items, sep)
}
