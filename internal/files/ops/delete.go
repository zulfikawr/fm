package ops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

// Delete removes a file or directory recursively.
func Delete(opts DeleteOptions) error {
	if len(opts.Paths) == 0 || opts.Paths[0] == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Delete", "")
	}
	path := opts.Paths[0]
	select {
	case <-opts.OpCtx.Context.Done():
		return opts.OpCtx.Context.Err()
	default:
	}

	if opts.OpCtx.Progress != nil {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 0,
			Label:   "Deleting " + opts.OpCtx.FS.Base(path) + "...",
		}:
		default:
		}
	}

	err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, path)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Delete", path)
	}

	if opts.OpCtx.Progress != nil {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 1.0,
			Label:   "Deleted " + opts.OpCtx.FS.Base(path),
		}:
		default:
		}
	}

	return nil
}

// Trash moves a file or directory to the system trash.
func Trash(ctx context.Context, fs core.FileSystem, path string) error {
	if path == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Trash", "")
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
		path = fs.Clean(path)
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

		cmd := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::%s('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`, op, path)
		err = exec.Command("powershell", "-Command", cmd).Run()
		if err != nil {
			return errors.WrapErrorWithPath(err, "Trash", path)
		}
	default:
		return errors.WrapErrorWithPath(fmt.Errorf("not supported on %s", runtime.GOOS), "Trash", path)
	}
	return nil
}
