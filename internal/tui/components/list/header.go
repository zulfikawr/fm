package list

import (
	"fmt"
	"strings"
)

// renderHeaderRows renders the column headers and separators
func renderHeaderRows(props Props, layout Layout) []string {
	var rows []string
	if !props.ShowHeader || len(props.Items) == 0 {
		return rows
	}

	nameH := fmt.Sprintf("% -*s", layout.NameWidth, "Name")
	dateH := ""
	if props.ShowDateModified {
		dateH = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth, "Date Modified")
	}
	sizeH := ""
	if props.ShowSize {
		sizeH = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.SizeWidth, "Size")
	}

	headerContent := fmt.Sprintf("%*s%*s%s%s%s", layout.MarkerWidth, "", layout.GitMarkerWidth, "", nameH, dateH, sizeH)

	// Top separator
	sep := props.Styles.Separator.Width(props.Width).Render(strings.Repeat("-", props.Width))
	rows = append(rows, sep)

	// Header text
	rows = append(rows, props.Styles.ListHeader.Width(props.Width).Render(headerContent))

	// Bottom separator
	rows = append(rows, sep)

	return rows
}
