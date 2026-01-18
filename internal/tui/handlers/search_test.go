package handlers

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"
)

func TestSearch_Msg(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Search.Query = "test"

	results := []core.FileResult{
		{
			Path: "/test/file1.txt",
			Matches: []core.Match{
				{Line: 1, Content: "test content"},
			},
		},
	}

	msg := SearchMsg{
		Query:   "test",
		Results: results,
	}

	HandleSearch(m, msg)

	if len(m.Search.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(m.Search.Results))
	}
}

func TestSearch_Keys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Inputs.Mode = tuictx.InputFuzzySearch
	m.Search.Results = []core.FileResult{
		{
			Path: "/test/file1.txt",
			Matches: []core.Match{
				{Line: 1, Content: "match 1"},
				{Line: 2, Content: "match 2"},
			},
		},
	}
	m.Search.CursorFile = 0
	m.Search.CursorMatch = 0

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Move down to next match with 'alt+j'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j"), Alt: true})
	time.Sleep(10 * time.Millisecond)
	if m.Search.CursorMatch != 1 {
		t.Errorf("expected match cursor at 1 with 'alt+j', got %d", m.Search.CursorMatch)
	}

	// Move up to previous match with 'alt+k'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k"), Alt: true})
	time.Sleep(10 * time.Millisecond)
	if m.Search.CursorMatch != 0 {
		t.Errorf("expected match cursor at 0 with 'alt+k', got %d", m.Search.CursorMatch)
	}

	// Add another file for alt+n/m test
	m.Search.Results = append(m.Search.Results, core.FileResult{
		Path: "/test/file2.txt",
		Matches: []core.Match{
			{Line: 1, Content: "file 2 match"},
		},
	})

	// Move to next file with 'alt+n'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n"), Alt: true})
	time.Sleep(10 * time.Millisecond)
	if m.Search.CursorFile != 1 {
		t.Errorf("expected file cursor at 1 with 'alt+n', got %d", m.Search.CursorFile)
	}

	// Move back to first file with 'alt+m'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	time.Sleep(10 * time.Millisecond)
	if m.Search.CursorFile != 0 {
		t.Errorf("expected file cursor at 0 with 'alt+m', got %d", m.Search.CursorFile)
	}

	// Collapse results
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(10 * time.Millisecond)
	if !m.Search.Results[0].Collapsed {
		t.Error("expected results to be collapsed after tab")
	}

	_ = tm.Quit()
}
