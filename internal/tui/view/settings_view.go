package view

import (
	"fmt"

	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/components/settings"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// RenderSettingsView maps ViewState to settings.Props and renders the settings list
func RenderSettingsView(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	viewportHeight := s.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	groups := []settings.SettingGroup{
		{
			Title: "File Operations",
			Settings: []settings.SettingItem{
				{Label: "Show Hidden Files", Value: settings.FormatBool(s.Config.ShowHidden, styles)},
				{Label: "Case-Sensitive Search", Value: settings.FormatBool(s.Config.CaseSensitive, styles)},
				{Label: "Confirm Operations", Value: settings.FormatBool(s.Config.ConfirmOperations, styles)},
				{Label: "Wrap Navigation", Value: settings.FormatBool(s.Config.WrapNavigation, styles)},
				{Label: "Preferred Editor", Value: fmt.Sprintf("< %s >", ops.Editors[s.Config.EditorIndex])},
				{Label: "Use Trash (Move to Trash)", Value: settings.FormatBool(s.Config.UseTrash, styles)},
			},
		},
		{
			Title: "Display Options",
			Settings: []settings.SettingItem{
				{Label: "Show Column Headers", Value: settings.FormatBool(s.Config.ShowHeader, styles)},
				{Label: "Enable Git Status", Value: settings.FormatBool(s.Config.EnableGit, styles)},
				{Label: "Show File Size", Value: settings.FormatBool(s.Config.ShowSize, styles)},
				{Label: "Size Format", Value: fmt.Sprintf("< %s >", format.SizeFormats[s.Config.SizeFormatIndex])},
				{Label: "Show Date Modified", Value: settings.FormatBool(s.Config.ShowDateModified, styles)},
				{Label: "Date Format", Value: fmt.Sprintf("< %s >", format.DateFormats[s.Config.DateFormatIndex].Name)},
			},
		},
		{
			Title: "Appearance",
			Settings: []settings.SettingItem{
				{Label: "Theme", Value: fmt.Sprintf("< %s >", theme.Themes[s.Config.ThemeIndex].Name)},
			},
		},
		{
			Title: "Keybindings",
			Settings: []settings.SettingItem{
				{Label: "Open", Value: "Enter/→/l"},
				{Label: "Back", Value: "Backspace/←/h"},
				{Label: "Select", Value: "Space"},
				{Label: "New Tab", Value: "Alt+T"},
				{Label: "Switch Tab", Value: "Alt+1-9"},
				{Label: "Sort", Value: "s"},
				{Label: "Search", Value: "/"},
				{Label: "Go to Path", Value: "g"},
				{Label: "Copy", Value: "c"},
				{Label: "Cut", Value: "x"},
				{Label: "Paste", Value: "v"},
				{Label: "Rename", Value: "r"},
				{Label: "Delete", Value: "d"},
				{Label: "Clear/Esc", Value: "Esc"},
				{Label: "Settings", Value: "."},
				{Label: "Quit", Value: "q"},
			},
		},
	}

	// Mark inactive settings
	if !s.Config.ShowSize {
		groups[1].Settings[3].Inactive = true
	}
	if !s.Config.ShowDateModified {
		groups[1].Settings[5].Inactive = true
	}

	return settings.Render(settings.Props{
		Width:  s.Width,
		Height: viewportHeight,
		Cursor: s.SettingsCursor,
		Offset: s.SettingsOffset,
		Groups: groups,
		Styles: styles,
	})
}

// UpdateSettingsScroll calculates the correct scroll offset for the settings view
func UpdateSettingsScroll(m *state.Model) {
	s := GetViewState(m)
	styles := theme.NewStylesheet(theme.Themes[s.Config.ThemeIndex])

	headerProps := header.Props{
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
	}
	headerStr := header.Render(headerProps)

	footerMode := DetermineMode(&s)
	footerProps := footer.Props{
		Mode:           footerMode,
		Width:          s.Width,
		SettingsCursor: s.SettingsCursor,
		Styles:         styles,
	}
	footerStr := footer.Render(footerProps)

	viewportHeight := s.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	// Map cursor to actual row index (accounting for group headers and spacing)
	rowIndex := 0
	groupStartRow := 0 // Start of the group that contains current cursor

	if m.Settings.Cursor <= 5 {
		rowIndex = m.Settings.Cursor + 2
		groupStartRow = 0
	} else if m.Settings.Cursor <= 11 {
		rowIndex = 10 + (m.Settings.Cursor - 6)
		groupStartRow = 8
	} else if m.Settings.Cursor == 12 {
		rowIndex = 18 + (m.Settings.Cursor - 12)
		groupStartRow = 16
	} else {
		rowIndex = 21 + (m.Settings.Cursor - 13)
		groupStartRow = 19
	}

	if rowIndex < m.Settings.Offset {
		m.Settings.Offset = groupStartRow
	} else if rowIndex >= m.Settings.Offset+viewportHeight {
		m.Settings.Offset = rowIndex - viewportHeight + 1
	}

	if m.Settings.Offset < 0 {
		m.Settings.Offset = 0
	}
}
