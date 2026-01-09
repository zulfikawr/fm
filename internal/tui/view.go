package tui

import (
	"fmt"
	"os"
	"strings"

	"filemanager/internal/files"

	"github.com/charmbracelet/lipgloss"
)

// View renders the application UI.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var body string
	if m.settingsOpen {
		body = m.renderSettingsList(header, footer)
	} else {
		body = m.renderList(header, footer)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *Model) renderSettingsList(header, footer string) string {
	viewportHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	type setting struct {
		label string
		value string
	}

	groups := []struct {
		title    string
		settings []setting
		startIdx int
	}{
		{
			title:    "File Operations",
			startIdx: 0,
			settings: []setting{
				{"Show Hidden Files", m.formatBool(m.cfg.ShowHidden)},
				{"Case-Sensitive Search", m.formatBool(m.cfg.CaseSensitive)},
				{"Confirm Operations", m.formatBool(m.cfg.ConfirmOperations)},
				{"Wrap Navigation", m.formatBool(m.cfg.WrapNavigation)},
			},
		},
		{
			title:    "Display Options",
			startIdx: 4,
			settings: []setting{
				{"Show Column Headers", m.formatBool(m.cfg.ShowHeader)},
				{"Enable Git Status", m.formatBool(m.cfg.EnableGit)},
				{"Show File Size", m.formatBool(m.cfg.ShowSize)},
				{"Size Format", fmt.Sprintf("< %s >", files.SizeFormats[m.cfg.SizeFormatIndex])},
				{"Show Date Modified", m.formatBool(m.cfg.ShowDateModified)},
				{"Date Format", fmt.Sprintf("< %s >", files.DateFormats[m.cfg.DateFormatIndex].Name)},
			},
		},
		{
			title:    "Appearance",
			startIdx: 10,
			settings: []setting{
				{"Theme", fmt.Sprintf("< %s >", Themes[m.cfg.ThemeIndex].Name)},
			},
		},
	}

	var rows []string
	rows = append(rows, "") // Add a line above the first group
	for i, g := range groups {
		if i > 0 {
			rows = append(rows, "")
		}
		rows = append(rows, m.styles.SettingsHeader.Width(m.width).Render(g.title))
		for j, s := range g.settings {
			idx := g.startIdx + j
			style := m.styles.SettingsItem
			if idx == m.settingsCursor {
				style = m.styles.SettingsSelectedItem
			}

			// Dim inactive settings
			inactive := false
			if idx == 7 && !m.cfg.ShowSize { // Size Format
				inactive = true
			} else if idx == 9 && !m.cfg.ShowDateModified { // Date Format
				inactive = true
			}
			val := s.value
			if inactive {
				style = m.styles.DimCol.PaddingLeft(2)
				val = m.styles.DimCol.Render(s.value)
			}

			labelWidth := 25
			if m.width < 40 {
				labelWidth = m.width - 12
			}
			if labelWidth < 5 {
				labelWidth = 5
			}

			label := s.label + ":"
			if len(label) > labelWidth {
				label = label[:labelWidth-1] + "…"
			}

			// Calculate width without ANSI codes for proper alignment
			valWidth := lipgloss.Width(s.value)
			content := fmt.Sprintf("%-*s %s", labelWidth, label, val)

			if labelWidth+1+valWidth > m.width-2 {
				// If still too long, prioritize showing the value
				availableLabelWidth := m.width - 2 - valWidth - 1
				if availableLabelWidth > 0 {
					label = s.label + ":"
					if len(label) > availableLabelWidth {
						label = label[:availableLabelWidth-1] + "…"
					}
					content = fmt.Sprintf("%-*s %s", availableLabelWidth, label, val)
				}
			}

			rows = append(rows, style.Width(m.width).Render(content))
		}
	}

	for i := len(rows); i < viewportHeight; i++ {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

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

	parts := strings.Split(m.path, string(os.PathSeparator))
	var cleanParts []string
	if m.path == "/" {
		cleanParts = []string{"/"}
	} else {
		for _, p := range parts {
			if p != "" {
				cleanParts = append(cleanParts, p)
			}
		}
	}

	// Ensure all parts in header inherit header's background
	dimHeaderStyle := m.styles.DimCol.Inherit(m.styles.Header)

	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, m.styles.Header.UnsetPadding().UnsetWidth().Render(p))
	}

	separator := dimHeaderStyle.Render(" > ")
	breadcrumb := strings.Join(styledParts, separator)
	if !strings.HasPrefix(breadcrumb, "/") && strings.HasPrefix(m.path, "/") {
		breadcrumb = dimHeaderStyle.Render("/ ") + breadcrumb
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
			// Cursor is on ".." - show as 0 or handle specifically
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

func (m *Model) renderList(header, footer string) string {
	viewportHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	if m.loading && len(m.filteredItems) == 0 {
		return lipgloss.NewStyle().
			Height(viewportHeight).
			Width(m.width).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Loading...")
	}

	// Column Headers
	var headerRows []string
	if m.cfg.ShowHeader && len(m.filteredItems) > 0 {
		sizeWidth := 10
		if m.cfg.SizeFormatIndex == 1 {
			sizeWidth = 12
		} else if m.cfg.SizeFormatIndex == 2 {
			sizeWidth = 15
		}

		dateWidth := len(files.DateFormats[m.cfg.DateFormatIndex].Layout)
		if dateWidth < 10 {
			dateWidth = 10
		}

		const columnGap = 2
		markerWidth := 0
		if m.selectMode {
			markerWidth = 4
		}
		gitMarkerWidth := 2

		nameWidth := m.width - markerWidth - gitMarkerWidth
		if m.cfg.ShowSize {
			nameWidth -= (sizeWidth + columnGap)
		}
		if m.cfg.ShowDateModified {
			nameWidth -= (dateWidth + columnGap)
		}

		if nameWidth < 1 {
			nameWidth = 1
		}

		nameH := fmt.Sprintf("% -*s", nameWidth, "Name")
		dateH := ""
		if m.cfg.ShowDateModified {
			dateH = fmt.Sprintf("%*s%*s", columnGap, "", dateWidth, "Date Modified")
		}
		sizeH := ""
		if m.cfg.ShowSize {
			sizeH = fmt.Sprintf("%*s%*s", columnGap, "", sizeWidth, "Size")
		}

		headerContent := fmt.Sprintf("%*s%*s%s%s%s", markerWidth, "", gitMarkerWidth, "", nameH, dateH, sizeH)

		// Top separator
		sep := m.styles.Separator.Width(m.width).Render(strings.Repeat("-", m.width))
		headerRows = append(headerRows, sep)

		// Header text
		headerRows = append(headerRows, m.styles.ListHeader.Width(m.width).Render(headerContent))

		// Bottom separator
		headerRows = append(headerRows, sep)

		viewportHeight -= 3 // Room for 2 separators and 1 header line
	}

	var rows []string
	if len(headerRows) > 0 {
		rows = append(rows, headerRows...)
	}

	end := m.offset + viewportHeight
	if end > len(m.filteredItems) {
		end = len(m.filteredItems)
	}

	for i := m.offset; i < end; i++ {
		item := m.filteredItems[i]

		rows = append(rows, m.renderRow(item, i == m.cursor))
	}

	for i := len(rows); i < (viewportHeight + len(headerRows)); i++ {

		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func (m *Model) renderRow(item files.Item, selected bool) string {
	marker := ""
	if m.selectMode {
		marker = "[ ] "
		if item.Selected {
			marker = "[x] "
		}
		if item.IsUp {
			marker = "    "
		}
	}

	gitMarker := "  "
	if item.GitStatus != "" {
		gitMarker = item.GitStatus + " "
	}

	var nameStr string
	if item.IsUp {
		nameStr = item.Name
	} else if item.IsDir {
		nameStr = item.Name + "/"
	} else {
		nameStr = item.Name
	}

	sizeStr := ""
	if m.cfg.ShowSize {
		if item.IsUp {
			sizeStr = ""
		} else {
			sizeStr = files.FormatSize(item.Size, m.cfg.SizeFormatIndex)
		}
	}

	dateStr := ""
	if m.cfg.ShowDateModified {
		if item.IsUp {
			dateStr = ""
		} else {
			layout := files.DateFormats[m.cfg.DateFormatIndex].Layout
			dateStr = item.MTime.Format(layout)
		}
	}

	// Calculate widths
	sizeWidth := 10
	if m.cfg.SizeFormatIndex == 1 { // Full (KB, MB, GB)
		sizeWidth = 12
	} else if m.cfg.SizeFormatIndex == 2 { // Bytes
		sizeWidth = 15
	}

	dateWidth := len(files.DateFormats[m.cfg.DateFormatIndex].Layout)
	if dateWidth < 10 {
		dateWidth = 10
	}
	markerWidth := len(marker)
	gitMarkerWidth := 2 // "M "
	const columnGap = 2

	availableWidth := m.width - markerWidth - gitMarkerWidth

	if m.cfg.ShowSize {
		availableWidth -= (sizeWidth + columnGap)
	}
	if m.cfg.ShowDateModified {
		availableWidth -= (dateWidth + columnGap)
	}

	nameWidth := availableWidth
	if nameWidth < 1 {
		nameWidth = 1
	}

	if len(nameStr) > nameWidth {
		nameStr = nameStr[:nameWidth-1] + "…"
	}

	// Prepare columns
	datePart := ""
	if m.cfg.ShowDateModified {
		datePart = fmt.Sprintf("%*s%*s", columnGap, "", dateWidth, dateStr)
	}
	sizePart := ""
	if m.cfg.ShowSize {
		sizePart = fmt.Sprintf("%*s%*s", columnGap, "", sizeWidth, sizeStr)
	}

	lineContent := fmt.Sprintf("%s%s%-*s%s%s", marker, gitMarker, nameWidth, nameStr, datePart, sizePart)

	if selected {
		return m.styles.SelectedItem.Width(m.width).Render(lineContent)
	}

	// Apply dimmed style to date and size
	if m.cfg.ShowDateModified {
		datePart = m.styles.DimCol.Render(datePart)
	}
	if m.cfg.ShowSize {
		sizePart = m.styles.DimCol.Render(sizePart)
	}

	// Git Status Coloring for Name
	nameStyle := m.styles.FileCol
	if item.IsUp {
		nameStyle = m.styles.DimCol
	} else if item.IsDir {
		nameStyle = m.styles.DirCol
	} else if item.Mode&0111 != 0 {
		nameStyle = m.styles.ExecCol
	}

	if item.IsGhost {
		nameStyle = m.styles.GitGhost
	} else {
		switch item.GitStatus {
		case "M":
			nameStyle = m.styles.GitMod
		case "A":
			nameStyle = m.styles.GitStaged
		case "?":
			nameStyle = m.styles.GitUntracked
		case "U":
			nameStyle = m.styles.GitConflict
		case "!":
			nameStyle = m.styles.GitIgnored
		}
	}

	styledName := nameStyle.Render(nameStr)

	// Git Marker Coloring
	styledGitMarker := gitMarker
	switch item.GitStatus {
	case "M":
		styledGitMarker = m.styles.GitMod.Render(gitMarker)
	case "A":
		styledGitMarker = m.styles.GitStaged.Render(gitMarker)
	case "?":
		styledGitMarker = m.styles.GitUntracked.Render(gitMarker)
	case "U":
		styledGitMarker = m.styles.GitConflict.Render(gitMarker)
	case "D":
		styledGitMarker = m.styles.GitConflict.Render(gitMarker)
	case "!":
		styledGitMarker = m.styles.GitIgnored.Render(gitMarker)
	}

	gap := nameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}

	row := fmt.Sprintf("%s%s%s%s%s%s", marker, styledGitMarker, styledName, strings.Repeat(" ", gap), datePart, sizePart)
	return m.styles.Item.Width(m.width).Render(row)
}
