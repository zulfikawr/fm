package header

import (
	"fm/internal/tui/theme"
)

// TitleProps contains data for rendering the header title
type TitleProps struct {
	Path          string
	SettingsOpen  bool
	LogOpen       bool
	ClipboardOpen bool
	Style         theme.Stylesheet
}

// GetTitle returns the appropriate title string based on the current UI state
func GetTitle(props TitleProps) string {
	if props.SettingsOpen {
		return "Settings"
	}
	if props.LogOpen {
		return "Operation Log"
	}
	if props.ClipboardOpen {
		return "Clipboard Contents"
	}
	return props.Path
}
