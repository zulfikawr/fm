package tui

import (
	"strings"
	"testing"
	"time"

	"filemanager/internal/config"
	"filemanager/internal/files"
)

func TestRowRendering(t *testing.T) {
	m := NewModel("/")
	m.width = 80
	m.cfg = config.DefaultConfig()
	m.cfg.ShowSize = true
	m.cfg.ShowDateModified = true

	now := time.Date(2026, 1, 9, 15, 4, 0, 0, time.UTC)
	item := files.Item{
		Name:  "this_is_a_very_long_file_name_that_should_be_truncated_if_it_exceeds_available_space.txt",
		Size:  1234,
		MTime: now,
		IsDir: false,
	}

	t.Run("Truncation with all columns", func(t *testing.T) {
		m.width = 60
		row := m.renderRow(item, false)

		// Expected widths:
		// marker: 0
		// gitMarker: 2
		// columnGap: 2 * 2 = 4
		// date: 16
		// size: 10
		// Total non-name: 2 + 4 + 16 + 10 = 32
		// Available name: 60 - 32 = 28

		if !strings.Contains(row, "…") {
			t.Errorf("Expected row to contain ellipsis for long name, got: %s", row)
		}
	})

	t.Run("Back icon name", func(t *testing.T) {
		backItem := files.Item{Name: "↑ ..", IsUp: true, IsDir: true}
		row := m.renderRow(backItem, false)
		if !strings.Contains(row, "↑ ..") {
			t.Errorf("Expected row to contain ↑ .., got: %s", row)
		}
	})

	t.Run("No size column", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.ShowDateModified = true
		m.width = 60
		row := m.renderRow(item, false)

		if strings.Contains(row, "1.2 K") {
			t.Error("Did not expect size in row")
		}
		if !strings.Contains(row, "09/01/2026 15:04") {
			t.Error("Expected date in row")
		}
	})

	t.Run("No columns", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.ShowDateModified = false
		m.width = 40
		row := m.renderRow(item, false)

		if strings.Contains(row, "1.2 K") || strings.Contains(row, "09/01/2026 15:04") {
			t.Error("Did not expect size or date in row")
		}
		if !strings.Contains(row, "this_is_a_very_long_file_name") {
			t.Error("Expected name to be present")
		}
	})

	t.Run("Size formats", func(t *testing.T) {
		m.cfg.ShowSize = true
		m.cfg.SizeFormatIndex = 0 // Short
		row := m.renderRow(item, false)
		if !strings.Contains(row, "1.2 K") {
			t.Errorf("Expected 1.2 K, got %s", row)
		}

		m.cfg.SizeFormatIndex = 1 // Full
		row = m.renderRow(item, false)
		if !strings.Contains(row, "1.2 KB") {
			t.Errorf("Expected 1.2 KB, got %s", row)
		}

		m.cfg.SizeFormatIndex = 2 // Bytes
		row = m.renderRow(item, false)
		if !strings.Contains(row, "1234 B") {
			t.Errorf("Expected 1234 B, got %s", row)
		}
	})

	t.Run("Date formats", func(t *testing.T) {
		m.cfg.ShowDateModified = true

		m.cfg.DateFormatIndex = 0 // Default
		row := m.renderRow(item, false)
		if !strings.Contains(row, "09/01/2026 15:04") {
			t.Errorf("Expected Default date, got %s", row)
		}

		m.cfg.DateFormatIndex = 1 // ISO
		row = m.renderRow(item, false)
		if !strings.Contains(row, "2026-01-09 15:04") {
			t.Errorf("Expected ISO date, got %s", row)
		}
	})
}

func TestThemeApplication(t *testing.T) {
	for i, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			styles := NewStylesheet(theme)
			// Check some basic style properties
			if styles.Header.GetForeground() != theme.Dir {
				t.Errorf("Expected header foreground %v, got %v", theme.Dir, styles.Header.GetForeground())
			}

			m := NewModel("/")
			m.cfg.ThemeIndex = i
			m.styles = styles
			m.width = 80

			header := m.renderHeader()
			if header == "" {
				t.Error("Header should not be empty")
			}
		})
	}
}

func TestHeaderFooterRendering(t *testing.T) {
	m := NewModel("/home/user/docs")
	m.width = 100
	m.gitBranch = "main"

	t.Run("Header with git branch", func(t *testing.T) {
		header := m.renderHeader()
		if !strings.Contains(header, "docs") || !strings.Contains(header, "(main)") {
			t.Errorf("Header missing info: %s", header)
		}
	})

	t.Run("Footer item counter", func(t *testing.T) {
		m.filteredItems = []files.Item{
			{Name: "↑ ..", IsUp: true},
			{Name: "file1"},
			{Name: "file2"},
		}
		m.cursor = 1 // On file1
		footer := m.renderFooter()
		// Should show 1/2 (excluding ..)
		if !strings.Contains(footer, " 1/2 ") {
			t.Errorf("Expected footer to show 1/2, got: %s", footer)
		}
	})

	t.Run("Settings footer help", func(t *testing.T) {
		m.settingsOpen = true
		m.settingsCursor = 0
		footer := m.renderFooter()
		// Check for key parts of the help message to be more robust against ANSI/wrapping
		if !strings.Contains(footer, "Show/hide") || !strings.Contains(footer, "starting with") {
			t.Errorf("Expected hidden files help in footer, got: %s", footer)
		}

		m.settingsCursor = 10 // Theme
		footer = m.renderFooter()
		if !strings.Contains(footer, "Change") || !strings.Contains(footer, "color scheme") {
			t.Errorf("Expected theme help in footer, got: %s", footer)
		}
	})
}
