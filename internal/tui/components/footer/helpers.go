package footer

import (
	"fmt"
	"strings"

	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// --- Shared Internal Helpers ---

func assembleFooterContent(parts []string, width int, styles theme.Stylesheet) string {
	if len(parts) == 0 {
		return ""
	}

	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	spacer := dimStyle.Render(" | ")
	content := strings.Join(parts, spacer)

	if lipgloss.Width(content) > width-2 {
		maxWidth := width - 5
		if maxWidth < 0 {
			maxWidth = 0
		}
		if len(content) > maxWidth {
			content = content[:maxWidth] + "..."
		}
	}

	return content
}

func calculateWidth(str string) int {
	return lipgloss.Width(str)
}

// PaginationInfo holds pagination display information
type PaginationInfo struct {
	Current int
	Total   int
	Width   int
}

func renderPaginationInfo(info PaginationInfo, styles theme.Stylesheet) string {
	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()

	if info.Total == 0 {
		return normalStyle.Render("-/0")
	}

	var currentStr string
	if info.Current < 0 {
		currentStr = "-"
	} else {
		currentStr = fmt.Sprintf("%d", info.Current+1)
	}

	return normalStyle.Render(currentStr) + dimStyle.Render("/") + normalStyle.Render(fmt.Sprintf("%d", info.Total))
}

func renderPermissionInfo(items []core.Item, cursor int, styles theme.Stylesheet) string {
	if cursor < 0 || cursor >= len(items) {
		return ""
	}

	item := items[cursor]
	if item.IsUp {
		return ""
	}

	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()
	permStr := formatPermissions(uint32(item.Mode.Perm()))

	return normalStyle.Render(permStr)
}

func formatPermissions(perm uint32) string {
	r, w, x := "-", "-", "-"
	if perm&0o400 != 0 {
		r = "r"
	}
	if perm&0o200 != 0 {
		w = "w"
	}
	if perm&0o100 != 0 {
		x = "x"
	}

	gr, gw, gx := "-", "-", "-"
	if perm&0o040 != 0 {
		gr = "r"
	}
	if perm&0o020 != 0 {
		gw = "w"
	}
	if perm&0o010 != 0 {
		gx = "x"
	}

	or, ow, ox := "-", "-", "-"
	if perm&0o004 != 0 {
		or = "r"
	}
	if perm&0o002 != 0 {
		ow = "w"
	}
	if perm&0o001 != 0 {
		ox = "x"
	}

	return r + w + x + gr + gw + gx + or + ow + ox
}

func renderSortMode(sortMode sorting.SortMode, styles theme.Stylesheet) string {
	sortStr := sortMode.String()
	if sortStr == "" {
		return ""
	}

	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := styles.Footer.UnsetPadding().UnsetWidth()

	return dimStyle.Render("Sort: ") + normalStyle.Render(sortStr)
}

func buildActionHints(props Props) string {
	dimStyle := props.Styles.DimCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	keyStyle := props.Styles.KeyCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	prefix := dimStyle.Render("[") + keyStyle.Render(fmt.Sprintf("%d selected", props.SelectedCount)) + dimStyle.Render("]  ")

	hints := []string{
		dimStyle.Render("[") + keyStyle.Render("c") + dimStyle.Render("]") + normalStyle.Render(" Copy"),
		dimStyle.Render("[") + keyStyle.Render("x") + dimStyle.Render("]") + normalStyle.Render(" Cut"),
		dimStyle.Render("[") + keyStyle.Render("z") + dimStyle.Render("]") + normalStyle.Render(" Zip"),
	}

	// Check if a zip file is focused or selected to show Unzip hint
	showUnzip := false
	if props.SelectedCount > 0 {
		for _, item := range props.Items {
			if item.Selected && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
				showUnzip = true
				break
			}
		}
	} else if props.Cursor >= 0 && props.Cursor < len(props.FilteredItems) {
		item := props.FilteredItems[props.Cursor]
		if !item.IsUp && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
			showUnzip = true
		}
	}

	if showUnzip {
		hints = append(hints, dimStyle.Render("[")+keyStyle.Render("u")+dimStyle.Render("]")+normalStyle.Render(" Unzip"))
	}

	hints = append(hints,
		dimStyle.Render("[")+keyStyle.Render("r")+dimStyle.Render("]")+normalStyle.Render(" Rename"),
		dimStyle.Render("[")+keyStyle.Render("d")+dimStyle.Render("]")+normalStyle.Render(" Delete"),
	)

	if props.ClipboardCount > 0 {
		hints = append(hints, dimStyle.Render("[")+keyStyle.Render("v")+dimStyle.Render("]")+normalStyle.Render(" Paste"))
	}

	return prefix + strings.Join(hints, dimStyle.Render(" | "))
}
