package footer

import (
	"fmt"

	"fm/internal/tui/theme"
)

// PaginationInfo holds pagination display information
type PaginationInfo struct {
	Current int
	Total   int
	Width   int
}

// renderPaginationInfo renders the pagination information (e.g., "3/10")
func renderPaginationInfo(info PaginationInfo, styles theme.Stylesheet) string {
	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()

	// Always show pagination
	if info.Total == 0 {
		return normalStyle.Render("-/0")
	}

	// Display cursor position (1-based)
	var currentStr string
	if info.Current < 0 {
		// On "Up" item, show "-"
		currentStr = "-"
	} else {
		currentStr = fmt.Sprintf("%d", info.Current+1)
	}

	// Style both numbers the same to avoid black background
	pagination := normalStyle.Render(currentStr) + dimStyle.Render("/") + normalStyle.Render(fmt.Sprintf("%d", info.Total))

	return pagination
}
