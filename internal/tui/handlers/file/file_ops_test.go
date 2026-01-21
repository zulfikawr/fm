package file

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestFileOps_Basic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	// Test StartCreate
	StartCreate(m)
	if m.Inputs.Mode != tuictx.InputCreate {
		t.Error("expected InputCreate mode")
	}

	// Test PerformCreate
	PerformCreate(m, "newfile.txt")
}

func TestFileOps_Rename(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.FilteredItems = []core.Item{
		{Name: "old.txt", Path: "/test/old.txt"},
	}

	StartRename(m)
	if m.Inputs.Mode != tuictx.InputRename {
		t.Error("expected InputRename mode")
	}

	PerformRename(m, "new.txt")
}

func TestFileOps_Conflict(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	msg := messages.ConflictMsg{
		Src:    "/src/a",
		Dst:    "/dest/a",
		OpType: "copy",
	}

	HandleConflict(m, msg)
	if !m.UI.Confirming {
		t.Error("expected Confirming state")
	}
}
