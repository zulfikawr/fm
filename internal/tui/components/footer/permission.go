package footer

import (
	"fm/internal/files"
	"fm/internal/tui/theme"
)

// renderPermissionInfo renders the permission information (e.g., "rwx")
func renderPermissionInfo(items []files.Item, cursor int, styles theme.Stylesheet) string {
	// Check if cursor is valid
	if cursor < 0 || cursor >= len(items) {
		return ""
	}

	item := items[cursor]

	// Don't show permissions for "Up" item
	if item.IsUp {
		return ""
	}

	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()

	// Get permission string from os.FileMode
	mode := item.Mode
	perm := mode.Perm()

	// Format as rwxrwxrwx
	permStr := formatPermissions(uint32(perm))

	return normalStyle.Render(permStr)
}

// formatPermissions formats os.FileMode permissions to rwx format
func formatPermissions(perm uint32) string {
	// User permissions
	r := "-"
	w := "-"
	x := "-"
	if perm&0400 != 0 {
		r = "r"
	}
	if perm&0200 != 0 {
		w = "w"
	}
	if perm&0100 != 0 {
		x = "x"
	}

	// Group permissions
	gr := "-"
	gw := "-"
	gx := "-"
	if perm&0040 != 0 {
		gr = "r"
	}
	if perm&0020 != 0 {
		gw = "w"
	}
	if perm&0010 != 0 {
		gx = "x"
	}

	// Other permissions
	or := "-"
	ow := "-"
	ox := "-"
	if perm&0004 != 0 {
		or = "r"
	}
	if perm&0002 != 0 {
		ow = "w"
	}
	if perm&0001 != 0 {
		ox = "x"
	}

	return r + w + x + gr + gw + gx + or + ow + ox
}
