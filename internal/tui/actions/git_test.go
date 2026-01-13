package actions

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/commands"
)

func TestApplyGitStatus(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.Items = []core.Item{
		{Name: "f1", Path: "/test/f1"},
	}

	msg := commands.GitStatusMsg{
		Path:     "/test",
		Branch:   "main",
		Statuses: map[string]string{"f1": "M", "ghost": "D"},
	}

	ApplyGitStatus(m, msg)

	if m.Git.Branch != "main" {
		t.Errorf("Expected branch main, got %s", m.Git.Branch)
	}
	if m.Navigation.Items[0].GitStatus != "M" {
		t.Errorf("Expected status M, got %s", m.Navigation.Items[0].GitStatus)
	}

	// Check ghost item added
	foundGhost := false
	for _, item := range m.Navigation.Items {
		if item.Name == "ghost" && item.IsGhost {
			foundGhost = true
			break
		}
	}
	if !foundGhost {
		t.Error("Expected ghost item to be added")
	}
}
