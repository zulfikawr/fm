package integration_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/integration"
	"github.com/zulfikawr/fm/internal/tui/messages"
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

	msg := messages.SearchMsg{
		Query:   "test",
		Results: results,
	}

	integration.HandleSearch(m, msg)

	if len(m.Search.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(m.Search.Results))
	}
}
