package ui

import (
	"github.com/charmbracelet/lipgloss"

	"fm/internal/tui/theme"
)

// Truncate ensures a string fits within a maximum width, adding an ellipsis if needed.
// It handles multi-byte characters correctly using lipgloss.Width.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}

	currentWidth := lipgloss.Width(s)
	if currentWidth <= width {
		return s
	}

	// If too small for ellipsis, just return what fits
	if width < 3 {
		return lipgloss.NewStyle().MaxWidth(width).Render(s)
	}

	return lipgloss.NewStyle().MaxWidth(width-1).Render(s) + "…"
}

// Dim renders text using the theme's diminished style
func Dim(styles theme.Stylesheet, content string) string {
	return styles.DimCol.Render(content)
}

// Bold renders text in bold using the theme's primary color
func Bold(styles theme.Stylesheet, content string) string {
	return styles.KeyCol.Bold(true).Render(content)
}

// Highlight renders text using the theme's selected/highlight style
func Highlight(styles theme.Stylesheet, content string) string {
	return styles.SelectedItem.Render(content)
}

// Success renders text using the theme's success style
func Success(styles theme.Stylesheet, content string) string {
	return styles.Success.Render(content)
}

// Error renders text using the theme's error style
func Error(styles theme.Stylesheet, content string) string {
	return styles.Error.Render(content)
}
