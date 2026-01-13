package list

import (
	"strings"

	"fm/internal/files/core"
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
}

// Render renders the complete file list view
func Render(props Props) string {
	if props.Height < 0 {
		return ""
	}

	layout := CalculateLayout(props)
	headerRows := renderHeaderRows(props, layout)

	var rows []string
	if len(headerRows) > 0 {

		rows = append(rows, headerRows...)
		// Adjust viewport height for headers
		layout.ViewportHeight -= len(headerRows)
	}

	end := props.Offset + layout.ViewportHeight
	if end > len(props.Items) {
		end = len(props.Items)
	}

	for i := props.Offset; i < end; i++ {
		item := props.Items[i]

		rows = append(rows, renderRow(props, item, i == props.Cursor, layout))
	}

	// Fill remaining space with empty lines
	for i := len(rows); i < (layout.ViewportHeight + len(headerRows)); i++ {

		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}
