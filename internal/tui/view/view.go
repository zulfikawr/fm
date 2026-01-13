package view

import (
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// Render assembles the full UI from components.
func Render(s *ViewState, styles theme.Stylesheet) string {
	if s.Width == 0 {
		return "Loading..."
	}

	// Memoize prompts if needed
	if s.UI.Confirming || s.UI.HostConfirm {
		memoizePrompts(s, styles)
	}

	headerStr := renderHeader(s, styles)
	footerStr := renderFooter(s, styles)
	bodyStr := renderBody(s, headerStr, footerStr, styles)

	return lipgloss.JoinVertical(lipgloss.Left, headerStr, bodyStr, footerStr)
}

// renderHeader builds props and renders header
func renderHeader(s *ViewState, styles theme.Stylesheet) string {
	return header.Render(header.Props{
		Width:           s.Width,
		Path:            s.Path,
		Separator:       s.Separator,
		GitBranch:       s.GitBranch,
		ReadOnly:        s.ReadOnly,
		TabCount:        len(s.Tabs),
		ActiveTab:       s.ActiveTab,
		SettingsOpen:    s.UI.SettingsOpen,
		Styles:          styles,
		RemoteConnected: s.RemoteConnected,
		RemoteUser:      s.RemoteUser,
		RemoteHost:      s.RemoteHost,
	})
}

// renderFooter builds props and renders footer
func renderFooter(s *ViewState, styles theme.Stylesheet) string {
	totalItems := len(s.FilteredItems)
	cursor := s.Cursor
	if totalItems > 0 && len(s.FilteredItems) > 0 && s.FilteredItems[0].IsUp {
		totalItems--
		cursor--
	}

	return footer.Render(footer.Props{
		Mode:            DetermineMode(s),
		Width:           s.Width,
		ProgressLabel:   s.Progress.Label,
		ProgressPercent: s.Progress.Percent,
		ActiveInput:     s.ActiveInput,
		AltMode:         s.AltMode,
		RemoteConnected: s.RemoteConnected,
		Message:         s.Msg,
		SortMode:        s.SortMode,
		Cursor:          cursor,
		TotalItems:      totalItems,
		SelectedCount:   s.SelectedCount,
		Items:           s.Items,
		FilteredItems:   s.FilteredItems,
		SettingsCursor:  s.SettingsCursor,
		ActionType:      s.ActionType,
		ClipboardCount:  len(s.Clipboard.Paths),
		ConflictDst:     s.ConflictDst,
		HostConfirmReq:  s.HostConfirmReq,
		Styles:          styles,
		PromptCache:     s.UI.PromptCache,
	})
}

// renderBody renders the main content area (List or Settings)
func renderBody(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	if s.UI.SettingsOpen {
		return RenderSettingsView(s, headerStr, footerStr, styles)
	}
	return RenderFileList(s, headerStr, footerStr, styles)
}
