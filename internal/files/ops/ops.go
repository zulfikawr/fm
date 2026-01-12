package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fm/internal/files"
	"fm/internal/files/errors"
)

// Delete removes a file or directory recursively.
func Delete(ctx context.Context, fs files.FileSystem, path string, progChan chan<- files.Progress) error {
	if path == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Delete", path)
	}
	if progChan != nil {
		select {
		case progChan <- files.Progress{Percent: 0, Label: "Deleting " + fs.Base(path) + "..."}:
		default:
		}
	}
	err := fs.RemoveAll(ctx, path)
	if progChan != nil {
		select {
		case progChan <- files.Progress{Percent: 1.0, Label: "Deleted " + fs.Base(path)}:
		default:
		}
		time.Sleep(100 * time.Millisecond) // Give UI time to show 100%
	}
	return errors.WrapErrorWithPath(err, "Delete", path)
}

// Trash moves a file or directory to the system trash.
func Trash(ctx context.Context, fs files.FileSystem, path string) error {
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

		cmd := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.files.FileSystem]::%s('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`, op, path)
		err = exec.Command("powershell", "-Command", cmd).Run()
		if err != nil {
			return errors.WrapErrorWithPath(err, "Trash", path)
		}
	default:
		return errors.WrapErrorWithPath(fmt.Errorf("not supported on %s", runtime.GOOS), "Trash", path)
	}
	return nil
}

// Rename moves or renames a file or directory.
func Rename(ctx context.Context, fs files.FileSystem, oldPath, newPath string) error {
	if oldPath == "" || newPath == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
	}
	// Validate new filename component
	if err := ValidateFileName(fs.Base(newPath)); err != nil {
		return errors.WrapErrorWithPath(err, "Rename", newPath)
	}
	return errors.WrapErrorWithPath(fs.Rename(ctx, oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}

// Move moves a file or directory. It tries Rename first, and falls back to Copy+Delete if Rename fails.
func Move(ctx context.Context, fs files.FileSystem, src, dst string, progChan chan<- files.Progress) error {
	if src == "" || dst == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Move", fmt.Sprintf("%s -> %s", src, dst))
	}
	err := fs.Rename(ctx, src, dst)
	if err == nil {
		if progChan != nil {
			select {
			case progChan <- files.Progress{Percent: 1.0, Label: "Moved " + fs.Base(src)}:
			default:
			}
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	}

	// Fallback for cross-device moves
	if err := Copy(ctx, fs, src, dst, progChan); err != nil {
		return err // Copy already uses WrapError
	}

	if err := Delete(ctx, fs, src, nil); err != nil {
		return errors.WrapErrorWithPath(fmt.Errorf("move partially successful: items copied to destination but failed to remove from source: %w", err), "Move", src)
	}

	return nil
}

// Copy copies a file or directory recursively from src to dst within the same filesystem.
func Copy(ctx context.Context, fs files.FileSystem, src, dst string, progChan chan<- files.Progress) error {
	if src == "" || dst == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Copy", fmt.Sprintf("%s -> %s", src, dst))
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := fs.Lstat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Copy", src)
	}

	if info.IsDir() {
		return errors.WrapErrorWithPath(copyDir(ctx, fs, src, dst, progChan), "Copy", fmt.Sprintf("%s -> %s", src, dst))
	}
	return errors.WrapErrorWithPath(copyFile(ctx, fs, src, dst, progChan), "Copy", fmt.Sprintf("%s -> %s", src, dst))
}

func copyFile(ctx context.Context, fs files.FileSystem, src, dst string, progChan chan<- files.Progress) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	out, err := fs.Create(ctx, dst)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Create", dst)
	}
	defer out.Close()

	in, err := fs.Open(ctx, src)
	if err != nil {
		out.Close()
		fs.RemoveAll(ctx, dst) // Clean up partial file
		return errors.WrapErrorWithPath(err, "Open", src)
	}
	defer in.Close()

	info, err := fs.Stat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}

	if progChan != nil {
		// For small files, just report start/end. For large files, we use a custom writer.
		if info.Size() < 1024*1024 { // 1MB
			select {
			case progChan <- files.Progress{Percent: 0, Label: "Copying " + fs.Base(src) + "..."}:
			default:
			}
			_, err = io.Copy(out, in)
			select {
			case progChan <- files.Progress{Percent: 1.0, Label: "Copied " + fs.Base(src)}:
			default:
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			pw := &progressWriter{
				Writer:   out,
				Total:    info.Size(),
				Label:    "Copying " + fs.Base(src),
				progChan: progChan,
			}
			_, err = io.Copy(pw, in)
		}
	} else {
		_, err = io.Copy(out, in)
	}

	if err != nil {
		out.Close()
		fs.RemoveAll(ctx, dst) // Clean up partial file
		return errors.WrapErrorWithPath(err, "CopyFile", fmt.Sprintf("%s -> %s", src, dst))
	}

	return errors.WrapErrorWithPath(fs.Chmod(ctx, dst, info.Mode()), "Chmod", dst)
}

type progressWriter struct {
	io.Writer
	Total    int64
	Current  int64
	Label    string
	progChan chan<- files.Progress
	lastUpd  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Current += int64(n)
	if pw.progChan != nil && time.Since(pw.lastUpd) > 100*time.Millisecond {
		pw.lastUpd = time.Now()
		select {
		case pw.progChan <- files.Progress{
			Percent: float64(pw.Current) / float64(pw.Total),
			Label:   pw.Label + "...",
		}:
		default:
		}
	}
	if pw.Current == pw.Total && pw.progChan != nil {
		select {
		case pw.progChan <- files.Progress{
			Percent: 1.0,
			Label:   pw.Label + " (Done)",
		}:
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
	return n, err
}

func copyDir(ctx context.Context, fs files.FileSystem, src, dst string, progChan chan<- files.Progress) error {
	return copyDirRecursive(ctx, fs, src, dst, make(map[string]bool), progChan)
}

func copyDirRecursive(ctx context.Context, fs files.FileSystem, src, dst string, visited map[string]bool, progChan chan<- files.Progress) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	absSrc, err := fs.Abs(src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Abs", src)
	}
	if visited[absSrc] {
		return nil // Skip already visited path to avoid loops
	}
	visited[absSrc] = true

	info, err := fs.Stat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}

	if err := fs.MkdirAll(ctx, dst, info.Mode()); err != nil {
		return errors.WrapErrorWithPath(err, "MkdirAll", dst)
	}

	entries, err := fs.ReadDir(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ReadDir", src)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		srcPath := fs.Join(src, entry.Name())
		dstPath := fs.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(ctx, fs, srcPath, dstPath, visited, progChan); err != nil {
				return err
			}
		} else {
			if err := copyFile(ctx, fs, srcPath, dstPath, progChan); err != nil {
				return err
			}
		}
	}
	return nil
}
