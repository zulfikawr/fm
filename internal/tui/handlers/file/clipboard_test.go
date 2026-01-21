package file_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
)

func TestClipboard_Basic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.FilteredItems = []core.Item{
		{Name: "f1.txt", Path: "/test/f1.txt"},
	}

	file.CopySelected(m)
	if len(m.Operations.Clipboard.Paths) != 1 {
		t.Error("expected 1 path in clipboard")
	}

	file.CutSelected(m)
	if !m.Operations.Clipboard.IsCut {
		t.Error("expected IsCut to be true")
	}
}
