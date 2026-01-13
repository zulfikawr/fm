package actions

import (
	"errors"
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/commands"
)

func TestFinalizeDirectoryLoad_Success(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items: []core.Item{
			{Name: "file1.txt", IsDir: false},
			{Name: "dir1", IsDir: true},
		},
		GitBranch:  "main",
		GitRoot:    "/repo",
		IsReadOnly: false,
		Err:        nil,
	}

	cmd, handled := FinalizeDirectoryLoad(m, msg)

	if handled {
		t.Error("Expected handled to be false for successful load")
	}
	if cmd != nil {
		t.Error("Expected no command returned for successful load")
	}
	if m.UI.Loading {
		t.Error("Expected loading to be false after successful load")
	}
	if len(m.Navigation.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.Navigation.Items))
	}
	if m.Git.Branch != "main" {
		t.Errorf("Expected git branch 'main', got '%s'", m.Git.Branch)
	}
}

func TestFinalizeDirectoryLoad_Error(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.PathGen = 1
	m.Navigation.Path = "/test"
	m.UI.Loading = true

	testErr := errors.New("failed to read directory")
	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      nil,
		Err:        testErr,
	}

	_, handled := FinalizeDirectoryLoad(m, msg)

	if !handled {
		t.Error("Expected handled to be true when error occurs")
	}
	if m.Message.Error == nil {
		t.Error("Expected error to be set in model")
	}
}

func TestFinalizeDirectoryLoad_CursorRestore(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	m.Cache.CursorMemory.Put("/test", 5)
	m.Cache.OffsetMemory.Put("/test", 3)

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      make([]core.Item, 10),
		Err:        nil,
	}

	FinalizeDirectoryLoad(m, msg)

	if m.Navigation.Cursor != 5 {
		t.Errorf("Expected cursor to be restored to 5, got %d", m.Navigation.Cursor)
	}
	if m.Navigation.Offset != 3 {
		t.Errorf("Expected offset to be restored to 3, got %d", m.Navigation.Offset)
	}
}
