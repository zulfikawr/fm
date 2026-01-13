package view

import (
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/files/core"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

func TestRenderFileList(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])
	s := &ViewState{
		Width:  80,
		Height: 24,
		UI:     &state.UIState{},
		FilteredItems: []core.Item{
			{Name: "f1"},
		},
		Config: &config.Config{
			ShowHeader: true,
		},
	}

	res := RenderFileList(s, "H", "F", styles)
	if !strings.Contains(res, "f1") {
		t.Error("Expected f1 in list")
	}
	if !strings.Contains(res, "Name") {
		t.Error("Expected header in list")
	}
}
