package view

import (
	"fm/internal/tui/state"
)

// GetViewportHeight calculates the available viewport height efficiently
func GetViewportHeight(s *ViewState) int {
	if s.ViewportHeight > 0 {
		return s.ViewportHeight
	}
	return CalculateViewportHeightFromState(s)
}

// CalculateViewportHeight calculates viewport height from Model
func CalculateViewportHeight(m *state.Model) int {
	// App Header: 1 line
	// App Footer: 1 line
	h := m.Display.Height - 2

	// List Header: 3 lines (separator, text, separator)
	if m.Config.ShowHeader && !m.UI.SettingsOpen {
		h -= 3
	}

	if h < 1 {
		return 1
	}
	return h
}

// CalculateViewportHeightFromState calculates viewport height from ViewState
func CalculateViewportHeightFromState(s *ViewState) int {
	h := s.Height - 2
	if s.Config.ShowHeader && !s.UI.SettingsOpen {
		h -= 3
	}
	if h < 1 {
		return 1
	}
	return h
}
