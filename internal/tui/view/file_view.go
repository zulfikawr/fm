package view

import (
	"fm/internal/files/format"
	"fm/internal/tui/components/list"
	"fm/internal/tui/components/loading"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// RenderFileList maps ViewState to list.Props and renders the file list
func RenderFileList(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	viewportHeight := s.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	if s.UI.Loading && len(s.FilteredItems) == 0 {
		return loading.Render(loading.Props{
			Width:   s.Width,
			Height:  viewportHeight,
			Message: "Loading...",
			Spinner: s.LoadingSpinner,
			Styles:  styles,
		})
	}

	dateLayout := ""
	if s.Config.DateFormatIndex < len(format.DateFormats) {
		dateLayout = format.DateFormats[s.Config.DateFormatIndex].Layout
	}

	return list.Render(list.Props{
		Width:            s.Width,
		Height:           viewportHeight,
		Cursor:           s.Cursor,
		Offset:           s.Offset,
		Items:            s.FilteredItems,
		ShowHeader:       s.Config.ShowHeader,
		ShowSize:         s.Config.ShowSize,
		ShowDateModified: s.Config.ShowDateModified,
		SelectMode:       s.UI.SelectMode,
		SizeFormatIndex:  s.Config.SizeFormatIndex,
		DateFormatIndex:  s.Config.DateFormatIndex,
		DateLayout:       dateLayout,
		Styles:           styles,
	})
}
