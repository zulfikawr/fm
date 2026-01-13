package header

import (
	"fmt"

	"fm/internal/tui/theme"
)

// RenderRemote renders the remote connection indicator
func RenderRemote(connected bool, user, host string, styles theme.Stylesheet) string {
	// If no host is known, show nothing
	if host == "" {
		return ""
	}

	label := fmt.Sprintf("%s@%s", user, host)
	if user == "" {
		label = host
	}

	// Use a style that blends with the header (same bg, subtle fg)
	// We use DimCol but ensure we inherit Header background/unsetting width/padding
	// similar to how tabs are rendered.
	style := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	if connected {
		return style.Render(label)
	}

	// Disconnected
	return style.Render(fmt.Sprintf("%s (Disconnected)", label))
}
