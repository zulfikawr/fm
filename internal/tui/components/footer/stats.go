package footer

import (
	"strings"
)

// renderStatsFooter renders the normal status line with pagination and permissions
func renderStatsFooter(props Props) string {
	baseFooterStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	// Adjust for "↑ .."
	total := props.Status.TotalItems
	current := props.Status.Cursor
	if len(props.Status.FilteredItems) > 0 && props.Status.FilteredItems[0].State.IsUp {
		total--
		if current == 0 {
			current = -1
		} else {
			current--
		}
	}

	var parts []string
	pagination := renderPaginationInfo(PaginationInfo{
		Current: current,
		Total:   total,
		Width:   props.Width,
	}, props.Styles)
	if pagination != "" {
		parts = append(parts, pagination)
	}

	permission := renderPermissionInfo(props.Status.FilteredItems, props.Status.Cursor, props.Styles)
	if permission != "" {
		parts = append(parts, permission)
	}

	var rightParts []string
	sortMode := renderSortMode(props.Status.SortMode, props.Styles)
	if sortMode != "" {
		rightParts = append(rightParts, sortMode)
	}
	if props.Status.ShowRAM {
		ramUsage := renderRAMUsage(props.Styles)
		if ramUsage != "" {
			rightParts = append(rightParts, ramUsage)
		}
	}

	rightContent := ""
	if len(rightParts) > 0 {
		dimStyle := props.Styles.MutedCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
		separator := dimStyle.Render(" | ")
		rightContent = baseFooterStyle.Render(strings.Join(rightParts, separator))
	}
	rightWidth := calculateWidth(rightContent)

	// Calculate available width for the left side (all except rightContent and padding)
	availableWidth := props.Width - rightWidth - 2
	if availableWidth < 0 {
		availableWidth = 0
	}

	indicator := ""
	if props.Status.SelectedCount > 0 {
		indicator = buildSelectedIndicator(props)
	}

	// Calculate how much width to give to pagination/permissions
	partsWidthLimit := availableWidth
	if indicator != "" {
		indicatorWidth := calculateWidth(indicator)
		partsWidthLimit -= (indicatorWidth + 2) // 2 for spacer
	}
	if partsWidthLimit < 0 {
		partsWidthLimit = 0
	}

	leftContent := assembleFooterContent(parts, partsWidthLimit, props.Styles)

	if props.Status.SelectedCount > 0 {
		shortcuts := buildActionShortcuts(props)
		spacer := baseFooterStyle.Render("  ")

		leftWidth := calculateWidth(leftContent)
		indicatorWidth := calculateWidth(indicator)
		shortcutsWidth := calculateWidth(shortcuts)
		spacerWidth := calculateWidth(spacer)

		// Check if everything fits including shortcuts
		totalLeftWidthWithShortcuts := leftWidth
		if leftWidth > 0 {
			totalLeftWidthWithShortcuts += spacerWidth
		}
		totalLeftWidthWithShortcuts += indicatorWidth + spacerWidth + shortcutsWidth

		if totalLeftWidthWithShortcuts <= availableWidth {
			if leftContent != "" {
				leftContent = leftContent + spacer + indicator + spacer + shortcuts
			} else {
				leftContent = indicator + spacer + shortcuts
			}
		} else {
			if leftContent != "" {
				leftContent = leftContent + spacer + indicator
			} else {
				leftContent = indicator
			}
		}
	}

	fullContent := leftContent
	leftWidth := calculateWidth(leftContent)
	gap := props.Width - leftWidth - rightWidth - 2
	if gap > 0 {
		fullContent += baseFooterStyle.Render(strings.Repeat(" ", gap))
	}
	fullContent += rightContent

	return props.Styles.Footer.Width(props.Width).Render(" " + fullContent)
}
