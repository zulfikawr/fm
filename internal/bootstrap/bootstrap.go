package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zulfikawr/fm/internal/files/factory"
	"github.com/zulfikawr/fm/internal/files/trash"
	"github.com/zulfikawr/fm/internal/logger"
	"github.com/zulfikawr/fm/internal/tui"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
)

// InitializeApp sets up the filesystem and creates the TUI app.
func InitializeApp(remoteStr string, args []string) (*tui.App, error) {
	fs, remoteInfo, err := factory.CreateFileSystem(remoteStr, args)
	if err != nil {
		return nil, err
	}

	var startPath string
	if remoteStr != "" {
		startPath = remoteInfo.StartPath
	} else {
		// Determine start path from arguments (non-flag)
		if len(args) > 0 {
			argPath := args[0]

			absPath, err := filepath.Abs(argPath)
			if err != nil {
				return nil, fmt.Errorf("resolving path: %w", err)
			}

			info, err := os.Stat(absPath)
			if err != nil {
				return nil, fmt.Errorf("accessing path: %w", err)
			}

			if !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", argPath)
			}
			startPath = absPath
		} else {
			startPath, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("getting current directory: %w", err)
			}
		}
	}

	app := tui.NewApp(tuictx.NewModel(fs, startPath))

	if remoteStr != "" {
		app.Model.Remote.Host = remoteInfo.Host
		app.Model.Remote.User = remoteInfo.User
	}

	// Run trash cleanup in background if local filesystem
	if fs.IsLocal() && app.Model.Config.UseTrash {
		go func() {
			manager, err := trash.NewManager(fs)
			if err != nil {
				logger.Warnf("Failed to create trash manager: %v", err)
				return
			}

			// Recover interrupted deletions
			if err := manager.RecoverInterruptedDeletions(context.Background()); err != nil {
				logger.Warnf("Failed to recover interrupted deletions: %v", err)
			}

			// Auto-cleanup based on config
			if err := manager.AutoCleanup(
				context.Background(),
				app.Model.Config.TrashAutoCleanupDays,
				int64(app.Model.Config.TrashMaxSizeMB),
			); err != nil {
				logger.Warnf("Failed to auto-cleanup trash: %v", err)
			}
		}()
	}

	return app, nil
}
