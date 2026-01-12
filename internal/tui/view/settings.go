package view

import (
	"fmt"
	"strings"

	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// UpdateSettingsScroll calculates the correct scroll offset for the settings view
func UpdateSettingsScroll(m *state.Model) {
	s := GetViewState(m)
	styles := theme.GetStylesheet(s.Config.ThemeIndex)

	headerProps := header.Props{
		Width:        s.Width,
		Path:         s.Path,
		Separator:    s.Separator,
		GitBranch:    s.GitBranch,
		ReadOnly:     s.ReadOnly,
		TabCount:     len(s.Tabs),
		ActiveTab:    s.ActiveTab,
		SettingsOpen: s.UI.SettingsOpen,
		Styles:       styles,
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
	// Format: empty line, "File Operations", settings 0-5, empty line, "Display Options", settings 6-11, empty line, "Appearance", setting 12, empty line, "Keybindings", settings 13-28
	rowIndex := 0
	groupStartRow := 0 // Start of the group that contains current cursor

	if m.Settings.Cursor <= 5 {
		// File Operations group: empty line (0) + header (1) + cursor + 2
		rowIndex = m.Settings.Cursor + 2
		groupStartRow = 0 // Empty line before "File Operations"
	} else if m.Settings.Cursor <= 11 {
		// Display Options group: prev rows + empty line + header + cursor offset
		// 1 (empty) + 1 (header) + 6 (settings) + 1 (empty) + 1 (header) + cursor-6
		rowIndex = 10 + (m.Settings.Cursor - 6)
		groupStartRow = 8 // Empty line before "Display Options"
	} else if m.Settings.Cursor == 12 {
		// Appearance group: prev rows + empty line + header + cursor offset
		// Previous groups + 1 (empty) + 1 (header) + cursor-12
		rowIndex = 18 + (m.Settings.Cursor - 12)
		groupStartRow = 16 // Empty line before "Appearance"
	} else {
		// Keybindings group: prev rows + empty line + header + cursor offset
		// Previous groups + 1 (empty) + 1 (header) + cursor-13
		rowIndex = 21 + (m.Settings.Cursor - 13)
		groupStartRow = 19 // Empty line before "Keybindings"
	}

	// Scroll to keep cursor visible, ensuring group header is also visible when scrolling up
	if rowIndex < m.Settings.Offset {
		// When scrolling up, show the group header if possible
		m.Settings.Offset = groupStartRow
	} else if rowIndex >= m.Settings.Offset+viewportHeight {
		// When scrolling down, keep cursor at bottom of viewport
		m.Settings.Offset = rowIndex - viewportHeight + 1
	}

	if m.Settings.Offset < 0 {
		m.Settings.Offset = 0
	}
}

func RenderSettingsList(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	viewportHeight := s.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	type setting struct {
		label string
		value string
	}

	groups := []struct {
		title    string
		settings []setting
		startIdx int
	}{
		{
			title:    "File Operations",
			startIdx: 0,
			settings: []setting{
				{"Show Hidden Files", FormatBool(s, s.Config.ShowHidden, styles)},
				{"Case-Sensitive Search", FormatBool(s, s.Config.CaseSensitive, styles)},
				{"Confirm Operations", FormatBool(s, s.Config.ConfirmOperations, styles)},
				{"Wrap Navigation", FormatBool(s, s.Config.WrapNavigation, styles)},
				{"Preferred Editor", fmt.Sprintf("< %s >", ops.Editors[s.Config.EditorIndex])},
				{"Use Trash (Move to Trash)", FormatBool(s, s.Config.UseTrash, styles)},
			},
		},
		{
			title:    "Display Options",
			startIdx: 6,
			settings: []setting{
				{"Show Column Headers", FormatBool(s, s.Config.ShowHeader, styles)},
				{"Enable Git Status", FormatBool(s, s.Config.EnableGit, styles)},
				{"Show File Size", FormatBool(s, s.Config.ShowSize, styles)},
				{"Size Format", fmt.Sprintf("< %s >", format.SizeFormats[s.Config.SizeFormatIndex])},
				{"Show Date Modified", FormatBool(s, s.Config.ShowDateModified, styles)},
				{"Date Format", fmt.Sprintf("< %s >", format.DateFormats[s.Config.DateFormatIndex].Name)},
			},
		},
		{
			title:    "Appearance",
			startIdx: 12,
			settings: []setting{
				{"Theme", fmt.Sprintf("< %s >", theme.Themes[s.Config.ThemeIndex].Name)},
			},
		},
		{
			title:    "Keybindings",
			startIdx: 13,
			settings: []setting{
				{"Open", "Enter/→/l"},
				{"Back", "Backspace/←/h"},
				{"Select", "Space"},
				{"New Tab", "Alt+T"},
				{"Switch Tab", "Alt+1-9"},
				{"Sort", "s"},
				{"Search", "/"},
				{"Go to Path", "g"},
				{"Copy", "c"},
				{"Cut", "x"},
				{"Paste", "v"},
				{"Rename", "r"},
				{"Delete", "d"},
				{"Clear/Esc", "Esc"},
				{"Settings", "."},
				{"Quit", "q"},
			},
		},
	}

	var rows []string
	rows = append(rows, "") // Add a line above the first group
	for i, g := range groups {
		if i > 0 {
			rows = append(rows, "")
		}

		rows = append(rows, styles.SettingsHeader.Width(s.Width).Render(g.title))
		for j, s_item := range g.settings {
			idx := g.startIdx + j
			style := styles.SettingsItem
			if idx == s.SettingsCursor {
				style = styles.SettingsSelectedItem
			}

			// Dim inactive settings
			inactive := false
			if idx == 9 && !s.Config.ShowSize { // Size Format
				inactive = true
			} else if idx == 11 && !s.Config.ShowDateModified { // Date Format
				inactive = true
			}
			val := s_item.value
			if inactive {
				style = styles.DimCol.PaddingLeft(2)
				val = styles.DimCol.Render(s_item.value)
			}

			labelWidth := 35
			if s.Width < 60 {
				labelWidth = s.Width - 20
			}
			if labelWidth < 10 {
				labelWidth = 10
			}

			label := s_item.label + ":"
			if len(label) > labelWidth {
				label = label[:labelWidth-1] + "…"
			}

			// Calculate width without ANSI codes for proper alignment
			valWidth := lipgloss.Width(s_item.value)
			content := fmt.Sprintf("% -*s %s", labelWidth, label, val)

			if labelWidth+1+valWidth > s.Width-2 {
				// If still too long, prioritize showing the value
				availableLabelWidth := s.Width - 2 - valWidth - 1
				if availableLabelWidth > 0 {
					label = s_item.label + ":"
					if len(label) > availableLabelWidth {
						label = label[:availableLabelWidth-1] + "…"
					}
					content = fmt.Sprintf("% -*s %s", availableLabelWidth, label, val)
				}
			}
			rows = append(rows, style.Width(s.Width).Render(content))
		}
	}
	// Apply scroll offset
	if s.SettingsOffset > 0 && s.SettingsOffset < len(rows) {
		rows = rows[s.SettingsOffset:]
	} else if s.SettingsOffset >= len(rows) {
		rows = []string{}
	}

	// Ensure we fill the viewport
	if len(rows) > viewportHeight {
		rows = rows[:viewportHeight]
	} else {
		for i := len(rows); i < viewportHeight; i++ {
			rows = append(rows, "")
		}
	}

	return strings.Join(rows, "\n")
}
