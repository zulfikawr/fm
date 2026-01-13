package list

import "fm/internal/files/core"

// renderPermIndicator renders the permission indicator (!)
func renderPermIndicator(props Props, item core.Item, isCursor bool) string {
	indicator := " "
	if !item.CanWrite && !item.IsUp && !item.IsGhost {
		indicator = "!"
	}

	if !isCursor && indicator == "!" {
		return props.Styles.DimCol.Render("!")
	}
	return indicator
}
