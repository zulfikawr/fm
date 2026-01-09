package tui

import (
	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleSettingsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	const numSettings = 11 // Hidden, CaseSensitive, Confirmations, Wrapping, Git, ShowSize, SizeFormat, ShowDate, DateFormat, ShowHeader, Theme

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", ".", "q":
			m.settingsOpen = false
			m.cfg.Save()
			return m, m.reload()
		case "up", "k":
			if m.settingsCursor > 0 {
				m.settingsCursor--
				// Skip disabled settings
				if (m.settingsCursor == 7 && !m.cfg.ShowSize) || (m.settingsCursor == 9 && !m.cfg.ShowDateModified) {
					if m.settingsCursor > 0 {
						m.settingsCursor--
					}
				}
			}
		case "down", "j":
			if m.settingsCursor < numSettings-1 {
				m.settingsCursor++
				// Skip disabled settings
				if (m.settingsCursor == 7 && !m.cfg.ShowSize) || (m.settingsCursor == 9 && !m.cfg.ShowDateModified) {
					if m.settingsCursor < numSettings-1 {
						m.settingsCursor++
					}
				}
			}
		case "enter", "right", "l", " ":
			m.toggleSetting(m.settingsCursor)
			m.cfg.Save()
		case "left", "h":
			m.toggleSettingPrev(m.settingsCursor)
			m.cfg.Save()
		}
	}
	return m, nil
}

func (m *Model) toggleSetting(idx int) {
	switch idx {
	case 0:
		m.cfg.ShowHidden = !m.cfg.ShowHidden
	case 1:
		m.cfg.CaseSensitive = !m.cfg.CaseSensitive
	case 2:
		m.cfg.ConfirmOperations = !m.cfg.ConfirmOperations
	case 3:
		m.cfg.WrapNavigation = !m.cfg.WrapNavigation
	case 4:
		m.cfg.ShowHeader = !m.cfg.ShowHeader
	case 5:
		m.cfg.EnableGit = !m.cfg.EnableGit
	case 6:
		m.cfg.ShowSize = !m.cfg.ShowSize
	case 7:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex + 1) % len(files.SizeFormats)
		}
	case 8:
		m.cfg.ShowDateModified = !m.cfg.ShowDateModified
	case 9:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex + 1) % len(files.DateFormats)
		}
	case 10:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex + 1) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	}
}

func (m *Model) toggleSettingPrev(idx int) {
	switch idx {
	case 7:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex - 1 + len(files.SizeFormats)) % len(files.SizeFormats)
		}
	case 9:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex - 1 + len(files.DateFormats)) % len(files.DateFormats)
		}
	case 10:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex - 1 + len(Themes)) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	default:
		m.toggleSetting(idx)
	}
}
