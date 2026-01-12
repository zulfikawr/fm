package footer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderNormalFooter renders the default footer with status information
func renderNormalFooter(props Props) string {
	baseFooterStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	// Build left side: pagination and hints
	parts := buildNormalFooterParts(props)

	// Assemble left content
	leftContent := assembleFooterContent(parts, props.Width, props.Styles)

	// If we have selected items, add action hints
	if props.SelectedCount > 0 {
		hints := buildActionHints(props)
		if hints != "" {
			spacer := baseFooterStyle.Render("  ")
			leftContent = leftContent + spacer + hints
		}
	}

	// Build right side: sort mode
	rightContent := renderSortMode(props.SortMode, props.Styles)

	// Calculate gap for spacing
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)
	gap := props.Width - leftWidth - rightWidth - 2 // -2 for padding
	if gap < 1 {
		gap = 1
	}

	// Combine with proper spacing
	fullContent := leftContent + baseFooterStyle.Render(strings.Repeat(" ", gap)) + rightContent

	return props.Styles.Footer.Width(props.Width).Render(" " + fullContent)
}

// buildActionHints builds the action hints for selected items
func buildActionHints(props Props) string {
	dimStyle := props.Styles.DimCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	keyStyle := props.Styles.KeyCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	// Add selection count prefix
	prefix := dimStyle.Render("[") + keyStyle.Render(fmt.Sprintf("%d selected", props.SelectedCount)) + dimStyle.Render("]  ")

	hints := []string{
		dimStyle.Render("[") + keyStyle.Render("c") + dimStyle.Render("]") + normalStyle.Render(" Copy"),
		dimStyle.Render("[") + keyStyle.Render("x") + dimStyle.Render("]") + normalStyle.Render(" Cut"),
		dimStyle.Render("[") + keyStyle.Render("r") + dimStyle.Render("]") + normalStyle.Render(" Rename"),
		dimStyle.Render("[") + keyStyle.Render("d") + dimStyle.Render("]") + normalStyle.Render(" Delete"),
	}

	if props.ClipboardCount > 0 {
		hints = append(hints, dimStyle.Render("[")+keyStyle.Render("v")+dimStyle.Render("]")+normalStyle.Render(" Paste"))
	}

	return prefix + strings.Join(hints, dimStyle.Render(" | "))
}
