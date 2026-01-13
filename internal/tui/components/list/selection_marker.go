package list

import "fm/internal/files/core"

// renderMarker renders the selection indicator ([ ], [x])
func renderMarker(props Props, item core.Item) string {
	if !props.SelectMode {
		return ""
	}
	if item.IsUp {
		return "    "
	}
	if item.Selected {
		return "[x] "
	}
	return "[ ] "
}
