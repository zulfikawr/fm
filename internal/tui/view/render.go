package view

import (
	"fm/internal/tui/components/file"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/components/loading"
	"fm/internal/tui/components/views"
	"fm/internal/tui/context"

	"github.com/charmbracelet/lipgloss"
)

// Render assembles the full UI from components
func Render(m *context.Model) string {
	if m.Display.Width == 0 || m.Display.Height == 0 {
		return "Initializing..."
	}

	// Use cached layout or calculate if needed
	if m.Display.Layout.Width != m.Display.Width || m.Display.Layout.Height != m.Display.Height {
		m.Display.Layout = CalculateLayout(m)
	}
	layout := m.Display.Layout
	styles := m.Display.Styles

	// 1. Render Header
	headerStr := header.Render(header.Props{
		Width:         layout.Width,
		Path:          m.Navigation.Path,
		Separator:     m.FS.Separator(),
		RemoteStr:     formatRemoteStr(m),
		GitBranch:     m.Git.Branch,
		ReadOnly:      m.Display.ReadOnly,
		TabCount:      len(m.Tabs),
		ActiveTab:     m.ActiveTab,
		SettingsOpen:  m.UI.SettingsOpen,
		LogOpen:       m.UI.LogOpen,
		ClipboardOpen: m.UI.ClipboardOpen,
		Style:         styles,
	})

	// 2. Render Body
	var bodyStr string
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
		})
	}

	// 3. Render Footer
	footerStr := footer.Render(footer.Props{
		Mode:                 determineFooterMode(m),
		Width:                layout.Width,
		ProgressLabel:        m.Operations.Progress.Label,
		ProgressPercent:      m.Operations.Progress.Percent,
		ActiveInput:          m.Inputs.ActiveInput,
		AltMode:              m.Inputs.AltMode,
		RemoteConnected:      !m.FS.IsLocal(),
		Message:              m.Message.Text,
		SortMode:             m.Display.SortMode,
		Cursor:               m.Navigation.Cursor,
		TotalItems:           len(m.Navigation.FilteredItems),
		SelectedCount:        m.Navigation.SelectedCount,
		Items:                m.Navigation.Items,
		FilteredItems:        m.Navigation.FilteredItems,
		SettingsCursor:       m.Settings.Cursor,
		ActionType:           m.Operations.ActionType,
		ClipboardCount:       len(m.Operations.Clipboard.Paths),
		ConflictDst:          m.Operations.Conflict.Destination,
		ConflictPendingCount: len(m.Operations.Conflict.PendingItems),
		HostConfirmReq:       m.Remote.HostConfirmReq,
		Styles:               styles,
		PromptCache:          m.UI.PromptCache,
	})

	// Ensure body takes full available height to push footer to bottom
	bodyStr = lipgloss.NewStyle().Height(layout.BodyHeight).MaxHeight(layout.BodyHeight).Render(bodyStr)

	return lipgloss.JoinVertical(lipgloss.Left, headerStr, bodyStr, footerStr)
}

func determineFooterMode(m *context.Model) footer.Mode {
	if m.UI.InputActive {
		switch m.Inputs.Mode {
		case context.InputSearch:
			return footer.ModeSearching
		case context.InputRename:
			return footer.ModeRenaming
		case context.InputGoto:
			return footer.ModeGoto
		case context.InputAuth:
			return footer.ModeAuth
		case context.InputFuzzySearch:
			return footer.ModeFuzzySearch
		case context.InputZip:
			return footer.ModeZip
		case context.InputUnzip:
			return footer.ModeUnzip
		}
	}
	if m.UI.HostConfirm {
		return footer.ModeHostConfirm
	}
	if m.UI.Confirming {
		return footer.ModeConfirming
	}
	if m.Operations.Progress.Visible {
		return footer.ModeProgress
	}
	if m.Message.Text != "" {
		return footer.ModeMessage
	}
	if m.UI.SettingsOpen {
		return footer.ModeSettings
	}
	if m.UI.LogOpen {
		return footer.ModeLog
	}
	if m.UI.ClipboardOpen {
		return footer.ModeClipboard
	}
	return footer.ModeNormal
}

func formatRemoteStr(m *context.Model) string {
	if m.FS.IsLocal() {
		return ""
	}
	user := m.FS.User()
	addr := m.FS.Address()

	if user != "" && addr != "" {
		return user + "@" + addr
	}
	if addr != "" {
		return addr
	}
	return "Remote"
}
