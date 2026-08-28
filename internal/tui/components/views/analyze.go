package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

type AnalyzeProps struct {
	Width           int
	Height          int
	ActiveNode      *core.AnalysisResult
	Cursor          int
	Offset          int
	Style           theme.Stylesheet
	EnableIcons     bool
	SizeFormatIndex int
	IsRoot          bool
}

func RenderAnalyze(props AnalyzeProps) string {
	if props.ActiveNode == nil {
		return lipgloss.Place(props.Width, props.Height, lipgloss.Center, lipgloss.Center, "No analysis data available.")
	}

	var s strings.Builder

	availableHeight := props.Height
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Prepare list with synthetic "up" item
	var displayItems []*core.AnalysisResult
	if !props.IsRoot {
		upNode := &core.AnalysisResult{
			Name:        "↑ ..",
			IsDirectory: true,
			Percentage:  0,
			Size:        0,
		}
		displayItems = append(displayItems, upNode)
	}
	displayItems = append(displayItems, props.ActiveNode.Children...)

	if len(displayItems) == 0 {
		s.WriteString(lipgloss.Place(props.Width, availableHeight, lipgloss.Center, lipgloss.Center, "No sub-items found."))
		return s.String()
	}

	// Pagination
	end := props.Offset + availableHeight
	if end > len(displayItems) {
		end = len(displayItems)
	}

	const (
		sizeWidth = 10
		percWidth = 7
		iconWidth = 3
	)

	for i := props.Offset; i < end; i++ {
		child := displayItems[i]
		isCursor := i == props.Cursor

		lineStyle := props.Style.Item
		if isCursor {
			lineStyle = props.Style.SelectedItem
		}

		// Icon logic
		iconPart := ""
		if props.EnableIcons {
			icon := ""
			// For "↑ ..", we don't show an extra Nerd Font icon, just the text
			if child.Name != "↑ .." {
				icon = theme.GetIcon(core.Item{IsDir: child.IsDirectory, Name: child.Name, State: core.ItemState{}})
			}

			if icon != "" {
				iconPart = icon + "  "
			} else {
				iconPart = "   "
			}
		}

		// Metadata logic
		var sizeStr, percStr string
		if child.Name == "↑ .." {
			sizeStr = ""
			percStr = ""
		} else {
			sizeStr = format.FormatSize(child.Size, props.SizeFormatIndex)
			percStr = fmt.Sprintf("%5.1f%%", child.Percentage*100)
		}

		nameWidth := props.Width / 4
		if nameWidth < 15 {
			nameWidth = 15
		}

		nameStr := child.Name
		if child.Name != "↑ .." && child.IsDirectory {
			nameStr = child.Name + "/"
		}
		nameStr = ui.Truncate(nameStr, nameWidth)

		occupiedWidth := iconWidth + nameWidth + sizeWidth + percWidth + 6
		barWidth := props.Width - occupiedWidth
		if barWidth < 5 {
			barWidth = 5
		}

		var bar string
		if child.Name == "↑ .." {
			bar = lineStyle.Width(barWidth).Render("")
		} else {
			bar = renderMiniBar(child.Percentage, barWidth, props.Style, isCursor)
		}

		// Colors
		nameStyle := props.Style.FileCol
		if child.Name == "↑ .." {
			nameStyle = props.Style.DimCol
		} else if child.IsDirectory {
			nameStyle = props.Style.DirCol
		}

		metaSizeStyle := props.Style.AccentCol
		metaPercStyle := props.Style.MutedCol

		bgStyle := props.Style.Item
		if isCursor {
			bgStyle = props.Style.SelectedItem.UnsetPadding().UnsetWidth()
		}

		prefixPart := nameStyle.Inherit(bgStyle).Render(iconPart)
		namePart := nameStyle.Inherit(bgStyle).Width(nameWidth).Render(nameStr)
		sizePart := metaSizeStyle.Inherit(bgStyle).Width(sizeWidth).Align(lipgloss.Right).Render(sizeStr)
		percPart := metaPercStyle.Inherit(bgStyle).Width(percWidth).Align(lipgloss.Right).Render(percStr)
		gapPart := bgStyle.Render(" ")

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			gapPart, prefixPart, namePart, gapPart, bar, gapPart, sizePart, gapPart, percPart,
		)
		s.WriteString(lineStyle.Width(props.Width).Render(line) + "\n")
	}

	remaining := availableHeight - (end - props.Offset)
	for i := 0; i < remaining; i++ {
		s.WriteString("\n")
	}

	return s.String()
}

func renderMiniBar(percent float64, width int, style theme.Stylesheet, isSelected bool) string {
	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bgStyle := style.Item
	if isSelected {
		bgStyle = style.SelectedItem
	}

	filledStyle := style.PrimaryCol.Inherit(bgStyle).UnsetPadding().UnsetWidth()
	emptyStyle := style.MutedCol.Inherit(bgStyle).UnsetPadding().UnsetWidth()
	bracketStyle := style.MutedCol.Inherit(bgStyle).UnsetPadding().UnsetWidth()

	bar := filledStyle.Render(strings.Repeat("#", filled)) +
		emptyStyle.Render(strings.Repeat(".", width-filled))

	return bracketStyle.Render("[") + bar + bracketStyle.Render("]")
}

func RenderAnalyzeFooter(width int, style theme.Stylesheet) string {
	props := messages.Props{Style: style}
	prompt := " [Esc/Ctrl+U] Back | [d] Delete"
	return style.Footer.Width(width).Render(messages.ColorizeKeys(props, prompt))
}
