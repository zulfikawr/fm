package ui

import (
	"fm/internal/tui/theme"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// ProgressBar renders a standardized progress bar.
func ProgressBar(label string, percent float64, width int, styles theme.Stylesheet) string {
	percentInt := int(percent * 100)
	if percentInt > 100 {
		percentInt = 100
	}
	if percentInt < 0 {
		percentInt = 0
	}
	percStr := fmt.Sprintf(" %3d%%", percentInt)

	// Available width for label and bar (account for percStr and leading space)
	availableWidth := width - len(percStr) - 2
	if availableWidth < 10 {
		// Just show label and percent if too narrow
		content := Truncate(label, availableWidth) + percStr
		return styles.Footer.Width(width).Render(" " + content)
	}

	maxLabelWidth := int(float64(availableWidth) * 0.4)
	displayLabel := Truncate(label, maxLabelWidth)

	barWidth := availableWidth - lipgloss.Width(displayLabel) - 3
	if barWidth < 5 {
		barWidth = 5
	}

	filled := int(float64(barWidth) * percent)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	progressStyle := styles.KeyCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	baseStyle := styles.Footer.UnsetPadding().UnsetWidth()

	colorizedBar := dimStyle.Render("[") +
		progressStyle.Render(strings.Repeat("#", filled)) +
		dimStyle.Render(strings.Repeat(".", barWidth-filled)) +
		dimStyle.Render("]")

	content := baseStyle.Render(" "+displayLabel+" ") + colorizedBar + baseStyle.Render(percStr)
	return styles.Footer.Width(width).Render(content)
}
