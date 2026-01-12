package header

import (
	"fmt"
	"strings"

	"fm/internal/tui/theme"
)

// TabConfig holds configuration for rendering tabs
type TabConfig struct {
	TabCount     int
	ActiveIndex  int
	ShowShortcut bool
}

// renderTabList renders the list of tab indicators
func renderTabList(config TabConfig, styles theme.Stylesheet) string {
	if config.TabCount == 0 {
		return ""
	}

	activeTabStyle := styles.KeyCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	inactiveTabStyle := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	spacerStyle := styles.Header.UnsetPadding().UnsetWidth()

	var tabParts []string
	for i := 0; i < config.TabCount; i++ {
		tabLabel := formatTabLabel(i, config.ShowShortcut)

		if i == config.ActiveIndex {
			tabParts = append(tabParts, activeTabStyle.Render(tabLabel))
		} else {
			tabParts = append(tabParts, inactiveTabStyle.Render(tabLabel))
		}
	}

	return strings.Join(tabParts, spacerStyle.Render(" "))
}

// formatTabLabel formats a tab label with optional keyboard shortcut
func formatTabLabel(index int, showShortcut bool) string {
	if showShortcut && index < 9 {
		return fmt.Sprintf("[%d]", index+1)
	}
	return fmt.Sprintf("[%d]", index+1)
}

// calculateTabWidth calculates the total width needed for tabs
func calculateTabWidth(tabCount int) int {
	if tabCount == 0 {
		return 0
	}

	// Each tab is "[N]" = 3 characters
	// Plus space between tabs = tabCount - 1
	tabWidth := tabCount * 3
	spacerWidth := tabCount - 1
	return tabWidth + spacerWidth
}

// shouldShowTabs determines if tabs should be shown based on count
func shouldShowTabs(tabCount int) bool {
	return tabCount > 1
}
