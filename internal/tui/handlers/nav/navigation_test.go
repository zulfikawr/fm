package nav_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestHandleNavKeys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	item1 := core.Item{Name: "file1.txt", Path: "/test/file1.txt"}
	item2 := core.Item{Name: "file2.txt", Path: "/test/file2.txt"}
	m.Navigation.Items = []core.Item{item1, item2}
	m.Navigation.FilteredItems = []core.Item{item1, item2}
	m.Navigation.Cursor = 0

	t.Run("Move Cursor Down", func(t *testing.T) {
		nav.HandleNavKeys(m, tea.KeyMsg{Type: tea.KeyDown})
		if m.Navigation.Cursor != 1 {
			t.Errorf("expected cursor 1, got %d", m.Navigation.Cursor)
		}
	})

	t.Run("Toggle Selection", func(t *testing.T) {
		nav.HandleNavKeys(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		if m.Navigation.SelectedCount != 1 {
			t.Errorf("expected 1 selected, got %d", m.Navigation.SelectedCount)
		}
	})

	t.Run("Select All", func(t *testing.T) {
		nav.HandleNavKeys(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true})
		if m.Navigation.SelectedCount != 2 {
			t.Errorf("expected 2 selected, got %d", m.Navigation.SelectedCount)
		}
	})
}

func TestNavigation_History(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{FName: filepath.Base(path), FIsDir: true}, nil
	}
	m := tuictx.NewModel(fs, "/test")

	nav.NavigateToPath(m, "/home")
	if m.Navigation.Path != "/home" {
		t.Errorf("expected path /home, got %s", m.Navigation.Path)
	}

	nav.NavigateBack(m)
	if m.Navigation.Path != "/test" {
		t.Errorf("expected path /test after back, got %s", m.Navigation.Path)
	}

	nav.NavigateForward(m)
	if m.Navigation.Path != "/home" {
		t.Errorf("expected path /home after forward, got %s", m.Navigation.Path)
	}
}

func TestTabManagement(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Create Tab", func(t *testing.T) {
		nav.CreateTab(m)
		if len(m.Tabs) != 2 {
			t.Errorf("expected 2 tabs, got %d", len(m.Tabs))
		}
	})

	t.Run("Switch Tab", func(t *testing.T) {
		nav.SwitchTab(m, 1)
		if m.ActiveTab != 0 {
			t.Errorf("expected active tab 0, got %d", m.ActiveTab)
		}
	})

	t.Run("Close Tab", func(t *testing.T) {
		nav.CreateTab(m)
		nav.CloseTab(m)
		if len(m.Tabs) != 2 {
			t.Errorf("expected 2 tabs, got %d", len(m.Tabs))
		}
	})
}

func TestNavigation_Extra(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("HandleNavigation messages", func(t *testing.T) {
		nav.HandleNavigation(m, messages.PartialItemsMsg{Items: []core.Item{{Name: "p1"}}})
		nav.HandleNavigation(m, messages.ArchiveEnteredMsg{ParentPath: "/test", ParentFS: fs, FS: fs})
	})

	t.Run("SwitchToLocal", func(t *testing.T) {
		nav.SwitchToLocal(m, "/home")

		// Test switching when already local
		m.FS = testutil.NewMockFileSystem() // Re-mock
		nav.SwitchToLocal(m, "/tmp")
	})

	t.Run("Archive methods", func(t *testing.T) {
		m.Navigation.ParentFS = fs
		m.Navigation.ParentPath = "/parent"
		nav.ExitArchive(m)

		item := core.Item{Name: "test.zip", Path: "/test.zip"}
		nav.EnterArchive(m, item)
	})

	t.Run("HandleGotoFinalize", func(t *testing.T) {
		nav.HandleGotoFinalize(m, "/tmp")
		nav.HandleGotoFinalize(m, "user@host")
	})

	t.Run("TriggerFilter", func(t *testing.T) {
		nav.TriggerFilter(m)
	})
}

func TestNavigation_Reload(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Reload", func(t *testing.T) {
		nav.Reload(m, false)
	})
}

func TestWatchDirAction(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Path = "/test"

	nav.WatchDirAction(m)
}
