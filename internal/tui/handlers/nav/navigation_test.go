package nav_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestNavigation_Basic(t *testing.T) {
	tmpDir := testutil.TempDir(t)
	config.SetConfigPath(filepath.Join(tmpDir, "config.json"))
	defer config.SetConfigPath("")

	fs := testutil.NewMockFileSystem()
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		return []os.FileInfo{
			&testutil.MockFileInfo{FName: "file1.txt"},
			&testutil.MockFileInfo{FName: "file2.txt"},
		}, nil
	}

	m := tuictx.NewModel(fs, "/test")
	nav.HandleNavigation(m, messages.LoadedItemsMsg{Items: []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "file2.txt", Path: "/test/file2.txt"},
	}})
}

func TestNavigation_Deep(t *testing.T) {
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{FName: "subdir", FIsDir: true}, nil
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		return nil, nil
	}

	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "subdir", Path: "/test/subdir", IsDir: true, CanRead: true},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	// Test Enter into directory
	nav.NavigateToSelected(m)
	if m.Navigation.Path != "/test/subdir" {
		t.Errorf("expected path to be /test/subdir, got %s", m.Navigation.Path)
	}

	// Test Parent navigation
	nav.NavigateToParent(m)
	if m.Navigation.Path != "/test" {
		t.Errorf("expected path to return to /test, got %s", m.Navigation.Path)
	}
}
