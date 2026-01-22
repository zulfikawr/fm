package file

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
)

// RowContext encapsulates all data needed to render a single row
type RowContext struct {
	Props    Props
	Item     core.Item
	IsCursor bool
	Layout   Layout
}

// renderRow renders a single row in the file list
func renderRow(ctx RowContext) string {
	marker := renderMarker(ctx)
	gitMarker := renderGitMarker(ctx)
	permIndicator := renderPermIndicator(ctx)
	namePart := renderNamePart(ctx)
	metaPart := renderMetaPart(ctx)

	content := marker + gitMarker + permIndicator + namePart + metaPart

	if ctx.IsCursor {
		return ui.ItemRow(content, ui.ListProps{
			Width:    ctx.Props.Width,
			IsCursor: true,
			Styles:   ctx.Props.Styles,
		})
	}

	return ctx.Props.Styles.Item.Width(ctx.Props.Width).Render(content)
}

func renderMarker(ctx RowContext) string {
	if !ctx.Props.SelectMode {
		return ""
	}
	return ui.Marker(ui.ListProps{
		Selected: ctx.Item.Selected,
		IsUp:     ctx.Item.IsUp,
		IsCursor: ctx.IsCursor,
		Styles:   ctx.Props.Styles,
	})
}

func renderGitMarker(ctx RowContext) string {
	gitMarker := " "
	if ctx.Item.GitStatus != "" {
		gitMarker = ctx.Item.GitStatus
	}

	if ctx.IsCursor {
		sStyle := ctx.Props.Styles.SelectedItem.UnsetPadding().UnsetWidth()
		if style, ok := ctx.Props.Styles.GitStyles[ctx.Item.GitStatus]; ok {
			return style.Inherit(sStyle).Render(gitMarker)
		}
		return sStyle.Render(gitMarker)
	}

	if style, ok := ctx.Props.Styles.GitStyles[ctx.Item.GitStatus]; ok {
		return style.Render(gitMarker)
	}
	return gitMarker
}

func renderPermIndicator(ctx RowContext) string {
	indicator := " "
	if !ctx.Item.CanWrite && !ctx.Item.IsUp && !ctx.Item.IsGhost {
		indicator = "!"
	}

	if ctx.IsCursor {
		return ctx.Props.Styles.SelectedItem.UnsetPadding().UnsetWidth().Render(indicator)
	}

	if indicator == "!" {
		return ctx.Props.Styles.DimCol.Render("!")
	}
	return indicator
}

func renderNamePart(ctx RowContext) string {
	var nameStr string
	if ctx.Item.IsUp {
		nameStr = ctx.Item.Name
	} else if ctx.Item.IsDir {
		nameStr = ctx.Item.Name + "/"
	} else {
		nameStr = ctx.Item.Name
	}

	nameStr = ui.Truncate(nameStr, ctx.Layout.NameWidth)

	// Git Status Coloring for Name
	nameStyle := ctx.Props.Styles.FileCol
	if ctx.Item.IsUp {
		nameStyle = ctx.Props.Styles.DimCol
	} else if ctx.Item.IsDir {
		nameStyle = ctx.Props.Styles.DirCol
	} else if ctx.Item.HasMetadata && ctx.Item.Mode&0o111 != 0 {
		nameStyle = ctx.Props.Styles.ExecCol
	}

	if ctx.Item.IsGhost {
		nameStyle = ctx.Props.Styles.GitGhost
	} else if style, ok := ctx.Props.Styles.GitStyles[ctx.Item.GitStatus]; ok {
		nameStyle = style
	}

	if !ctx.Item.CanRead && !ctx.Item.IsUp {
		nameStyle = ctx.Props.Styles.DimCol
	}

	if ctx.IsCursor {
		sStyle := ctx.Props.Styles.SelectedItem.UnsetPadding().UnsetWidth()
		styledName := nameStyle.Inherit(sStyle).Render(nameStr)
		gap := ctx.Layout.NameWidth - len(nameStr)
		if gap < 0 {
			gap = 0
		}
		return styledName + sStyle.Render(strings.Repeat(" ", gap))
	}

	styledName := nameStyle.Render(nameStr)
	gap := ctx.Layout.NameWidth - len(nameStr)
	if gap < 0 {
		gap = 0
	}
	return styledName + strings.Repeat(" ", gap)
}

func renderMetaPart(ctx RowContext) string {
	sStyle := ctx.Props.Styles.SelectedItem.UnsetPadding().UnsetWidth()

	if !ctx.Item.HasMetadata && !ctx.Item.IsUp && !ctx.Item.IsGhost {
		metaWidth := 0
		if ctx.Layout.ShowDate {
			metaWidth += ctx.Layout.DateWidth + ctx.Layout.ColumnGap
		}
		if ctx.Layout.ShowSize {
			metaWidth += ctx.Layout.SizeWidth + ctx.Layout.ColumnGap
		}

		if metaWidth == 0 {
			return ""
		}

		content := fmt.Sprintf("%*s", metaWidth, "...")
		if ctx.IsCursor {
			return ctx.Props.Styles.DimCol.Inherit(sStyle).Render(content)
		}
		return ctx.Props.Styles.DimCol.Render(content)
	}

	datePart := ""
	if ctx.Layout.ShowDate {
		dateStr := ctx.Item.FormattedDate
		content := fmt.Sprintf("%*s%*s", ctx.Layout.ColumnGap, "", ctx.Layout.DateWidth, dateStr)
		if ctx.IsCursor {
			datePart = ctx.Props.Styles.DimCol.Inherit(sStyle).Render(content)
		} else {
			datePart = ctx.Props.Styles.DimCol.Render(content)
		}
	}

	sizePart := ""
	if ctx.Layout.ShowSize {
		sizeStr := ctx.Item.FormattedSize
		content := fmt.Sprintf("%*s%*s", ctx.Layout.ColumnGap, "", ctx.Layout.SizeWidth, sizeStr)
		if ctx.IsCursor {
			sizePart = ctx.Props.Styles.DimCol.Inherit(sStyle).Render(content)
		} else {
			sizePart = ctx.Props.Styles.DimCol.Render(content)
		}
	}

	return datePart + sizePart
}
