package file

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
)

// renderRow renders a single row in the file list
func renderRow(props Props, item core.Item, isCursor bool, layout Layout) string {
	marker := renderMarker(props, item, isCursor)
	gitMarker := renderGitMarker(props, item, isCursor)
	permIndicator := renderPermIndicator(props, item, isCursor)
	namePart := renderNamePart(props, item, isCursor, layout)
	metaPart := renderMetaPart(props, item, isCursor, layout)

	content := marker + gitMarker + permIndicator + namePart + metaPart

	if isCursor {
		return ui.ItemRow(content, props.Width, true, props.Styles)
	}

	return props.Styles.Item.Width(props.Width).Render(content)
}

func renderMarker(props Props, item core.Item, isCursor bool) string {
	if !props.SelectMode {
		return ""
	}
	return ui.Marker(item.Selected, item.IsUp, isCursor, props.Styles)
}

func renderGitMarker(props Props, item core.Item, isCursor bool) string {
	gitMarker := " "
	if item.GitStatus != "" {
		gitMarker = item.GitStatus
	}

	if isCursor {
		sStyle := props.Styles.SelectedItem.UnsetPadding().UnsetWidth()
		if style, ok := props.Styles.GitStyles[item.GitStatus]; ok {
			return style.Inherit(sStyle).Render(gitMarker)
		}
		return sStyle.Render(gitMarker)
	}

	if style, ok := props.Styles.GitStyles[item.GitStatus]; ok {
		return style.Render(gitMarker)
	}
	return gitMarker
}

func renderPermIndicator(props Props, item core.Item, isCursor bool) string {
	indicator := " "
	if !item.CanWrite && !item.IsUp && !item.IsGhost {
		indicator = "!"
	}

	if isCursor {
		return props.Styles.SelectedItem.UnsetPadding().UnsetWidth().Render(indicator)
	}

	if indicator == "!" {
		return props.Styles.DimCol.Render("!")
	}
	return indicator
}

func renderNamePart(props Props, item core.Item, isCursor bool, layout Layout) string {
	var nameStr string
	if item.IsUp {
		nameStr = item.Name
	} else if item.IsDir {
		nameStr = item.Name + "/"
	} else {
		nameStr = item.Name
	}

	nameStr = ui.Truncate(nameStr, layout.NameWidth)

	// Git Status Coloring for Name
	nameStyle := props.Styles.FileCol
	if item.IsUp {
		nameStyle = props.Styles.DimCol
	} else if item.IsDir {
		nameStyle = props.Styles.DirCol
	} else if item.HasMetadata && item.Mode&0o111 != 0 {
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

	if isCursor {
		sStyle := props.Styles.SelectedItem.UnsetPadding().UnsetWidth()
		styledName := nameStyle.Inherit(sStyle).Render(nameStr)
		gap := layout.NameWidth - len(nameStr)
		if gap < 0 {
			gap = 0
		}
		return styledName + sStyle.Render(strings.Repeat(" ", gap))
	}

	styledName := nameStyle.Render(nameStr)
	gap := layout.NameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}
	return styledName + strings.Repeat(" ", gap)
}

func renderMetaPart(props Props, item core.Item, isCursor bool, layout Layout) string {
	sStyle := props.Styles.SelectedItem.UnsetPadding().UnsetWidth()

	if !item.HasMetadata && !item.IsUp && !item.IsGhost {
		content := fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth+layout.ColumnGap+layout.SizeWidth, "...")
		if isCursor {
			return props.Styles.DimCol.Inherit(sStyle).Render(content)
		}
		return props.Styles.DimCol.Render(content)
	}

	datePart := ""
	if props.ShowDateModified {
		dateStr := item.FormattedDate
		content := fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.DateWidth, dateStr)
		if isCursor {
			datePart = props.Styles.DimCol.Inherit(sStyle).Render(content)
		} else {
			datePart = props.Styles.DimCol.Render(content)
		}
	}

	sizePart := ""
	if props.ShowSize {
		sizeStr := item.FormattedSize
		content := fmt.Sprintf("%*s%*s", layout.ColumnGap, "", layout.SizeWidth, sizeStr)
		if isCursor {
			sizePart = props.Styles.DimCol.Inherit(sStyle).Render(content)
		} else {
			sizePart = props.Styles.DimCol.Render(content)
		}
	}

	return datePart + sizePart
}
