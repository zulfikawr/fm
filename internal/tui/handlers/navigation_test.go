package handlers

import (
	"context"
	"os"
	"testing"
	"time"

	"fm/internal/files/core"
	"fm/internal/testutil"
	tuictx "fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNavigation_Basic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		return []os.FileInfo{
			&testutil.MockFileInfo{NameStr: "file1.txt"},
			&testutil.MockFileInfo{NameStr: "file2.txt"},
		}, nil
	}

	m := tuictx.NewModel(fs, "/test")
	HandleUpdate(m, LoadedItemsMsg{Items: []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "file2.txt", Path: "/test/file2.txt"},
	}})

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Move down
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(10 * time.Millisecond)
	if m.Navigation.Cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.Navigation.Cursor)
	}

	// Move up
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(10 * time.Millisecond)
	if m.Navigation.Cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.Navigation.Cursor)
	}

	tm.Quit()
}

func TestNavigation_Deep(t *testing.T) {
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: "subdir", IsDirBool: true}, nil
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		return nil, nil
	}

	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "subdir", Path: "/test/subdir", IsDir: true, CanRead: true},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// 1. Test Enter into directory
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)
	if m.Navigation.Path != "/test/subdir" {
		t.Errorf("expected path to be /test/subdir, got %s", m.Navigation.Path)
	}

	// 2. Test Parent navigation (backspace)
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	time.Sleep(50 * time.Millisecond)
	if m.Navigation.Path != "/test" {
		t.Errorf("expected path to return to /test, got %s", m.Navigation.Path)
	}

	tm.Quit()
}

func TestNavigation_Tabs(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Create tab
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true})
	time.Sleep(50 * time.Millisecond)
	if len(m.Tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(m.Tabs))
	}

	// Switch to tab 1
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1"), Alt: true})
	time.Sleep(50 * time.Millisecond)
	if m.ActiveTab != 0 {
		t.Errorf("expected active tab 0, got %d", m.ActiveTab)
	}

	// Close tab
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w"), Alt: true})
	time.Sleep(50 * time.Millisecond)
	if len(m.Tabs) != 1 {
		t.Errorf("expected 1 tab, got %d", len(m.Tabs))
	}

	tm.Quit()
}

func TestNavigation_Selection(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "f1.txt", Path: "/test/f1.txt"},
		{Name: "f2.txt", Path: "/test/f2.txt"},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// Toggle selection on first item
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(10 * time.Millisecond)
	tm.Quit()
}

func TestNavigation_Memory(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: "subdir", IsDirBool: true}, nil
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		return nil, nil
	}

	m := tuictx.NewModel(fs, "/test")
	m.Navigation.Items = []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "subdir", Path: "/test/subdir", IsDir: true, CanRead: true},
	}
	m.Navigation.FilteredItems = m.Navigation.Items

	wrapper := newTestModelWrapper(m)
	tm := testutil.NewTestModel(t, wrapper)

	// 1. Move cursor to index 1 (subdir)
	m.Navigation.Cursor = 1
	m.Cache.CursorMemory.Put("/test", 1)
	syncOffset(m)
	// 2. Navigate into subdir
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)
	if m.Navigation.Path != "/test/subdir" {
		t.Fatalf("expected path /test/subdir, got %s", m.Navigation.Path)
	}

	// 3. Navigate back to /test
	tm.Send(tea.KeyMsg{Type: tea.KeyBackspace})
	time.Sleep(50 * time.Millisecond)

	if m.Navigation.Path != "/test" {
		t.Fatalf("expected path to be /test, got %s", m.Navigation.Path)
	}

	if val, ok := m.Cache.CursorMemory.Get("/test"); !ok || val != 1 {
		t.Fatalf("expected cache to have 1 for /test, got %d (ok=%v)", val, ok)
	}

	// Need to simulate LoadedItemsMsg since NavigateToPath/Reload triggers it
	testItems := []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "subdir", Path: "/test/subdir", IsDir: true, CanRead: true},
	}
	HandleUpdate(m, LoadedItemsMsg{
		Path:       "/test",
		Items:      testItems,
		Generation: m.Navigation.PathGen,
	}) // 4. Check if cursor was restored
	if m.Navigation.Cursor != 1 {
		t.Errorf("expected cursor to be restored to 1, got %d", m.Navigation.Cursor)
	}
	tm.Quit()
}
