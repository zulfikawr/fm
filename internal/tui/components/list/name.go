package list

import (
	"fmt"
	"strings"

	"fm/internal/files/core"
)

// renderNamePart renders the filename/path with intelligent truncation and coloring
func renderNamePart(props Props, item core.Item, isCursor bool, layout Layout) string {
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
		return fmt.Sprintf("% -*s", layout.NameWidth, nameStr)
	}

	// Git Status Coloring for Name
	nameStyle := props.Styles.FileCol
	if item.IsUp {
		nameStyle = props.Styles.DimCol
	} else if item.IsDir {
		nameStyle = props.Styles.DirCol
	} else if item.Mode&0111 != 0 {
		nameStyle = props.Styles.ExecCol
	}

	if item.IsGhost {
		nameStyle = props.Styles.GitGhost
	} else if style, ok := props.Styles.GitStyles[item.GitStatus]; ok {
		nameStyle = style
	}

	if !item.CanRead && !item.IsUp {
		nameStyle = props.Styles.DimCol
	}

	styledName := nameStyle.Render(nameStr)
	gap := layout.NameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}
	return styledName + strings.Repeat(" ", gap)
}
