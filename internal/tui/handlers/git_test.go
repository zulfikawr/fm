package handlers

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

func TestGit_StatusMsg(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
	}

	msg := GitStatusMsg{
		Path:   "/test",
		Branch: "main",
		Statuses: map[string]string{
			"file1.txt": "M",
			"ghost.txt": "D",
		},
	}

	HandleGit(m, msg)

	if m.Git.Branch != "main" {
		t.Errorf("expected branch main, got %s", m.Git.Branch)
	}

	foundModified := false
	foundGhost := false
	for _, item := range m.Navigation.Items {
		if item.Name == "file1.txt" && item.GitStatus == "M" {
			foundModified = true
		}
		if item.Name == "ghost.txt" && item.IsGhost && item.GitStatus == "D" {
			foundGhost = true
		}
	}

	if !foundModified {
		t.Error("expected file1.txt to have modified status")
	}
	if !foundGhost {
		t.Error("expected ghost.txt to be added to items")
	}
}
