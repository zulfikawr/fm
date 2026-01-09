package tui

import (
	"fmt"
	"strings"

	"filemanager/internal/files"

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
			},
		},
		{
			title:    "Display Options",
			startIdx: 5,
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
			startIdx: 11,
			settings: []setting{
				{"Theme", fmt.Sprintf("< %s >", Themes[m.cfg.ThemeIndex].Name)},
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
			if idx == 8 && !m.cfg.ShowSize { // Size Format
				inactive = true
			} else if idx == 10 && !m.cfg.ShowDateModified { // Date Format
				inactive = true
			}
			val := s.value
			if inactive {
				style = m.styles.DimCol.PaddingLeft(2)
				val = m.styles.DimCol.Render(s.value)
			}

			labelWidth := 25
			if m.width < 40 {
				labelWidth = m.width - 12
			}
			if labelWidth < 5 {
				labelWidth = 5
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
	for i := len(rows); i < viewportHeight; i++ {
		rows = append(rows, "")
	}

	return strings.Join(rows, "\n")
}
