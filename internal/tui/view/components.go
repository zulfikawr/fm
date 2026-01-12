package view

import "fm/internal/tui/theme"

// FormatBool formats a boolean as [ON] or [OFF] with styling
func FormatBool(s *ViewState, b bool, styles theme.Stylesheet) string {
	if b {
		return styles.KeyCol.Render("[ON]")
	}
	return styles.DimCol.Render("[OFF]")
}
