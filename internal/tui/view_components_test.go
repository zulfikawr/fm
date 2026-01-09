package tui

import (
	"strings"
	"testing"

	"filemanager/internal/files"
)

func TestViewComponents(t *testing.T) {
	m := NewModel("/home/user/docs")
	m.width = 100
	m.gitBranch = "main"

	t.Run("Header Rendering", func(t *testing.T) {
		header := m.renderHeader()
		if !strings.Contains(header, "docs") || !strings.Contains(header, "(main)") {
			t.Errorf("Header missing info: %s", header)
		}
	})

	t.Run("Footer Rendering", func(t *testing.T) {
		m.filteredItems = []files.Item{
			{Name: "↑ ..", IsUp: true},
			{Name: "file1"},
		}
		m.cursor = 1
		footer := m.renderFooter()
		if !strings.Contains(footer, " 1/1 ") {
			t.Errorf("Expected footer to show 1/1, got: %s", footer)
		}
	})

	t.Run("Colorize Keys", func(t *testing.T) {
		s := "[k] Key"
		res := m.colorizeKeys(s)
		if !strings.Contains(res, "k") {
			t.Error("Expected k to be present in colorized string")
		}
	})

	t.Run("Footer Edge Cases", func(t *testing.T) {
		m.width = 40
		m.renderFooter()

		m.searching = true
		m.renderFooter()
		m.searching = false

		m.renaming = true
		m.renderFooter()
		m.renaming = false

		m.confirming = true
		m.actionType = "delete"
		m.renderFooter()
	})
}
