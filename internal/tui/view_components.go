package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) formatBool(b bool) string {
	if b {
		return m.styles.KeyCol.Render("[ON]")
	}
	return m.styles.DimCol.Render("[OFF]")
}

func (m *Model) renderHeader() string {
	if m.settingsOpen {
		return m.styles.Header.Width(m.width).Render("Settings")
	}

	// Render tabs if there are multiple tabs
	var tabsStr string
	if len(m.tabs) > 1 {
		activeTabStyle := m.styles.KeyCol.Inherit(m.styles.Header).UnsetPadding().UnsetWidth()
		inactiveTabStyle := m.styles.DimCol.Inherit(m.styles.Header).UnsetPadding().UnsetWidth()
		spacerStyle := m.styles.Header.UnsetPadding().UnsetWidth()

		var tabParts []string
		for i := range m.tabs {
			tabLabel := fmt.Sprintf("[%d]", i+1)
			if i == m.activeTab {
				// Active tab in primary color
				tabParts = append(tabParts, activeTabStyle.Render(tabLabel))
			} else {
				// Inactive tab in dim color
				tabParts = append(tabParts, inactiveTabStyle.Render(tabLabel))
			}
		}
		tabsStr = strings.Join(tabParts, spacerStyle.Render(" "))
	}

	sep := m.fs.Separator()
	parts := strings.Split(m.path, sep)
	var cleanParts []string
	if m.path == sep {
		cleanParts = []string{sep}
	} else {
		for _, p := range parts {
			if p != "" {
				cleanParts = append(cleanParts, p)
			}
		}
	}

	dimHeaderStyle := m.styles.DimCol.Inherit(m.styles.Header)

	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, m.styles.Header.UnsetPadding().UnsetWidth().Render(p))
	}

	separator := dimHeaderStyle.Render(" > ")
	breadcrumb := strings.Join(styledParts, separator)
	if !strings.HasPrefix(breadcrumb, sep) && strings.HasPrefix(m.path, sep) {
		breadcrumb = dimHeaderStyle.Render(sep+" ") + breadcrumb
	}

	if m.gitBranch != "" {
		breadcrumb += dimHeaderStyle.Render(fmt.Sprintf(" (%s)", m.gitBranch))
	}

	if m.readOnly {
		roStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Inherit(m.styles.Header)
		breadcrumb += roStyle.Render(" [RO]")
	}

	// Combine breadcrumb and tabs (tabs at the right end fixed position)
	baseHeaderStyle := m.styles.Header.UnsetPadding().UnsetWidth()
	fullHeader := breadcrumb
	if len(m.tabs) > 1 {
		breadcrumbWidth := lipgloss.Width(breadcrumb)
		tabsWidth := lipgloss.Width(tabsStr)
		gap := m.width - breadcrumbWidth - tabsWidth - 2 // -2 for padding
		if gap < 1 {
			gap = 1
		}
		fullHeader = breadcrumb + baseHeaderStyle.Render(strings.Repeat(" ", gap)) + tabsStr
	}

	return m.styles.Header.Width(m.width).Render(fullHeader)
}

func (m *Model) renderFooter() string {
	baseFooterStyle := m.styles.Footer.UnsetPadding().UnsetWidth()

	if m.showProgress {
		// Custom responsive progress bar: Label [#######.......] 100%
		percent := int(m.progressPercent * 100)
		if percent > 100 {
			percent = 100
		}
		percStr := fmt.Sprintf(" %3d%%", percent)

		// Calculate available width for the bar itself
		// Format: " Label [###...] 100%"
		// Padding: 1 (start) + 1 (after label) + 1 (before brackets) + 2 (brackets) + 1 (before percent) = 6
		label := m.progressLabel
		availableWidth := m.width - len(percStr) - 6

		if availableWidth < 10 {
			// Extremely narrow, just show label and percent
			content := label + percStr
			if len(content) > m.width-2 {
				content = content[:m.width-5] + "..." + percStr
			}
			return m.styles.Footer.Width(m.width).Render(" " + content)
		}

		// Truncate label if it takes more than 40% of space
		maxLabelWidth := int(float64(availableWidth) * 0.4)
		if len(label) > maxLabelWidth {
			label = label[:maxLabelWidth-3] + "..."
		}

		barWidth := availableWidth - len(label)
		filledWidth := int(float64(barWidth) * m.progressPercent)
		if filledWidth > barWidth {
			filledWidth = barWidth
		}

		dimStyle := m.styles.DimCol.Inherit(m.styles.Footer)
		barStyle := m.styles.ProgressBar.Inherit(m.styles.Footer)

		bar := dimStyle.Render("[")
		bar += barStyle.Render(strings.Repeat("#", filledWidth))
		bar += dimStyle.Render(strings.Repeat(".", barWidth-filledWidth))
		bar += dimStyle.Render("]")

		styledLabel := baseFooterStyle.Render(label)
		styledPerc := baseFooterStyle.Render(percStr)

		content := styledLabel + " " + bar + styledPerc
		return m.styles.Footer.Width(m.width).Render(" " + content)
	}

	if m.searching {
		return m.styles.Footer.Width(m.width).Render(" " + m.searchInput.View())
	}

	if m.renaming {
		return m.styles.Footer.Width(m.width).Render(" " + m.renameInput.View())
	}

	if m.confirming {
		prompt := ""
		switch m.actionType {
		case "delete":
			prompt = "Delete selected items? (y/n)"
		case "paste":
			prompt = fmt.Sprintf("Paste %d items? (y/n)", len(m.clipboard))
		case "reset-settings":
			prompt = "Reset all settings to defaults? (y/n)"
		case "conflict":
			baseName := m.fs.Base(m.conflictDst)
			prompt = fmt.Sprintf("'%s' exists. [y] Overwrite | [n] Skip | [r] Rename", baseName)
		}
		return m.styles.Footer.Width(m.width).Render(" " + m.colorizeKeys(prompt))
	}

	if m.settingsOpen {
		helpMsg := ""
		switch m.settingsCursor {
		case 0:
			helpMsg = "Show/hide files starting with '.'"
		case 1:
			helpMsg = "Search respects capitalization"
		case 2:
			helpMsg = "Ask before destructive actions"
		case 3:
			helpMsg = "Cursor loops at list boundaries"
		case 4:
			helpMsg = "Choose default editor for opening files"
		case 5:
			helpMsg = "Move deleted items to trash (off = permanent delete)"
		case 6:
			helpMsg = "Show/hide list column headers"
		case 7:
			helpMsg = "Enable git status markers"
		case 8:
			helpMsg = "Show file size in list"
		case 9:
			helpMsg = "Change the file size display unit"
		case 10:
			helpMsg = "Show last modification time"
		case 11:
			helpMsg = "Change the date and time format"
		case 12:
			helpMsg = "Change the application color scheme"
		}

		leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [r] Reset | [Esc/.] Back"
		rightPart := helpMsg + " "

		gap := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
		if gap < 0 {
			gap = 0
		}

		footerContent := m.colorizeKeys(leftPart) + baseFooterStyle.Render(strings.Repeat(" ", gap)) + baseFooterStyle.Render(rightPart)
		return m.styles.Footer.Width(m.width).Render(footerContent)
	}

	// If there's a message (notification/error/confirmation), show only the message
	if m.msg != "" {
		return m.styles.Footer.Width(m.width).Render(" " + m.msg)
	}

	selectedCount := 0
	for _, item := range m.items {
		if item.Selected {
			selectedCount++
		}
	}

	statusMsg := ""
	if selectedCount > 0 {
		statusMsg = "[c] Copy | [x] Cut | [r] Rename | [d] Delete"
		if len(m.clipboard) > 0 {
			statusMsg += " | [v] Paste"
		}
	}

	if selectedCount > 0 {
		statusMsg = fmt.Sprintf("[%d selected] %s", selectedCount, statusMsg)
	}

	sortStr := m.sortMode.String()

	dimFooterStyle := m.styles.DimCol.Inherit(m.styles.Footer)

	// Calculate correct item count and current index excluding the "up dir"
	totalItems := len(m.filteredItems)
	currentIndex := m.cursor
	hasUp := len(m.filteredItems) > 0 && m.filteredItems[0].IsUp

	if hasUp {
		totalItems--
		currentIndex--
	}

	paginationStr := ""
	if totalItems > 0 {
		if currentIndex < 0 {
			paginationStr = fmt.Sprintf(" -/%d  ", totalItems)
		} else {
			paginationStr = fmt.Sprintf(" %d/%d  ", currentIndex+1, totalItems)
		}
	} else {
		paginationStr = " 0/0  "
	}

	pagination := baseFooterStyle.Render(paginationStr)
	keys := m.colorizeKeys(statusMsg)
	sorting := dimFooterStyle.Render(sortStr) + baseFooterStyle.Render(" ")

	leftWidth := lipgloss.Width(pagination) + lipgloss.Width(keys)
	rightWidth := lipgloss.Width(sorting)

	gap := m.width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	footerContent := pagination + keys + baseFooterStyle.Render(strings.Repeat(" ", gap)) + sorting
	return m.styles.Footer.Width(m.width).Render(footerContent)
}

func (m *Model) colorizeKeys(s string) string {
	var result strings.Builder
	inBracket := false
	keyStyle := m.styles.KeyCol.Inherit(m.styles.Footer)
	baseStyle := m.styles.Footer.UnsetPadding().UnsetWidth()

	var current strings.Builder
	for _, r := range s {
		switch r {
		case '[':
			if current.Len() > 0 {
				result.WriteString(baseStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = true
			result.WriteString(keyStyle.Render("["))
		case ']':
			if current.Len() > 0 {
				result.WriteString(keyStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = false
			result.WriteString(keyStyle.Render("]"))
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		if inBracket {
			result.WriteString(keyStyle.Render(current.String()))
		} else {
			result.WriteString(baseStyle.Render(current.String()))
		}
	}
	return result.String()
}

func (m *Model) getViewportHeight() int {
	headerH := lipgloss.Height(m.renderHeader())
	footerH := lipgloss.Height(m.renderFooter())
	h := m.height - headerH - footerH
	if m.cfg.ShowHeader && !m.settingsOpen {
		h -= 3 // 3 for header lines (separator + text + separator)
	}
	if h < 1 {
		return 1
	}
	return h
}
