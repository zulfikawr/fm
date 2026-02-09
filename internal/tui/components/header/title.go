package header

import (
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// TitleProps contains data for rendering the header title
type TitleProps struct {
	Path       string
	ActiveView context.ViewMode
	Style      theme.Stylesheet
}

// GetTitle returns the appropriate title string based on the current UI state
func GetTitle(props TitleProps) string {
	switch props.ActiveView {
	case context.ViewSettings:
		return "Settings"
	case context.ViewHelp:
		return "Help"
	case context.ViewLogs:
		return "Operation Log"
	case context.ViewClipboard:
		return "Clipboard Contents"
	case context.ViewTrash:
		return "Trash"
	default:
		return props.Path
	}
}
