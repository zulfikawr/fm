package footer

import (
	"fmt"
	"strings"

	"fm/internal/tui/theme"
)

// renderProgress renders a responsive progress bar
func renderProgress(props Props) string {
	return renderProgressBar(props.ProgressLabel, props.ProgressPercent, props.Width, props.Styles)
}

// renderProgressBar renders a custom progress bar with label and percentage
func renderProgressBar(label string, percent float64, width int, styles theme.Stylesheet) string {
	// Calculate percentage
	percentInt := int(percent * 100)
	if percentInt > 100 {
		percentInt = 100
	}
	if percentInt < 0 {
		percentInt = 0
	}
	percStr := fmt.Sprintf(" %3d%%", percentInt)

	// Calculate available width for the bar itself
	// Format: " Label [###...] 100%"
	availableWidth := width - len(percStr) - 6

	if availableWidth < 10 {
		// Extremely narrow, just show label and percent
		content := label + percStr
		if len(content) > width-2 {
			maxLen := width - 5
			if maxLen < 0 {
				maxLen = 0
			}
			content = content[:maxLen] + "..." + percStr
		}
		return styles.Footer.Width(width).Render(" " + content)
	}

	// Truncate label if it takes more than 40% of space
	maxLabelWidth := int(float64(availableWidth) * 0.4)
	if len(label) > maxLabelWidth && maxLabelWidth > 3 {
		label = label[:maxLabelWidth-3] + "..."
	}

	// Recalculate bar width after potential label truncation
	barWidth := availableWidth - len(label) - 1 // -1 for space after label
	if barWidth < 5 {
		barWidth = 5
	}

	// Build the progress bar
	filled := int(float64(barWidth) * percent)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	// Colorize the bar
	progressStyle := styles.KeyCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()

	colorizedBar := dimStyle.Render("[") +
		progressStyle.Render(strings.Repeat("#", filled)) +
		dimStyle.Render(strings.Repeat(".", barWidth-filled)) +
		dimStyle.Render("]")

	content := label + " " + colorizedBar + percStr
	return styles.Footer.Width(width).Render(" " + content)
}
