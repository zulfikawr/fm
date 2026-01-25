package file

import (
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"
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
	EnableIcons      bool
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
	IconWidth          int
	ColumnGap          int
	ShowSize           bool
	ShowDate           bool
	EnableIcons        bool
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
		rows = append(rows, renderRow(RowContext{
			Props:    props,
			Item:     item,
			IsCursor: i == props.Cursor,
			Layout:   layout,
		}))
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
	prefixWidth := layout.MarkerWidth + layout.GitMarkerWidth + layout.PermIndicatorWidth + layout.IconWidth
	columns = append(columns, ui.Column{Title: "Name", Width: layout.NameWidth + prefixWidth})

	if layout.ShowDate {
		columns = append(columns, ui.Column{Title: "Date Modified", Width: layout.DateWidth})
	}
	if layout.ShowSize {
		columns = append(columns, ui.Column{Title: "Size", Width: layout.SizeWidth})
	}

	header := ui.RenderHeader(ui.HeaderProps{
		Width:   props.Width,
		Columns: columns,
		Gap:     layout.ColumnGap,
		Styles:  props.Styles,
	})
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
	permIndicatorWidth := 1
	iconWidth := 0
	if props.EnableIcons {
		iconWidth = 3 // Icon + 2 spaces
	}

	// Initial available width for name, size and date
	// -2 for safety margin/padding
	// -1 for gap after markers
	availableWidth := props.Width - markerWidth - gitMarkerWidth - permIndicatorWidth - iconWidth - 3

	showSize := props.ShowSize
	showDate := props.ShowDateModified

	const minNameWidth = 20

	nameWidth := availableWidth
	if showSize {
		nameWidth -= (sizeWidth + columnGap)
	}
	if showDate {
		nameWidth -= (dateWidth + columnGap)
	}

	// If name width is too small, hide Date Modified first
	if nameWidth < minNameWidth && showDate {
		showDate = false
		nameWidth += (dateWidth + columnGap)
	}

	// If still too small, hide Size
	if nameWidth < minNameWidth && showSize {
		showSize = false
		nameWidth += (sizeWidth + columnGap)
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
		PermIndicatorWidth: permIndicatorWidth,
		IconWidth:          iconWidth,
		ColumnGap:          columnGap,
		ShowSize:           showSize,
		ShowDate:           showDate,
		EnableIcons:        props.EnableIcons,
	}
}
