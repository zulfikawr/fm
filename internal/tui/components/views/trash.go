package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/files/trash"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// TrashProps contains data for rendering the trash view
type TrashProps struct {
	Width  int
	Height int
	Cursor int
	Offset int
	Items  []trash.TrashItem
	Style  theme.Stylesheet
}

// RenderTrash renders the trash view
func RenderTrash(props TrashProps) string {
	if props.Height <= 0 {
		return ""
	}

	rows := renderTrashRows(props)

	// Apply scroll offset
	if props.Offset > 0 && props.Offset < len(rows) {
		rows = rows[props.Offset:]
	} else if props.Offset >= len(rows) {
		rows = []string{}
	}

	// Ensure we fill the viewport
	if len(rows) > props.Height {
		rows = rows[:props.Height]
	} else {
		for i := len(rows); i < props.Height; i++ {
			rows = append(rows, "")
		}
	}

	return strings.Join(rows, "\n")
}

func renderTrashRows(props TrashProps) []string {
	var rows []string

	// Header with trash info
	totalSize := int64(0)
	for i := range props.Items {
		totalSize += props.Items[i].SizeBytes
	}

	header := fmt.Sprintf("Trash (%d items, %s)",
		len(props.Items),
		format.FormatSize(totalSize, 1))

	rows = append(rows, props.Style.SettingsHeader.Width(props.Width).Render(header))
	rows = append(rows, "") // Spacer

	if len(props.Items) == 0 {
		// Center the empty message vertically and horizontally
		emptyMsg := "Trash is empty"
		verticalPadding := (props.Height - 3) / 2 // Subtract header and spacer
		for i := 0; i < verticalPadding; i++ {
			rows = append(rows, "")
		}
		rows = append(rows, strings.Repeat(" ", props.Width/2-len(emptyMsg)/2)+emptyMsg)
		return rows
	}

	// Column headers
	nameWidth := props.Width - 50
	if nameWidth < 20 {
		nameWidth = 20
	}

	headerRow := renderTrashHeader(props, nameWidth)
	rows = append(rows, headerRow)
	rows = append(rows, "") // Spacer

	// Items
	for i, item := range props.Items {
		rows = append(rows, renderTrashItem(props, item, i == props.Cursor, nameWidth))
	}

	return rows
}

func renderTrashHeader(props TrashProps, nameWidth int) string {
	nameStyle := props.Style.SecondaryCol.UnsetPadding().UnsetWidth().Bold(true)
	pathStyle := props.Style.SecondaryCol.UnsetPadding().UnsetWidth().Bold(true)
	timeStyle := props.Style.SecondaryCol.UnsetPadding().UnsetWidth().Bold(true)

	name := nameStyle.Width(nameWidth).PaddingLeft(3).Render("Name")
	path := pathStyle.Width(25).Render("Original Path")
	deleted := timeStyle.Width(15).Render("Deleted")

	return lipgloss.JoinHorizontal(lipgloss.Top, name, path, deleted)
}

func renderTrashItem(props TrashProps, item trash.TrashItem, isCursor bool, nameWidth int) string {
	nameStyle := props.Style.PrimaryCol.UnsetPadding().UnsetWidth()
	pathStyle := props.Style.MutedCol.UnsetPadding().UnsetWidth()
	timeStyle := props.Style.SecondaryCol.UnsetPadding().UnsetWidth()
	rowStyle := props.Style.Item

	if isCursor {
		rowStyle = props.Style.SelectedItem.UnsetPadding().UnsetWidth()
		nameStyle = nameStyle.Inherit(rowStyle)
		pathStyle = pathStyle.Inherit(rowStyle)
		timeStyle = timeStyle.Inherit(rowStyle)
	}

	// Name
	displayName := item.OriginalPath
	if len(displayName) > len(item.TrashedName) {
		// Extract just the filename
		parts := strings.Split(item.OriginalPath, "/")
		displayName = parts[len(parts)-1]
	}
	if item.IsDirectory {
		displayName += "/"
	}

	nameContent := nameStyle.Render(displayName)
	namePart := rowStyle.PaddingLeft(3).Width(nameWidth).Render(nameContent)

	// Original path (truncated)
	pathDisplay := item.OriginalPath
	if len(pathDisplay) > 23 {
		pathDisplay = "..." + pathDisplay[len(pathDisplay)-20:]
	}
	pathContent := pathStyle.Render(pathDisplay)
	pathPart := rowStyle.Width(25).Render(pathContent)

	// Time ago
	timeAgo := formatTimeAgo(item.DeletionTime)
	timeContent := timeStyle.Render(timeAgo)
	timePart := rowStyle.Width(15).Render(timeContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, namePart, pathPart, timePart)
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// RenderTrashFooter renders hints for the trash view
func RenderTrashFooter(width int, isEmpty bool, styles theme.Stylesheet) string {
	hint := "[t/Esc] Back"
	if !isEmpty {
		hint = "[r] Restore | [d] Delete | [e] Empty | " + hint
	}
	return styles.Footer.Width(width).Render(" " + messages.ColorizeKeys(messages.Props{Style: styles}, hint))
}
