package actions

import (
	"context"
	testutil "fm/internal/testutil"
	tuitestutil "fm/internal/tui/testutil"
	"os"
	"testing"

	"fm/internal/files/core"
)

func TestUpdateProgress(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Progress.Visible = false

	UpdateProgress(m, "Copying...", 0.5)

	if !m.Operations.Progress.Visible {
		t.Error("Expected progress to be visible")
	}
	if m.Operations.Progress.Percent != 0.5 {
		t.Errorf("Expected progress percent 0.5, got %f", m.Operations.Progress.Percent)
	}
	if m.Operations.Progress.Label != "Copying..." {
		t.Errorf("Expected progress label 'Copying...', got '%s'", m.Operations.Progress.Label)
	}
}

func TestFinalizeOperation(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Progress.Visible = true
	m.Operations.ProcessingItems["/test/file1.txt"] = true
	m.Navigation.SelectedPaths["/test/file1.txt"] = true
	m.UI.SelectMode = true

	FinalizeOperation(m, []string{"/test/file1.txt"})

	if m.Operations.Progress.Visible {
		t.Error("Expected progress to be hidden after operation finished")
	}
	if m.Operations.ProcessingItems["/test/file1.txt"] {
		t.Error("Expected item to be removed from processing items")
	}
	if m.Navigation.SelectedPaths["/test/file1.txt"] {
		t.Error("Expected item to be removed from selected paths")
	}
	if m.UI.SelectMode {
		t.Error("Expected select mode to be disabled when no items are selected")
	}
}

func TestPerformDelete(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Items = []core.Item{
		{Name: "f1", Path: "/test/f1", Selected: true},
	}
	m.Navigation.FilteredItems = m.Navigation.Items
	m.Navigation.Cursor = 0

	cmds := PerformDelete(m)
	if len(cmds) == 0 {
		t.Error("Expected commands for deletion")
	}
	if !m.UI.Loading {
		t.Error("Expected loading true")
	}
}

func TestPerformDelete_Cursor(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.FilteredItems = []core.Item{
		{Name: "f1", Path: "/test/f1"},
	}
	m.Navigation.Cursor = 0

	cmds := PerformDelete(m)
	if len(cmds) == 0 {
		t.Error("Expected commands for deletion via cursor")
	}
}

func TestPerformPaste_Empty(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Clipboard.Paths = []string{}

	cmds := PerformPaste(m)
	if len(cmds) != 0 {
		t.Error("Expected no commands for empty paste")
	}
}

func TestPerformPaste(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Clipboard.Paths = []string{"/other/f1"}
	m.Operations.Clipboard.IsCut = false

	cmds := PerformPaste(m)
	if len(cmds) == 0 {
		t.Error("Expected commands for paste")
	}
}

func TestPerformRename(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Items = []core.Item{
		{Name: "old", Path: "/test/old"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items
	m.Navigation.Cursor = 0

	// Test success mock
	mockFS := m.FS.(*testutil.MockFileSystem)
	mockFS.RenameFunc = func(ctx context.Context, old, new string) error { return nil }
	mockFS.JoinFunc = func(elem ...string) string { return "/test/new" }

	cmd := PerformRename(m, "new")
	if cmd == nil {
		t.Error("Expected batch command for rename success")
	}
}

func TestResolveConflict(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Conflict.Source = "src"
	m.Operations.Conflict.Destination = "dst"

	cmd := ResolveConflict("overwrite", m)
	if cmd == nil {
		t.Error("Expected command for resolve conflict")
	}
	if m.UI.Confirming {
		t.Error("Expected confirming false after resolve")
	}
}

func TestResolveConflict_Skip(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Conflict.Source = "src"

	_ = ResolveConflict("skip", m)
	if m.UI.Confirming {
		t.Error("Expected confirming false")
	}
}

func TestResolveConflict_Rename(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Operations.Conflict.Source = "src"
	m.Operations.Conflict.Destination = "dst.txt"

	mockFS := m.FS.(*testutil.MockFileSystem)
	mockFS.StatFunc = func(ctx context.Context, p string) (os.FileInfo, error) {
		if p == "dst (1).txt" {
			return nil, os.ErrNotExist
		}
		return &testutil.MockFileInfo{}, nil
	}

	_ = ResolveConflict("rename", m)
	if m.UI.Confirming {
		t.Error("Expected confirming false")
	}
}
