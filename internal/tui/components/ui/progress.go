package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

// ProgressProps encapsulates data for rendering a progress bar
type ProgressProps struct {
	Label   string
	Percent float64
	Width   int
	Styles  theme.Stylesheet
}

// ProgressBar renders a standardized progress bar.
func ProgressBar(props ProgressProps) string {
	percentInt := int(props.Percent * 100)
	if percentInt > 100 {
		percentInt = 100
	}
	if percentInt < 0 {
		percentInt = 0
	}
	percStr := fmt.Sprintf(" %3d%%", percentInt)

	// Available width for label and bar (account for percStr and leading space)
	availableWidth := props.Width - len(percStr) - 2
	if availableWidth < 10 {
		// Just show label and percent if too narrow
		content := Truncate(props.Label, availableWidth) + percStr
		return props.Styles.Footer.Width(props.Width).Render(" " + content)
	}

	maxLabelWidth := int(float64(availableWidth) * 0.4)
	displayLabel := Truncate(props.Label, maxLabelWidth)

	barWidth := availableWidth - lipgloss.Width(displayLabel) - 3
	if barWidth < 5 {
		barWidth = 5
	}

	filled := int(float64(barWidth) * props.Percent)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	progressStyle := props.Styles.KeyCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	dimStyle := props.Styles.DimCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	baseStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	colorizedBar := dimStyle.Render("[") +
		progressStyle.Render(strings.Repeat("#", filled)) +
		dimStyle.Render(strings.Repeat(".", barWidth-filled)) +
		dimStyle.Render("]")

	content := baseStyle.Render(" "+displayLabel+" ") + colorizedBar + baseStyle.Render(percStr)
	return props.Styles.Footer.Width(props.Width).Render(content)
}
