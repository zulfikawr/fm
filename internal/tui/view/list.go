package view

import (
	"fmt"
	"strings"

	"fm/internal/files"
	"fm/internal/files/format"
	"fm/internal/tui/components/loading"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

type ListLayout struct {
	ViewportHeight     int
	NameWidth          int
	DateWidth          int
	SizeWidth          int
	MarkerWidth        int
	GitMarkerWidth     int
	PermIndicatorWidth int
	ColumnGap          int
}

// RenderList renders the file list view
func RenderList(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	viewportHeight := s.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	if s.UI.Loading && len(s.FilteredItems) == 0 {
		return renderLoading(s, viewportHeight, styles)
	}

	layout := calculateLayout(s, viewportHeight)

	// Column Headers
	var headerRows []string
	if s.Config.ShowHeader && len(s.FilteredItems) > 0 {
		nameH := fmt.Sprintf("% -*s", layout.NameWidth, "Name")
		dateH := ""
		if s.Config.ShowDateModified {
			dateH = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth, "Date Modified")
		}
		sizeH := ""
		if s.Config.ShowSize {
			sizeH = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.SizeWidth, "Size")
		}

		headerContent := fmt.Sprintf("%*s%*s%s%s%s", layout.MarkerWidth, "", layout.GitMarkerWidth, "", nameH, dateH, sizeH)

		// Top separator
		sep := styles.Separator.Width(s.Width).Render(strings.Repeat("-", s.Width))
		headerRows = append(headerRows, sep)

		// Header text
		headerRows = append(headerRows, styles.ListHeader.Width(s.Width).Render(headerContent))

		// Bottom separator
		headerRows = append(headerRows, sep)

		layout.ViewportHeight -= 3 // Room for 2 separators and 1 header line
	}

	var rows []string
	if len(headerRows) > 0 {
		rows = append(rows, headerRows...)
	}

	end := s.Offset + layout.ViewportHeight
	if end > len(s.FilteredItems) {
		end = len(s.FilteredItems)
	}

	for i := s.Offset; i < end; i++ {
		item := s.FilteredItems[i]
		rows = append(rows, renderRow(s, item, i == s.Cursor, styles, layout))
	}

	for i := len(rows); i < (layout.ViewportHeight + len(headerRows)); i++ {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func calculateLayout(s *ViewState, viewportHeight int) ListLayout {
	sizeWidth := 11
	switch s.Config.SizeFormatIndex {
	case 1:
		sizeWidth = 12
	case 2:
		sizeWidth = 15
	}

	dateWidth := len(format.DateFormats[s.Config.DateFormatIndex].Layout)
	if dateWidth < 10 {
		dateWidth = 10
	}

	const columnGap = 2
	markerWidth := 0
	if s.UI.SelectMode {
		markerWidth = 4
	}
	gitMarkerWidth := 3 // git status + perm indicator space

	nameWidth := s.Width - markerWidth - gitMarkerWidth
	if s.Config.ShowSize {
		nameWidth -= (sizeWidth + columnGap)
	}
	if s.Config.ShowDateModified {
		nameWidth -= (dateWidth + columnGap)
	}

	if nameWidth < 1 {
		nameWidth = 1
	}

	return ListLayout{
		ViewportHeight:     viewportHeight,
		NameWidth:          nameWidth,
		DateWidth:          dateWidth,
		SizeWidth:          sizeWidth,
		MarkerWidth:        markerWidth,
		GitMarkerWidth:     gitMarkerWidth,
		PermIndicatorWidth: 1,
		ColumnGap:          columnGap,
	}
}

func renderRow(s *ViewState, item files.Item, isCursor bool, styles theme.Stylesheet, layout ListLayout) string {
	marker := renderMarker(s, item)
	gitMarker := renderGitMarker(item, isCursor, styles)
	permIndicator := renderPermIndicator(item, isCursor, styles)
	namePart := renderNamePart(s, item, isCursor, styles, layout)
	metaPart := renderMetaPart(s, item, isCursor, styles, layout)

	if isCursor {
		sStyle := styles.SelectedItem.UnsetPadding().UnsetWidth()

		row := sStyle.Render(marker) +
			gitMarker +
			sStyle.Render(permIndicator) +
			sStyle.Render(namePart) +
			sStyle.Render(metaPart)

		currentWidth := lipgloss.Width(row)
		if currentWidth < s.Width {
			row += sStyle.Render(strings.Repeat(" ", s.Width-currentWidth))
		}
		return row
	}

	row := fmt.Sprintf("%s%s%s%s%s", marker, gitMarker, permIndicator, namePart, metaPart)
	return styles.Item.Width(s.Width).Render(row)
}

func renderMarker(s *ViewState, item files.Item) string {
	if !s.UI.SelectMode {
		return ""
	}
	if item.IsUp {
		return "    "
	}
	if item.Selected {
		return "[x] "
	}
	return "[ ] "
}

func renderGitMarker(item files.Item, isCursor bool, styles theme.Stylesheet) string {
	gitMarker := "  "
	if item.GitStatus != "" {
		gitMarker = item.GitStatus + " "
	}

	if isCursor {
		sStyle := styles.SelectedItem.UnsetPadding().UnsetWidth()
		if style, ok := styles.GitStyles[item.GitStatus]; ok {
			return style.Inherit(sStyle).Render(gitMarker)
		}
		return sStyle.Render(gitMarker)
	}

	if style, ok := styles.GitStyles[item.GitStatus]; ok {
		return style.Render(gitMarker)
	}
	return gitMarker
}

func renderPermIndicator(item files.Item, isCursor bool, styles theme.Stylesheet) string {
	indicator := " "
	if !item.CanWrite && !item.IsUp && !item.IsGhost {
		indicator = "!"
	}

	if !isCursor && indicator == "!" {
		return styles.DimCol.Render("!")
	}
	return indicator
}

func renderNamePart(_ *ViewState, item files.Item, isCursor bool, styles theme.Stylesheet, layout ListLayout) string {
	var nameStr string
	if item.IsUp {
		nameStr = item.Name
	} else if item.IsDir {
		nameStr = item.Name + "/"
	} else {
		nameStr = item.Name
	}

	if len(nameStr) > layout.NameWidth {
		nameStr = nameStr[:layout.NameWidth-1] + "…"
	}

	if isCursor {
		return fmt.Sprintf("%-*s", layout.NameWidth, nameStr)
	}

	// Git Status Coloring for Name
	nameStyle := styles.FileCol
	if item.IsUp {
		nameStyle = styles.DimCol
	} else if item.IsDir {
		nameStyle = styles.DirCol
	} else if item.Mode&0111 != 0 {
		nameStyle = styles.ExecCol
	}

	if item.IsGhost {
		nameStyle = styles.GitGhost
	} else if style, ok := styles.GitStyles[item.GitStatus]; ok {
		nameStyle = style
	}

	if !item.CanRead && !item.IsUp {
		nameStyle = styles.DimCol
	}

	styledName := nameStyle.Render(nameStr)
	gap := layout.NameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}
	return styledName + strings.Repeat(" ", gap)
}

func renderMetaPart(s *ViewState, item files.Item, isCursor bool, styles theme.Stylesheet, layout ListLayout) string {
	datePart := ""
	if s.Config.ShowDateModified {
		dateStr := item.FormattedDate
		datePart = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth, dateStr)
		if !isCursor {
			datePart = styles.DimCol.Render(datePart)
		}
	}

	sizePart := ""
	if s.Config.ShowSize {
		sizeStr := item.FormattedSize
		sizePart = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.SizeWidth, sizeStr)
		if !isCursor {
			sizePart = styles.DimCol.Render(sizePart)
		}
	}

	return datePart + sizePart
}

// renderLoading renders the loading state with spinner
func renderLoading(s *ViewState, viewportHeight int, styles theme.Stylesheet) string {
	return loading.Render(loading.Props{
		Width:   s.Width,
		Height:  viewportHeight,
		Message: "Loading...",
		Spinner: s.LoadingSpinner,
		Styles:  styles,
	})
}
