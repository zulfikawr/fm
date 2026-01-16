package view

import (
	"fm/internal/tui/context"
)

// CalculateLayout computes heights for all UI areas based on current state
func CalculateLayout(m *context.Model) context.Layout {
	// App Header: 1 line
	headerHeight := 1

	// App Footer: 1 line
	footerHeight := 1

	bodyHeight := m.Display.Height - headerHeight - footerHeight
	if bodyHeight < 0 {
		bodyHeight = 0
	}

	return context.Layout{
		Width:        m.Display.Width,
		Height:       m.Display.Height,
		HeaderHeight: headerHeight,
		FooterHeight: footerHeight,
		BodyHeight:   bodyHeight,
	}
}

// GetViewportHeight returns the height available for the main content area
func GetViewportHeight(m *context.Model) int {
	return m.Display.ViewportHeight
}
