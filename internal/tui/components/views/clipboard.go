package views

import (
	"strings"

	"fm/internal/files/core"
	"fm/internal/tui/components/messages"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
)

// ClipboardProps contains all data needed to render the clipboard view
type ClipboardProps struct {
	Width    int
	Height   int
	Cursor   int
	Offset   int
	Paths    []string
	SourceFS core.FileSystem
	IsCut    bool
	Style    theme.Stylesheet
}

// RenderClipboard renders the clipboard manager view
func RenderClipboard(props ClipboardProps) string {
	if props.Height <= 0 {
		return ""
	}

	if len(props.Paths) == 0 {
		return renderEmptyClipboard(props.Width, props.Height)
	}

	var rows []string
	viewportHeight := props.Height

	// Calculate end of viewport
	end := props.Offset + viewportHeight
	if end > len(props.Paths) {
		end = len(props.Paths)
	}

	for i := props.Offset; i < end; i++ {
		path := props.Paths[i]
		rows = append(rows, renderClipboardRow(props, path, i == props.Cursor))
	}
	// Fill remaining space
	for len(rows) < props.Height {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func renderClipboardRow(props ClipboardProps, path string, selected bool) string {
	prefix := " [Copy] "
	if props.IsCut {
		prefix = " [Cut]  "
	}

	content := prefix + path
	return ui.FlexRow(props.Width, ui.SelectableRow(content, props.Width, selected, props.Style))
}

func renderEmptyClipboard(width, height int) string {
	return strings.Repeat("\n", height/2) +
		strings.Repeat(" ", width/2-10) + "Clipboard is empty"
}

// RenderClipboardFooter renders hints for clipboard actions
func RenderClipboardFooter(width int, isEmpty bool, styles theme.Stylesheet) string {
	hint := "[Esc/Alt+c] Back"
	if !isEmpty {
		hint += " | [x] Remove | [v] Paste"
	}
	return styles.Footer.Width(width).Render(" " + messages.ColorizeKeys(messages.Props{Style: styles}, hint))
}
