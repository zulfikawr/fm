package ops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fm/internal/files/core"
	"fm/internal/files/errors"
)

// Delete removes a file or directory recursively.
func Delete(ctx context.Context, fs core.FileSystem, path string, progChan chan<- core.Progress) error {
	if path == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Delete", path)
	}
	if progChan != nil {
		select {
		case progChan <- core.Progress{Percent: 0, Label: "Deleting " + fs.Base(path) + "..."}:
		default:
		}
	}
	err := fs.RemoveAll(ctx, path)
	if progChan != nil {
		select {
		case progChan <- core.Progress{Percent: 1.0, Label: "Deleted " + fs.Base(path)}:
		default:
		}
		time.Sleep(100 * time.Millisecond) // Give UI time to show 100%
	}
	return errors.WrapErrorWithPath(err, "Delete", path)
}

// Trash moves a file or directory to the system trash.
func Trash(ctx context.Context, fs core.FileSystem, path string) error {
	if path == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Trash", path)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !fs.IsLocal() {
		return errors.WrapErrorWithPath(&errors.UnsupportedOperationError{Op: "Trash", Filesystem: "Remote"}, "Trash", path)
	}

	var err error
	switch runtime.GOOS {
	case "linux":
		// Try gio first
		err = exec.Command("gio", "trash", path).Run()
		if err != nil {
			// Fallback to trash-cli
			err = exec.Command("trash-put", path).Run()
			if err != nil {
				// If both fail, check if commands exist
				if _, gioErr := exec.LookPath("gio"); gioErr != nil {
					if _, trashErr := exec.LookPath("trash-put"); trashErr != nil {
						return errors.WrapErrorWithPath(fmt.Errorf("neither 'gio' nor 'trash-put' found. Install glib2 or trash-cli"), "Trash", path)
					}
				}
				return errors.WrapErrorWithPath(err, "Trash", path)
			}
		}
	case "darwin":
		err = exec.Command("osascript", "-e", `tell application "Finder" to delete POSIX file "`+path+`"`).Run()
		if err != nil {
			return errors.WrapErrorWithPath(err, "Trash", path)
		}
	case "windows":
		path = filepath.Clean(path)
		info, sErr := os.Stat(path)
		if sErr != nil {
			return errors.WrapErrorWithPath(sErr, "Trash", path)
		}

		op := "DeleteFile"
		if info.IsDir() {
			op = "DeleteDirectory"
		}

		// Check if PowerShell is available
		if _, err := exec.LookPath("powershell"); err != nil {
			return errors.WrapErrorWithPath(fmt.Errorf("PowerShell not found (required on Windows)"), "Trash", path)
		}

		cmd := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.core.FileSystem]::%s('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`, op, path)
		err = exec.Command("powershell", "-Command", cmd).Run()
		if err != nil {
			return errors.WrapErrorWithPath(err, "Trash", path)
		}
	default:
		return errors.WrapErrorWithPath(fmt.Errorf("not supported on %s", runtime.GOOS), "Trash", path)
	}
	return nil
}
