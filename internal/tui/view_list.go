package tui

import (
	"fmt"
	"strings"

	"filemanager/internal/files"

	"github.com/charmbracelet/lipgloss"
)

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
