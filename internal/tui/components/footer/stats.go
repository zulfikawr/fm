package footer

import (
	"strings"
)

// renderStatsFooter renders the normal status line with pagination and permissions
func renderStatsFooter(props Props) string {
	baseFooterStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	var parts []string
	pagination := renderPaginationInfo(PaginationInfo{
		Current: props.Cursor,
		Total:   props.TotalItems,
		Width:   props.Width,
	}, props.Styles)
	if pagination != "" {
		parts = append(parts, pagination)
	}

	permission := renderPermissionInfo(props.FilteredItems, props.Cursor, props.Styles)
	if permission != "" {
		parts = append(parts, permission)
	}

	leftContent := assembleFooterContent(parts, props.Width, props.Styles)

	if props.SelectedCount > 0 {
		hints := buildActionHints(props)
		if hints != "" {
			spacer := baseFooterStyle.Render("  ")
			leftContent = leftContent + spacer + hints
		}
	}

	rightContent := renderSortMode(props.SortMode, props.Styles)

	fullContent := leftContent

	leftWidth := calculateWidth(leftContent)
	rightWidth := calculateWidth(rightContent)
	gap := props.Width - leftWidth - rightWidth - 2
	if gap > 0 {
		fullContent += baseFooterStyle.Render(strings.Repeat(" ", gap))
	}
	fullContent += rightContent

	return props.Styles.Footer.Width(props.Width).Render(" " + fullContent)
}
