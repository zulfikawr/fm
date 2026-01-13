package list

import (
	"fmt"

	"fm/internal/files/core"
)

// renderMetaPart renders the file size and modification date columns
func renderMetaPart(props Props, item core.Item, isCursor bool, layout Layout) string {
	datePart := ""
	if props.ShowDateModified {
		dateStr := item.FormattedDate
		datePart = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth, dateStr)
		if !isCursor {
			datePart = props.Styles.DimCol.Render(datePart)
		}
	}

	sizePart := ""
	if props.ShowSize {
		sizeStr := item.FormattedSize
		sizePart = fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.SizeWidth, sizeStr)
		if !isCursor {
			sizePart = props.Styles.DimCol.Render(sizePart)
		}
	}

	return datePart + sizePart
}
