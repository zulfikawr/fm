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
		Name:  "long_file_name_that_should_be_truncated.txt",
		Size:  1234,
		MTime: now,
		IsDir: false,
	}

	t.Run("Truncation", func(t *testing.T) {
		m.width = 40
		row := m.renderRow(item, false)
		if !strings.Contains(row, "…") {
			t.Errorf("Expected row to contain ellipsis, got: %s", row)
		}
	})

	t.Run("Git status coloring", func(t *testing.T) {
		item.GitStatus = "M"
		row := m.renderRow(item, false)
		if row == "" {
			t.Error("Expected non-empty row")
		}

		item.GitStatus = "A"
		m.renderRow(item, false)
		item.GitStatus = "?"
		m.renderRow(item, false)
		item.GitStatus = "U"
		m.renderRow(item, false)
		item.GitStatus = "D"
		m.renderRow(item, false)
		item.GitStatus = "!"
		m.renderRow(item, false)
	})

	t.Run("Select mode markers", func(t *testing.T) {
		m.selectMode = true
		item.Selected = true
		m.renderRow(item, false)
		item.Selected = false
		m.renderRow(item, false)
	})
}

func TestViewRenderList(t *testing.T) {
	m := NewModel("/")
	m.width = 80
	m.height = 24
	m.items = []files.Item{
		{Name: "file1", Path: "/file1"},
	}
	m.applyFilter()

	t.Run("renderList with header", func(t *testing.T) {
		m.cfg.ShowHeader = true
		m.renderList("header", "footer")
	})

	t.Run("renderList without header", func(t *testing.T) {
		m.cfg.ShowHeader = false
		m.renderList("header", "footer")
	})
}
