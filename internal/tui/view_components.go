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

	return m.styles.Header.Width(m.width).Render(breadcrumb)
}

func (m *Model) renderFooter() string {
	if m.searching {
		return m.styles.Footer.Width(m.width).Render(" " + m.searchInput.View())
	}

	if m.renaming {
		return m.styles.Footer.Width(m.width).Render(" " + m.renameInput.View())
	}

	if m.confirming {
		prompt := ""
		if m.actionType == "delete" {
			prompt = "Delete selected items? (y/n)"
		} else if m.actionType == "paste" {
			prompt = fmt.Sprintf("Paste %d items? (y/n)", len(m.clipboard))
		}
		return m.styles.Footer.Width(m.width).Render(" " + m.colorizeKeys(prompt))
	}

	baseFooterStyle := m.styles.Footer.UnsetPadding().UnsetWidth()

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
			helpMsg = "Show/hide list column headers"
		case 5:
			helpMsg = "Enable git status markers"
		case 6:
			helpMsg = "Show file size in list"
		case 7:
			helpMsg = "Change the file size display unit"
		case 8:
			helpMsg = "Show last modification time"
		case 9:
			helpMsg = "Change the date and time format"
		case 10:
			helpMsg = "Change the application color scheme"
		}

		leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [Esc/.] Back"
		rightPart := helpMsg + " "

		gap := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
		if gap < 0 {
			gap = 0
		}

		footerContent := m.colorizeKeys(leftPart) + baseFooterStyle.Render(strings.Repeat(" ", gap)) + baseFooterStyle.Render(rightPart)
		return m.styles.Footer.Width(m.width).Render(footerContent)
	}

	selectedCount := 0
	for _, item := range m.items {
		if item.Selected {
			selectedCount++
		}
	}

	statusMsg := m.msg
	if statusMsg == "" {
		if selectedCount > 0 {
			statusMsg = "[c] Copy | [r] Rename | [d] Delete"
		} else {
			statusMsg = "[↑↓] Nav | [Space] Sel | [s] Sort | [/] Search | [.] Settings"
		}
	}

	if selectedCount > 0 {
		statusMsg = fmt.Sprintf("[%d selected] %s", selectedCount, statusMsg)
	}

	if len(m.clipboard) > 0 && selectedCount == 0 && !strings.Contains(statusMsg, "[v] Paste") {
		statusMsg += " | [v] Paste"
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
		if r == '[' {
			if current.Len() > 0 {
				result.WriteString(baseStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = true
			result.WriteString(keyStyle.Render("["))
		} else if r == ']' {
			if current.Len() > 0 {
				result.WriteString(keyStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = false
			result.WriteString(keyStyle.Render("]"))
		} else {
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
