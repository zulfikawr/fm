package tui

import (
	"fmt"
	"strings"

	"fm/internal/files"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderSettingsList(header, footer string) string {
	viewportHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
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
				{"Show Hidden Files", m.formatBool(m.cfg.ShowHidden)},
				{"Case-Sensitive Search", m.formatBool(m.cfg.CaseSensitive)},
				{"Confirm Operations", m.formatBool(m.cfg.ConfirmOperations)},
				{"Wrap Navigation", m.formatBool(m.cfg.WrapNavigation)},
				{"Preferred Editor", fmt.Sprintf("< %s >", files.Editors[m.cfg.EditorIndex])},
				{"Use Trash (Move to Trash)", m.formatBool(m.cfg.UseTrash)},
			},
		},
		{
			title:    "Display Options",
			startIdx: 6,
			settings: []setting{
				{"Show Column Headers", m.formatBool(m.cfg.ShowHeader)},
				{"Enable Git Status", m.formatBool(m.cfg.EnableGit)},
				{"Show File Size", m.formatBool(m.cfg.ShowSize)},
				{"Size Format", fmt.Sprintf("< %s >", files.SizeFormats[m.cfg.SizeFormatIndex])},
				{"Show Date Modified", m.formatBool(m.cfg.ShowDateModified)},
				{"Date Format", fmt.Sprintf("< %s >", files.DateFormats[m.cfg.DateFormatIndex].Name)},
			},
		},
		{
			title:    "Appearance",
			startIdx: 12,
			settings: []setting{
				{"Theme", fmt.Sprintf("< %s >", Themes[m.cfg.ThemeIndex].Name)},
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
		rows = append(rows, m.styles.SettingsHeader.Width(m.width).Render(g.title))
		for j, s := range g.settings {
			idx := g.startIdx + j
			style := m.styles.SettingsItem
			if idx == m.settingsCursor {
				style = m.styles.SettingsSelectedItem
			}

			// Dim inactive settings
			inactive := false
			if idx == 9 && !m.cfg.ShowSize { // Size Format
				inactive = true
			} else if idx == 11 && !m.cfg.ShowDateModified { // Date Format
				inactive = true
			}
			val := s.value
			if inactive {
				style = m.styles.DimCol.PaddingLeft(2)
				val = m.styles.DimCol.Render(s.value)
			}

			labelWidth := 35
			if m.width < 60 {
				labelWidth = m.width - 20
			}
			if labelWidth < 10 {
				labelWidth = 10
			}

			label := s.label + ":"
			if len(label) > labelWidth {
				label = label[:labelWidth-1] + "…"
			}

			// Calculate width without ANSI codes for proper alignment
			valWidth := lipgloss.Width(s.value)
			content := fmt.Sprintf("%-*s %s", labelWidth, label, val)

			if labelWidth+1+valWidth > m.width-2 {
				// If still too long, prioritize showing the value
				availableLabelWidth := m.width - 2 - valWidth - 1
				if availableLabelWidth > 0 {
					label = s.label + ":"
					if len(label) > availableLabelWidth {
						label = label[:availableLabelWidth-1] + "…"
					}
					content = fmt.Sprintf("%-*s %s", availableLabelWidth, label, val)
				}
			}

			rows = append(rows, style.Width(m.width).Render(content))
		}
	}

	// Apply scroll offset
	if m.settingsOffset > 0 && m.settingsOffset < len(rows) {
		rows = rows[m.settingsOffset:]
	} else if m.settingsOffset >= len(rows) {
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
