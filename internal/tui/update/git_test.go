package update

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
)

func TestHandleGitMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.Items = []core.Item{{Name: "f1"}}

	msg := commands.GitStatusMsg{
		Path:     "/test",
		Branch:   "feat",
		Statuses: map[string]string{"f1": "M"},
	}

	_ = HandleGitMsg(m, msg)
	if m.Git.Branch != "feat" {
		t.Errorf("Expected branch feat, got %s", m.Git.Branch)
	}
	if m.Navigation.Items[0].GitStatus != "M" {
		t.Errorf("Expected status M, got %s", m.Navigation.Items[0].GitStatus)
	}
}

func TestHandleGitMsg_Empty(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.Items = []core.Item{}

	msg := commands.GitStatusMsg{
		Path:     "/test",
		Statuses: map[string]string{"ghost": "D"},
	}
	// Use actions.ApplyGitStatus directly to test logic
	actions.ApplyGitStatus(m, msg)
	if len(m.Navigation.Items) != 1 {
		t.Error("Expected ghost item added to empty list")
	}
}
