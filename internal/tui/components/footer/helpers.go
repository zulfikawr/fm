package footer

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// --- Shared Internal Helpers ---

func assembleFooterContent(parts []string, width int, styles theme.Stylesheet) string {
	if len(parts) == 0 {
		return ""
	}

	dimStyle := styles.MutedCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
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
	primaryStyle := styles.PrimaryCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	mutedStyle := styles.MutedCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()

	if info.Total == 0 {
		return primaryStyle.Render("-/0")
	}

	var currentStr string
	if info.Current < 0 {
		currentStr = "-"
	} else {
		currentStr = fmt.Sprintf("%d", info.Current+1)
	}

	return primaryStyle.Render(currentStr) + mutedStyle.Render("/") + primaryStyle.Render(fmt.Sprintf("%d", info.Total))
}

func renderPermissionInfo(items []core.Item, cursor int, styles theme.Stylesheet) string {
	if cursor < 0 || cursor >= len(items) {
		return ""
	}

	item := items[cursor]
	if item.State.IsUp {
		return ""
	}

	secondaryStyle := styles.SecondaryCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	permStr := formatPermissions(uint32(item.Metadata.Mode.Perm()))

	return secondaryStyle.Render(permStr)
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

	infoStyle := styles.InfoCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	return infoStyle.Render(sortStr)
}

func renderRAMUsage(styles theme.Stylesheet) string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ramMB := m.Alloc / 1024 / 1024

	infoStyle := styles.InfoCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	return infoStyle.Render(fmt.Sprintf("RAM: %dMB", ramMB))
}

func buildSelectedIndicator(props Props) string {
	mutedStyle := props.Styles.MutedCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	highlightStyle := props.Styles.HighlightCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()

	return mutedStyle.Render("[") + highlightStyle.Render(fmt.Sprintf("%d selected", props.Status.SelectedCount)) + mutedStyle.Render("]")
}

func buildActionShortcuts(props Props) string {
	mutedStyle := props.Styles.MutedCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	accentStyle := props.Styles.AccentCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
	normalStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	hints := []string{
		mutedStyle.Render("[") + accentStyle.Render("c") + mutedStyle.Render("]") + normalStyle.Render(" Copy"),
		mutedStyle.Render("[") + accentStyle.Render("x") + mutedStyle.Render("]") + normalStyle.Render(" Cut"),
		mutedStyle.Render("[") + accentStyle.Render("z") + mutedStyle.Render("]") + normalStyle.Render(" Zip"),
	}

	// Check if a zip file is focused or selected to show Unzip hint
	showUnzip := false
	if props.Status.SelectedCount > 0 {
		for _, item := range props.Status.Items {
			if item.State.Selected && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
				showUnzip = true
				break
			}
		}
	} else if props.Status.Cursor >= 0 && props.Status.Cursor < len(props.Status.FilteredItems) {
		item := props.Status.FilteredItems[props.Status.Cursor]
		if !item.State.IsUp && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
			showUnzip = true
		}
	}

	if showUnzip {
		hints = append(hints, mutedStyle.Render("[")+accentStyle.Render("u")+mutedStyle.Render("]")+normalStyle.Render(" Unzip"))
	}

	hints = append(hints,
		mutedStyle.Render("[")+accentStyle.Render("r")+mutedStyle.Render("]")+normalStyle.Render(" Rename"),
		mutedStyle.Render("[")+accentStyle.Render("d")+mutedStyle.Render("]")+normalStyle.Render(" Delete"),
	)

	if props.Confirm.ClipboardCount > 0 {
		hints = append(hints, mutedStyle.Render("[")+accentStyle.Render("v")+mutedStyle.Render("]")+normalStyle.Render(" Paste"))
	}

	return strings.Join(hints, mutedStyle.Render(" | "))
}

// GetActionAt returns the action shortcut key at the given x-coordinate
func GetActionAt(x int, props Props) string {
	// Re-calculate exactly like renderStatsFooter
	baseFooterStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

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

	rightContent := renderSortMode(props.Status.SortMode, props.Styles)
	rightWidth := calculateWidth(rightContent)
	availableWidth := props.Width - rightWidth - 2

	partsWidthLimit := availableWidth
	indicator := ""
	if props.Status.SelectedCount > 0 {
		indicator = buildSelectedIndicator(props)
		indicatorWidth := calculateWidth(indicator)
		partsWidthLimit -= (indicatorWidth + 2)
	}

	leftContent := assembleFooterContent(parts, partsWidthLimit, props.Styles)
	leftWidth := calculateWidth(leftContent)

	// Check if shortcuts are rendered
	if props.Status.SelectedCount > 0 {
		shortcuts := buildActionShortcuts(props)
		spacer := baseFooterStyle.Render("  ")
		indicatorWidth := calculateWidth(indicator)
		shortcutsWidth := calculateWidth(shortcuts)
		spacerWidth := calculateWidth(spacer)

		totalLeftWidthWithShortcuts := leftWidth
		if leftWidth > 0 {
			totalLeftWidthWithShortcuts += spacerWidth
		}
		totalLeftWidthWithShortcuts += indicatorWidth + spacerWidth + shortcutsWidth

		if totalLeftWidthWithShortcuts <= availableWidth {
			// Shortcuts ARE rendered. Map them.
			startX := 1 + leftWidth
			if leftWidth > 0 {
				startX += 2 // spacer
			}
			startX += indicatorWidth + 2 // spacer

			// Now startX is where shortcuts start
			// Parse buildActionShortcuts to get exact labels and keys
			type actionInfo struct {
				key   string
				label string
			}
			var actions []actionInfo
			actions = append(actions, actionInfo{"c", "Copy"}, actionInfo{"x", "Cut"}, actionInfo{"z", "Zip"})

			showUnzip := false
			if props.Status.SelectedCount > 0 {
				for _, item := range props.Status.Items {
					if item.State.Selected && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
						showUnzip = true
						break
					}
				}
			} else if props.Status.Cursor >= 0 && props.Status.Cursor < len(props.Status.FilteredItems) {
				item := props.Status.FilteredItems[props.Status.Cursor]
				if !item.State.IsUp && strings.HasSuffix(strings.ToLower(item.Name), ".zip") {
					showUnzip = true
				}
			}

			if showUnzip {
				actions = append(actions, actionInfo{"u", "Unzip"})
			}
			actions = append(actions, actionInfo{"r", "Rename"}, actionInfo{"d", "Delete"})

			if props.Confirm.ClipboardCount > 0 {
				actions = append(actions, actionInfo{"v", "Paste"})
			}

			currentX := startX
			for i, a := range actions {
				fullLen := 3 + 1 + len(a.label) // "[k] Label"
				if x >= currentX && x < currentX+fullLen {
					return a.key
				}
				currentX += fullLen
				if i < len(actions)-1 {
					currentX += 3 // " | "
				}
			}
		}
	}

	// Check if click is on sort mode (Right side)
	if x >= props.Width-rightWidth-1 && x < props.Width-1 {
		return "s"
	}

	return ""
}
