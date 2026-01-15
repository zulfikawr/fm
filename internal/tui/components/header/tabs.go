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

func formatTabLabel(index int, showShortcut bool) string {
	return fmt.Sprintf("[%d]", index+1)
}

func shouldShowTabs(tabCount int) bool {
	return tabCount > 1
}
