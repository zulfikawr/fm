package tui

import (
	"fmt"
	"os"
	"strings"

	"filemanager/internal/files"

	"github.com/charmbracelet/lipgloss"
)

// View renders the application UI.
func (m Model) View() string {
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

func (m Model) renderSettingsList(header, footer string) string {
	viewportHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	settings := []struct {
		label string
		value string
	}{
		{"Show Hidden Files", formatBool(m.cfg.ShowHidden)},
		{"Case-Sensitive Search", formatBool(m.cfg.CaseSensitive)},
		{"Confirm Operations", formatBool(m.cfg.ConfirmOperations)},
		{"Wrap Navigation", formatBool(m.cfg.WrapNavigation)},
		{"Enable Git", formatBool(m.cfg.EnableGit)},
		{"Theme", Themes[m.cfg.ThemeIndex].Name},
	}

	var rows []string
	for i, s := range settings {
		style := m.styles.Item
		if i == m.settingsCursor {
			style = m.styles.SelectedItem
		}

		content := fmt.Sprintf("% -25s %s", s.label, s.value)
		rows = append(rows, style.Width(m.width).Render(content))
	}

	for i := len(rows); i < viewportHeight; i++ {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func formatBool(b bool) string {
	if b {
		return "[ON]"
	}
	return "[OFF]"
}

func (m Model) renderHeader() string {
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

	breadcrumb := strings.Join(cleanParts, " > ")
	if !strings.HasPrefix(breadcrumb, "/") && strings.HasPrefix(m.path, "/") {
		breadcrumb = "/ " + breadcrumb
	}

	if m.gitBranch != "" {
		breadcrumb += fmt.Sprintf(" (%s)", m.gitBranch)
	}

	return m.styles.Header.Width(m.width).Render(breadcrumb)
}

func (m Model) renderFooter() string {
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
		return m.styles.Footer.Width(m.width).Render(" " + prompt)
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
			helpMsg = "Enable git status markers"
		case 5:
			helpMsg = "Change the application color scheme"
		}

		leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [Esc/.] Back"
		rightPart := helpMsg + " "

		gap := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
		if gap < 0 {
			gap = 0
		}

		footerContent := leftPart + strings.Repeat(" ", gap) + rightPart
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
			statusMsg = "[↑↓] Nav | [s] Sort | [/] Search | [.] Settings"
		}
	}

	if selectedCount > 0 {
		statusMsg = fmt.Sprintf("[%d selected] %s", selectedCount, statusMsg)
	}

	if len(m.clipboard) > 0 && selectedCount == 0 && !strings.Contains(statusMsg, "[v] Paste") {
		statusMsg += " | [v] Paste"
	}

	sortStr := m.sortMode.String()

	leftPart := fmt.Sprintf(" %d/%d  %s", m.cursor+1, len(m.filteredItems), statusMsg)
	rightPart := sortStr + " "

	gap := m.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if gap < 0 {
		gap = 0
	}

	footerContent := leftPart + strings.Repeat(" ", gap) + rightPart
	return m.styles.Footer.Width(m.width).Render(footerContent)
}

func (m Model) renderList(header, footer string) string {
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

	var rows []string
	end := m.offset + viewportHeight
	if end > len(m.filteredItems) {
		end = len(m.filteredItems)
	}

	for i := m.offset; i < end; i++ {
		item := m.filteredItems[i]
		rows = append(rows, m.renderRow(item, i == m.cursor))
	}

	for i := len(rows); i < viewportHeight; i++ {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderRow(item files.Item, selected bool) string {
	marker := "[ ] "
	if item.Selected {
		marker = "[x] "
	}
	if item.IsUp {
		marker = "    "
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

	var sizeStr string
	if item.IsUp {
		sizeStr = ""
	} else if item.IsDir {
		sizeStr = "<DIR>"
	} else {
		sizeStr = files.FormatSize(item.Size)
	}

	// nameWidth: m.width - Size(10) - Padding(2) - Marker(4) - GitMarker(2)
	nameWidth := m.width - 10 - 2 - 4 - 2
	if nameWidth < 1 {
		nameWidth = 1
	}

	if len(nameStr) > nameWidth {
		nameStr = nameStr[:nameWidth-1] + "…"
	}

	lineContent := fmt.Sprintf("%s%s%-*s%10s", marker, gitMarker, nameWidth, nameStr, sizeStr)

	if selected {
		return m.styles.SelectedItem.Width(m.width).Render(lineContent)
	}

	// Git Status Coloring for Name
	nameStyle := m.styles.FileCol
	if item.IsUp || item.IsDir {
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
	}

	gap := nameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}

	row := fmt.Sprintf("%s%s%s%s%10s", marker, styledGitMarker, styledName, strings.Repeat(" ", gap), sizeStr)
	return m.styles.Item.Width(m.width).Render(row)
}
