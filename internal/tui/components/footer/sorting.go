package footer

import (
	"fm/internal/files/sorting"
	"fm/internal/tui/theme"
)

// renderSortMode renders the sort mode indicator
func renderSortMode(sortMode sorting.SortMode, styles theme.Stylesheet) string {
	sortStr := sortMode.String()
	if sortStr == "" {
		return ""
	}

	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()

	return dimStyle.Render("Sort: ") + normalStyle.Render(sortStr)
}
