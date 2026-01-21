package tui_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui"
)

func TestApp_Integration(t *testing.T) {
	// 1. Setup App with Mock FS
	app, fs := SetupTestApp("/test")

	// Mock a file to appear in the list
	fs.ReadDirEntriesFunc = func(ctx context.Context, path string) ([]os.DirEntry, error) {
		return []os.DirEntry{
			&testutil.MockDirEntry{NameStr: "hello.txt", IsDirBool: false},
		}, nil
	}
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/test" {
			return &testutil.MockFileInfo{FName: "test", FIsDir: true}, nil
		}
		return &testutil.MockFileInfo{FName: "hello.txt", FIsDir: false}, nil
	}

	// 2. Initialize Teatest with dimensions
	tm := testutil.NewTestModel(t, app, teatest.WithInitialTermSize(80, 24))

	// 3. Assert initial view contains mocked file
	testutil.WaitAndAssertView(t, tm, "hello.txt", 2*time.Second)

	// 4. Test Navigation
	tm.Send("j")

	// 5. Quit
	_ = tm.Quit()
}

func TestApp_Lifecycle(t *testing.T) {
	app, _ := SetupTestApp("/test")

	t.Run("InitModel", func(t *testing.T) {
		cmd := tui.InitModel(app.Model)
		if cmd == nil {
			t.Error("InitModel should return a command")
		}
	})

	t.Run("Close", func(t *testing.T) {
		// Just ensure it doesn't panic
		tui.Close(app.Model)
	})
}
