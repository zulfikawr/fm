package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

// WindowProps contains all data needed to render a window container.
type WindowProps struct {
	Width  int
	Height int
	Title  string
	Styles theme.Stylesheet
}

// Window wraps content in a standardized bordered container.
func Window(content string, props WindowProps) string {
	if props.Height <= 0 || props.Width <= 0 {
		return ""
	}

	// Calculate inner dimensions
	innerPadding := 1
	innerWidth := props.Width - (innerPadding * 2)
	innerHeight := props.Height

	// Placeholder for more complex window logic (borders etc)
	// For now, we focus on consistent padding and height management

	style := lipgloss.NewStyle().
		Width(innerWidth).
		Height(innerHeight).
		Padding(0, innerPadding)

	return style.Render(content)
}
