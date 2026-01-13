package actions

import (
	"context"
	"os"
	"testing"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuitestutil "fm/internal/tui/testutil"
)

func TestNavigateToSelected_File(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.FilteredItems = []core.Item{
		{Name: "file.txt", IsDir: false},
	}
	m.Navigation.Cursor = 0

	cmd := NavigateToSelected(m)
	if cmd != nil {
		t.Error("Expected nil command for navigating to a file")
	}
}

func TestNavigateToSelected_Empty(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.FilteredItems = []core.Item{}

	cmd := NavigateToSelected(m)
	if cmd != nil {
		t.Error("Expected nil command for empty list")
	}
}

func TestNavigateToPath(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	mockFS := m.FS.(*testutil.MockFileSystem)

	mockFS.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}

	cmd := NavigateToPath(m, "/new")
	if cmd == nil {
		t.Fatal("Expected Reload command")
	}
	if m.Navigation.Path != "/new" {
		t.Errorf("Expected path /new, got %s", m.Navigation.Path)
	}
	if m.Navigation.PathGen != 1 {
		t.Errorf("Expected PathGen 1, got %d", m.Navigation.PathGen)
	}
}

func TestReload(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	cmd := Reload(m)
	if cmd == nil {
		t.Fatal("Expected reload command")
	}
	if !m.UI.Loading {
		t.Error("Expected loading to be true")
	}
}

func TestNavigateToParent(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test/dir"
	mockFS := m.FS.(*testutil.MockFileSystem)
	mockFS.DirFunc = func(path string) string { return "/test" }
	mockFS.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}

	_ = NavigateToParent(m)
	if m.Navigation.Path != "/test" {
		t.Errorf("Expected path /test, got %s", m.Navigation.Path)
	}
}

func TestNavigateToSelected(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.FilteredItems = []core.Item{
		{Name: "subdir", IsDir: true, CanRead: true},
	}
	m.Navigation.Cursor = 0

	mockFS := m.FS.(*testutil.MockFileSystem)
	mockFS.JoinFunc = func(elem ...string) string { return "/test/subdir" }
	mockFS.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}

	_ = NavigateToSelected(m)
	if m.Navigation.Path != "/test/subdir" {
		t.Errorf("Expected path /test/subdir, got %s", m.Navigation.Path)
	}
}

func TestNavigateToSelected_Unreadable(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.FilteredItems = []core.Item{
		{Name: "locked", IsDir: true, CanRead: false},
	}
	m.Navigation.Cursor = 0

	cmd := NavigateToSelected(m)
	if cmd == nil {
		t.Error("Expected SetMsg command for unreadable dir")
	}
}

func TestNavigateToSelected_IsUp(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test/sub"
	m.Navigation.FilteredItems = []core.Item{
		{Name: "..", IsUp: true},
	}
	m.Navigation.Cursor = 0

	mockFS := m.FS.(*testutil.MockFileSystem)
	mockFS.DirFunc = func(p string) string { return "/test" }
	mockFS.StatFunc = func(ctx context.Context, p string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}

	cmd := NavigateToSelected(m)
	if cmd == nil {
		t.Error("Expected navigation command")
	}
}

func TestClearSelection(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.SelectMode = true
	m.Navigation.Items = []core.Item{{Path: "/f1", Selected: true}}
	m.Navigation.SelectedPaths = map[string]bool{"/f1": true}
	m.Navigation.SelectedCount = 1

	m.ClearSelection()

	if m.UI.SelectMode {
		t.Error("Expected SelectMode false")
	}
	if m.Navigation.SelectedCount != 0 {
		t.Error("Expected SelectedCount 0")
	}
	if m.Navigation.Items[0].Selected {
		t.Error("Expected item deselected")
	}
}
