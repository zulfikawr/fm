package view

import (
	"strings"
	"testing"
	"time"

	"fm/internal/config"
	"fm/internal/files"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestRowRendering(t *testing.T) {
	s := &ViewState{
		Width: 80,
		Config: &config.Config{
			ShowSize:         true,
			ShowDateModified: true,
		},
		UI: &state.UIState{},
	}
	styles := theme.GetStylesheet(0)

	now := time.Date(2026, 1, 9, 15, 4, 0, 0, time.UTC)
	item := files.Item{
		Name:     "long_file_name_that_should_be_truncated.txt",
		Size:     1234,
		MTime:    now,
		IsDir:    false,
		CanRead:  true,
		CanWrite: true,
	}

	t.Run("Truncation", func(t *testing.T) {
		s.Width = 40
		layout := calculateLayout(s, 20)
		row := renderRow(s, item, false, styles, layout)
		if !strings.Contains(row, "…") {
			t.Errorf("Expected row to contain ellipsis, got: %s", row)
		}
	})

	t.Run("Git status coloring", func(t *testing.T) {
		layout := calculateLayout(s, 20)
		item.GitStatus = "M"
		row := renderRow(s, item, false, styles, layout)
		if row == "" {
			t.Error("Expected non-empty row")
		}

		item.GitStatus = "A"
		renderRow(s, item, false, styles, layout)
		item.GitStatus = "?"
		renderRow(s, item, false, styles, layout)
		item.GitStatus = "U"
		renderRow(s, item, false, styles, layout)
		item.GitStatus = "D"
		renderRow(s, item, false, styles, layout)
		item.GitStatus = "!"
		renderRow(s, item, false, styles, layout)
	})

	t.Run("Select mode markers", func(t *testing.T) {
		layout := calculateLayout(s, 20)
		s.UI.SelectMode = true
		item.Selected = true
		renderRow(s, item, false, styles, layout)
		item.Selected = false
		renderRow(s, item, false, styles, layout)
	})
}

func TestViewRenderList(t *testing.T) {
	s := &ViewState{
		Width:  80,
		Height: 24,
		Config: &config.Config{},
		UI:     &state.UIState{},
		Items: []files.Item{
			{Name: "file1", Path: "/file1"},
		},
		FilteredItems: []files.Item{
			{Name: "file1", Path: "/file1"},
		},
	}
	styles := theme.GetStylesheet(0)

	t.Run("renderList with header", func(t *testing.T) {
		s.Config.ShowHeader = true
		RenderList(s, "header", "footer", styles)
	})

	t.Run("renderList without header", func(t *testing.T) {
		s.Config.ShowHeader = false
		RenderList(s, "header", "footer", styles)
	})
}
