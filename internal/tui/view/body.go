package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/file"
	"github.com/zulfikawr/fm/internal/tui/components/loading"
	"github.com/zulfikawr/fm/internal/tui/components/views"
	"github.com/zulfikawr/fm/internal/tui/context"

	"github.com/charmbracelet/lipgloss"
)

func renderBody(m *context.Model, layout context.Layout) string {
	var bodyStr string
	styles := m.Display.Styles

	// Skeleton-First logic: Only show full screen loading if we have NO items to show
	if m.UI.Loading && len(m.Navigation.FilteredItems) == 0 {
		bodyStr = loading.Render(loading.Props{
			Width:   layout.Width,
			Height:  layout.BodyHeight,
			Message: "Loading...",
			Spinner: m.Display.LoadingSpinner,
			Style:   styles,
		})
	} else if m.UI.SettingsOpen {
		bodyStr = views.RenderSettings(views.SettingsProps{
			Width:  layout.Width,
			Height: layout.BodyHeight,
			Cursor: m.Settings.Cursor,
			Offset: m.Settings.Offset,
			Config: m.Config,
			Style:  styles,
		})
	} else if m.UI.HelpOpen {
		bodyStr = views.RenderHelp(views.HelpProps{
			Width:  layout.Width,
			Height: layout.BodyHeight,
			Cursor: m.Help.Cursor,
			Offset: m.Help.Offset,
			Style:  styles,
		})
	} else if m.UI.LogOpen {
		bodyStr = views.RenderLogs(views.LogsProps{
			Width:  layout.Width,
			Height: layout.BodyHeight,
			Cursor: m.Logs.Cursor,
			Offset: m.Logs.Offset,
			Logs:   m.Logs.Entries,
			Style:  styles,
		})
	} else if m.UI.ClipboardOpen {
		bodyStr = views.RenderClipboard(views.ClipboardProps{
			Width:    layout.Width,
			Height:   layout.BodyHeight,
			Cursor:   m.Operations.Clipboard.Cursor,
			Offset:   m.Operations.Clipboard.Offset,
			Paths:    m.Operations.Clipboard.Paths,
			SourceFS: m.Operations.Clipboard.SourceFS,
			IsCut:    m.Operations.Clipboard.IsCut,
			Style:    styles,
		})
	} else if m.Inputs.Mode == context.InputFuzzySearch || len(m.Search.Results) > 0 {
		bodyStr = views.RenderSearch(views.SearchProps{
			Width:       layout.Width,
			Height:      layout.BodyHeight,
			Query:       m.Search.Query,
			Results:     m.Search.Results,
			IsSearching: m.Search.IsSearching,
			CursorFile:  m.Search.CursorFile,
			CursorMatch: m.Search.CursorMatch,
			Offset:      m.Search.Offset,
			Spinner:     m.Display.LoadingSpinner,
			Style:       styles,
			EnableIcons: m.Config.EnableIcons,
		})
	} else {
		bodyStr = file.Render(file.Props{
			Width:            layout.Width,
			Height:           layout.BodyHeight,
			Cursor:           m.Navigation.Cursor,
			Offset:           m.Navigation.Offset,
			Items:            m.Navigation.FilteredItems,
			ShowHeader:       m.Config.ShowHeader,
			ShowSize:         m.Config.ShowSize,
			ShowDateModified: m.Config.ShowDateModified,
			SelectMode:       m.UI.SelectMode,
			SizeFormatIndex:  m.Config.SizeFormatIndex,
			DateFormatIndex:  m.Config.DateFormatIndex,
			Styles:           styles,
			SelectedPaths:    m.Navigation.SelectedPaths,
			EnableIcons:      m.Config.EnableIcons,
		})
	}

	// Ensure body takes full available height to push footer to bottom
	return lipgloss.NewStyle().Height(layout.BodyHeight).MaxHeight(layout.BodyHeight).Render(bodyStr)
}
