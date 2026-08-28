package views

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/tui/components/messages"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// LogsProps contains all data needed to render the log list
type LogsProps struct {
	Width  int
	Height int
	Cursor int
	Offset int
	Logs   []tuictx.LogEntry
	Style  theme.Stylesheet
}

// RenderLogs renders the complete log view
func RenderLogs(props LogsProps) string {
	if props.Height <= 0 {
		return ""
	}

	// Add one space after header
	viewHeight := props.Height - 1
	if viewHeight <= 0 {
		return ""
	}

	if len(props.Logs) == 0 {
		return renderLogsEmpty(props.Width, viewHeight, props.Style)
	}

	var rows []string
	rows = append(rows, "")
	// Latest first
	for i := 0; i < viewHeight-1 && i < len(props.Logs); i++ {
		// Use inverse index for latest first
		idx := len(props.Logs) - 1 - i - props.Offset
		if idx < 0 || idx >= len(props.Logs) {
			continue
		}
		entry := props.Logs[idx]

		rows = append(rows, renderLogEntry(props, entry, i == props.Cursor))
	}

	// Ensure we fill the viewport
	for len(rows) < props.Height {
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}

func renderLogEntry(props LogsProps, e tuictx.LogEntry, isCursor bool) string {
	levelColor := props.Style.FileCol.GetForeground()

	switch e.Level {
	case tuictx.LogSuccess:
		levelColor = lipgloss.Color("#50FA7B") // Greenish
	case tuictx.LogWarn:
		levelColor = lipgloss.Color("#F1FA8C") // Yellowish
	case tuictx.LogError:
		levelColor = lipgloss.Color("#FF5555") // Reddish
	case tuictx.LogInfo:
		levelColor = props.Style.DirCol.GetForeground()
	}

	timeStr := e.Timestamp.Format("15:04:05")
	typeText := fmt.Sprintf("[%s]", e.Type)
	const timeWidth = 8
	const typeWidth = 10

	// Base styles for parts
	timeStyle := lipgloss.NewStyle().Foreground(props.Style.DimCol.GetForeground())
	typeStyle := lipgloss.NewStyle().Foreground(levelColor).Bold(true)
	msgStyle := lipgloss.NewStyle().Foreground(props.Style.FileCol.GetForeground())

	if isCursor {
		sStyle := props.Style.SelectedItem.UnsetPadding().UnsetWidth()

		// Use fixed width for each part to maintain alignment
		timePart := timeStyle.Inherit(sStyle).Render(fmt.Sprintf("%-*s", timeWidth, timeStr))
		typePart := typeStyle.Inherit(sStyle).Render(fmt.Sprintf("%-*s", typeWidth, typeText))
		msgPart := msgStyle.Inherit(sStyle).Render(e.Message)
		spacePart := sStyle.Render(" ")

		content := spacePart + timePart + spacePart + typePart + spacePart + msgPart

		// Add details if present and selected
		// Indent = 1 (start) + timeWidth (8) + 1 (space) + typeWidth (10) + 1 (space) + 2 (extra) = 23
		if e.Details != "" {
			indent := strings.Repeat(" ", 23)
			content += "\n" + sStyle.Render(indent) + props.Style.DimCol.Inherit(sStyle).Render(e.Details)
		}

		// Use the full width for the background highlight
		return props.Style.SelectedItem.Width(props.Width).Render(content)
	}

	// For normal rows, use the same alignment as selected rows for consistency
	timePart := timeStyle.Render(fmt.Sprintf("%-*s", timeWidth, timeStr))
	typePart := typeStyle.Render(fmt.Sprintf("%-*s", typeWidth, typeText))
	msgPart := msgStyle.Render(e.Message)

	// Consistent spacing with selected items
	return " " + timePart + " " + typePart + " " + msgPart
}

func renderLogsEmpty(width, height int, styles theme.Stylesheet) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(styles.DimCol.Render("No operations logged yet."))
}

// RenderLogsFooter renders navigation hints for the log view
func RenderLogsFooter(width int, styles theme.Stylesheet) string {
	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	hint := "[Esc/Ctrl+L] Back"
	return styles.Footer.Width(width).Render(" " + messages.ColorizeKeys(messages.Props{Style: styles}, hint) + dimStyle.Render(""))
}
