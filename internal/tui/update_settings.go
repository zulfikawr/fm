package tui

import (
	"fm/internal/config"
	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) handleSettingsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	const numSettings = 28 // 13 settings + 15 keybindings

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", ".", "q":
			m.settingsOpen = false
			m.cfg.Save()
			return m, m.reload()
		case "r":
			// Trigger reset confirmation
			m.confirming = true
			m.actionType = "reset-settings"
			return m, nil
		case "y":
			// Confirm reset
			if m.confirming && m.actionType == "reset-settings" {
				m.cfg = config.DefaultConfig()
				m.cfg.Save()
				m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
				m.updateThemeStyles()
				m.confirming = false
				m.actionType = ""
				return m, tea.Batch(m.setMsg("Settings reset to defaults"), m.reload())
			}
		case "n":
			// Cancel reset
			if m.confirming && m.actionType == "reset-settings" {
				m.confirming = false
				m.actionType = ""
				return m, nil
			}
		case "up", "k":
			if m.confirming {
				return m, nil
			}
			if m.settingsCursor > 0 {
				m.settingsCursor--
				// Skip disabled settings
				if (m.settingsCursor == 9 && !m.cfg.ShowSize) || (m.settingsCursor == 11 && !m.cfg.ShowDateModified) {
					if m.settingsCursor > 0 {
						m.settingsCursor--
					}
				}
				// Update scroll offset
				m.updateSettingsScroll()
			}
		case "down", "j":
			if m.confirming {
				return m, nil
			}
			if m.settingsCursor < numSettings-1 {
				m.settingsCursor++
				// Skip disabled settings
				if (m.settingsCursor == 9 && !m.cfg.ShowSize) || (m.settingsCursor == 11 && !m.cfg.ShowDateModified) {
					if m.settingsCursor < numSettings-1 {
						m.settingsCursor++
					}
				}
				// Update scroll offset
				m.updateSettingsScroll()
			}
		case "enter", "right", "l", " ":
			if m.confirming {
				return m, nil
			}
			m.toggleSetting(m.settingsCursor)
			m.cfg.Save()
		case "left", "h":
			if m.confirming {
				return m, nil
			}
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
		m.cfg.EditorIndex = (m.cfg.EditorIndex + 1) % len(files.Editors)
	case 5:
		m.cfg.UseTrash = !m.cfg.UseTrash
	case 6:
		m.cfg.ShowHeader = !m.cfg.ShowHeader
	case 7:
		m.cfg.EnableGit = !m.cfg.EnableGit
	case 8:
		m.cfg.ShowSize = !m.cfg.ShowSize
	case 9:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex + 1) % len(files.SizeFormats)
		}
	case 10:
		m.cfg.ShowDateModified = !m.cfg.ShowDateModified
	case 11:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex + 1) % len(files.DateFormats)
		}
	case 12:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex + 1) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	}
}

// updateSettingsScroll adjusts the settings scroll offset to keep cursor visible
func (m *Model) updateSettingsScroll() {
	header := m.renderHeader()
	footer := m.renderFooter()
	viewportHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	// Map cursor to actual row index (accounting for group headers and spacing)
	// Format: empty line, "File Operations", settings 0-5, empty line, "Display Options", settings 6-11, empty line, "Appearance", setting 12
	rowIndex := 0
	groupStartRow := 0 // Start of the group that contains current cursor

	if m.settingsCursor <= 5 {
		// File Operations group: empty line (0) + header (1) + cursor + 2
		rowIndex = m.settingsCursor + 2
		groupStartRow = 0 // Empty line before "File Operations"
	} else if m.settingsCursor <= 11 {
		// Display Options group: prev rows + empty line + header + cursor offset
		// 1 (empty) + 1 (header) + 6 (settings) + 1 (empty) + 1 (header) + cursor-6
		rowIndex = 10 + (m.settingsCursor - 6)
		groupStartRow = 8 // Empty line before "Display Options"
	} else if m.settingsCursor == 12 {
		// Appearance group: prev rows + empty line + header + cursor offset
		// Previous groups + 1 (empty) + 1 (header) + cursor-12
		rowIndex = 18 + (m.settingsCursor - 12)
		groupStartRow = 16 // Empty line before "Appearance"
	} else {
		// Keybindings group: prev rows + empty line + header + cursor offset
		// Previous groups + 1 (empty) + 1 (header) + cursor-13
		rowIndex = 21 + (m.settingsCursor - 13)
		groupStartRow = 19 // Empty line before "Keybindings"
	}

	// Scroll to keep cursor visible, ensuring group header is also visible when scrolling up
	if rowIndex < m.settingsOffset {
		// When scrolling up, show the group header if possible
		m.settingsOffset = groupStartRow
	} else if rowIndex >= m.settingsOffset+viewportHeight {
		// When scrolling down, keep cursor at bottom of viewport
		m.settingsOffset = rowIndex - viewportHeight + 1
	}

	if m.settingsOffset < 0 {
		m.settingsOffset = 0
	}
}

func (m *Model) toggleSettingPrev(idx int) {
	switch idx {
	case 4:
		m.cfg.EditorIndex = (m.cfg.EditorIndex - 1 + len(files.Editors)) % len(files.Editors)
	case 9:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex - 1 + len(files.SizeFormats)) % len(files.SizeFormats)
		}
	case 11:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex - 1 + len(files.DateFormats)) % len(files.DateFormats)
		}
	case 12:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex - 1 + len(Themes)) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	default:
		m.toggleSetting(idx)
	}
}
