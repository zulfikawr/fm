package list

import (
	"fmt"
	"strings"

	"fm/internal/files/core"

	"github.com/charmbracelet/lipgloss"
)

// renderRow renders a single row in the file list
func renderRow(props Props, item core.Item, isCursor bool, layout Layout) string {
	marker := renderMarker(props, item)
	gitMarker := renderGitMarker(props, item, isCursor)
	permIndicator := renderPermIndicator(props, item, isCursor)
	namePart := renderNamePart(props, item, isCursor, layout)
	metaPart := renderMetaPart(props, item, isCursor, layout)

	if isCursor {
		sStyle := props.Styles.SelectedItem.UnsetPadding().UnsetWidth()

		row := sStyle.Render(marker) +
			gitMarker +
			sStyle.Render(permIndicator) +
			sStyle.Render(namePart) +
			sStyle.Render(metaPart)

		currentWidth := lipgloss.Width(row)
		if currentWidth < props.Width {
			row += sStyle.Render(strings.Repeat(" ", props.Width-currentWidth))
		}
		return row
	}

	row := fmt.Sprintf("%s%s%s%s%s", marker, gitMarker, permIndicator, namePart, metaPart)
	return props.Styles.Item.Width(props.Width).Render(row)
}
