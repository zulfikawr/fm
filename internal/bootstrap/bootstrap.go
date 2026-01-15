package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"fm/internal/files/factory"
	"fm/internal/tui"
	"fm/internal/tui/context"
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

	app := tui.NewApp(context.NewModel(fs, startPath))

	if remoteStr != "" {
		app.Model.Remote.Host = remoteInfo.Host
		app.Model.Remote.User = remoteInfo.User
	}

	return app, nil
}
