package file

import (
	"strings"

	"fm/internal/files/core"
	"fm/internal/files/format"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
)

// Props contains all data needed to render the file list
type Props struct {
	Width            int
	Height           int
	Cursor           int
	Offset           int
	Items            []core.Item
	ShowHeader       bool
	ShowSize         bool
	ShowDateModified bool
	SelectMode       bool
	SizeFormatIndex  int
	DateFormatIndex  int
	DateLayout       string
	Styles           theme.Stylesheet
	SelectedPaths    map[string]bool
}

// Layout contains calculated dimensions for the list view
type Layout struct {
	ViewportHeight     int
	NameWidth          int
	DateWidth          int
	SizeWidth          int
	MarkerWidth        int
	GitMarkerWidth     int
	PermIndicatorWidth int
	ColumnGap          int
}

// Render renders the complete file list view
func Render(props Props) string {
	if props.Height <= 0 {
		return ""
	}

	layout := calculateLayout(props)
	var rows []string

	// Render header if enabled
	headerRows := renderHeaderRows(props, layout)
	rows = append(rows, headerRows...)

	viewportHeight := props.Height - len(headerRows)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	// Calculate end of viewport
	end := props.Offset + viewportHeight
	if end > len(props.Items) {
		end = len(props.Items)
	}

	for i := props.Offset; i < end; i++ {
		item := props.Items[i]
		// Use source of truth for selection
		if props.SelectedPaths != nil {
			item.Selected = props.SelectedPaths[item.Path]
		}
		rows = append(rows, renderRow(props, item, i == props.Cursor, layout))
	}

	// Fill remaining space
	for len(rows) < props.Height {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}

func renderHeaderRows(props Props, layout Layout) []string {
	if !props.ShowHeader || len(props.Items) == 0 {
		return nil
	}

	var columns []ui.Column

	// Adjust for selection marker and git markers
	prefixWidth := layout.MarkerWidth + layout.GitMarkerWidth + layout.PermIndicatorWidth
	columns = append(columns, ui.Column{Title: "Name", Width: layout.NameWidth + prefixWidth})

	if props.ShowDateModified {
		columns = append(columns, ui.Column{Title: "Date Modified", Width: layout.DateWidth})
	}
	if props.ShowSize {
		columns = append(columns, ui.Column{Title: "Size", Width: layout.SizeWidth})
	}

	header := ui.RenderHeader(props.Width, columns, layout.ColumnGap, props.Styles)
	return strings.Split(header, "\n")
}

func calculateLayout(props Props) Layout {
	sizeWidth := 11
	if props.SizeFormatIndex < len(format.SizeFormats) {
		sizeWidth = 11
	}

	dateWidth := len(props.DateLayout)
	if dateWidth == 0 {
		if props.DateFormatIndex < len(format.DateFormats) {
			dateWidth = len(format.DateFormats[props.DateFormatIndex].Layout)
		} else {
			dateWidth = 10
		}
	}

	const columnGap = 2
	markerWidth := 0
	if props.SelectMode {
		markerWidth = 4
	}
	gitMarkerWidth := 2

	nameWidth := props.Width - markerWidth - gitMarkerWidth - 2
	if props.ShowSize {
		nameWidth -= (sizeWidth + columnGap)
	}
	if props.ShowDateModified {
		nameWidth -= (dateWidth + columnGap)
	}

	if nameWidth < 1 {
		nameWidth = 1
	}

	return Layout{
		ViewportHeight:     props.Height,
		NameWidth:          nameWidth,
		DateWidth:          dateWidth,
		SizeWidth:          sizeWidth,
		MarkerWidth:        markerWidth,
		GitMarkerWidth:     gitMarkerWidth,
		PermIndicatorWidth: 1,
		ColumnGap:          columnGap,
	}
}
