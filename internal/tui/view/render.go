package view

import (
	"github.com/zulfikawr/fm/internal/tui/context"

	"github.com/charmbracelet/lipgloss"
)

// Render assembles the full UI from components
func Render(m *context.Model) string {
	if m.Display.Width == 0 || m.Display.Height == 0 {
		return "Initializing..."
	}

	// Use cached layout or calculate if needed
	if m.Display.Layout.Width != m.Display.Width || m.Display.Layout.Height != m.Display.Height {
		m.Display.Layout = CalculateLayout(m)
	}
	layout := m.Display.Layout

	// 1. Render Header
	headerStr := renderHeader(m, layout)

	// 2. Render Body
	bodyStr := renderBody(m, layout)

	// 3. Render Footer
	footerStr := renderFooter(m, layout)

	return lipgloss.JoinVertical(lipgloss.Left, headerStr, bodyStr, footerStr)
}
