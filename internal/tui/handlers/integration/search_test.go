package integration_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestSearch_Msg(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Search.Query = "test"

	results := []core.FileResult{
		{
			Path: "/test/file1.txt",
			Matches: []core.Match{
				{Line: 1, Content: "test content"},
			},
		},
	}

	msg := messages.SearchMsg{
		Query:   "test",
		Results: results,
	}

	integration.HandleSearch(m, msg)

	if len(m.Navigation.Search.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(m.Navigation.Search.Results))
	}
}

func TestSearch_Keys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Inputs.Mode = tuictx.InputFuzzySearch
	m.Navigation.Search.Results = []core.FileResult{
		{FileName: "f1", Matches: []core.Match{{Line: 1}, {Line: 2}}},
		{FileName: "f2", Matches: []core.Match{{Line: 3}}},
	}

	t.Run("Move Cursor", func(t *testing.T) {
		integration.HandleSearch(m, tea.KeyMsg{Type: tea.KeyDown})
		if m.Navigation.Search.CursorMatch != 1 {
			t.Errorf("expected CursorMatch 1, got %d", m.Navigation.Search.CursorMatch)
		}
	})

	t.Run("Tab Collapse", func(t *testing.T) {
		integration.HandleSearch(m, tea.KeyMsg{Type: tea.KeyTab})
		if !m.Navigation.Search.Results[0].Collapsed {
			t.Error("expected result to be collapsed")
		}
	})
}

func TestSearch_TriggerStop(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Trigger", func(t *testing.T) {
		integration.TriggerSearch(m, "query")
		if !m.Navigation.Search.IsSearching {
			t.Error("expected IsSearching to be true")
		}
	})

	t.Run("Stop", func(t *testing.T) {
		integration.StopSearch(m)
		if m.Navigation.Search.IsSearching {
			t.Error("expected IsSearching to be false")
		}
	})
}
