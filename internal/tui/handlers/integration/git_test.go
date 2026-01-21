package integration

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestGit_Status(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "file1.txt"},
	}

	msg := messages.GitStatusMsg{
		Path: "/test",
		Statuses: map[string]string{
			"file1.txt": "M",
		},
		Branch: "main",
	}

	HandleGit(m, msg)

	if m.Navigation.Items[0].GitStatus != "M" {
		t.Errorf("expected GitStatus M, got %s", m.Navigation.Items[0].GitStatus)
	}
	if m.Git.Branch != "main" {
		t.Errorf("expected branch main, got %s", m.Git.Branch)
	}
}
