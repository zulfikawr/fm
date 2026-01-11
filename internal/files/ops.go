package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const FileOperationTimeout = 5 * time.Minute

// Progress represents the current state of a file operation.
type Progress struct {
	Percent float64
	Label   string
}

// Delete removes a file or directory recursively.
func Delete(ctx context.Context, fs FileSystem, path string, progChan chan<- Progress) error {
	if path == "" {
		return WrapError(fmt.Errorf("empty path"), "Delete")
	}
	if progChan != nil {
		select {
		case progChan <- Progress{Percent: 0, Label: "Deleting " + fs.Base(path) + "..."}:
		default:
		}
	}
	err := fs.RemoveAll(ctx, path)
	if progChan != nil {
		select {
		case progChan <- Progress{Percent: 1.0, Label: "Deleted " + fs.Base(path)}:
		default:
		}
		time.Sleep(100 * time.Millisecond) // Give UI time to show 100%
	}
	return WrapError(err, "Delete")
}

// Trash moves a file or directory to the system trash.
func Trash(ctx context.Context, fs FileSystem, path string) error {
	if path == "" {
		return WrapError(fmt.Errorf("empty path"), "Trash")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !fs.IsLocal() {
		return WrapError(fmt.Errorf("not supported on remote file systems"), "Trash")
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
						return WrapError(fmt.Errorf("neither 'gio' nor 'trash-put' found. Install glib2 or trash-cli"), "Trash")
					}
				}
				return WrapError(err, "Trash")
			}
		}
	case "darwin":
		err = exec.Command("osascript", "-e", `tell application "Finder" to delete POSIX file "`+path+`"`).Run()
		if err != nil {
			return WrapError(err, "Trash")
		}
	case "windows":
		path = filepath.Clean(path)
		info, sErr := os.Stat(path)
		if sErr != nil {
			return WrapError(sErr, "Trash")
		}

		op := "DeleteFile"
		if info.IsDir() {
			op = "DeleteDirectory"
		}

		// Check if PowerShell is available
		if _, err := exec.LookPath("powershell"); err != nil {
			return WrapError(fmt.Errorf("PowerShell not found (required on Windows)"), "Trash")
		}

		cmd := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::%s('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`, op, path)
		err = exec.Command("powershell", "-Command", cmd).Run()
		if err != nil {
			return WrapError(err, "Trash")
		}
	default:
		return WrapError(fmt.Errorf("not supported on %s", runtime.GOOS), "Trash")
	}
	return nil
}

// Rename moves or renames a file or directory.
func Rename(ctx context.Context, fs FileSystem, oldPath, newPath string) error {
	if oldPath == "" || newPath == "" {
		return WrapError(fmt.Errorf("empty path"), "Rename")
	}
	// Validate new filename component
	if err := ValidateFileName(fs.Base(newPath)); err != nil {
		return WrapError(err, "Rename")
	}
	return WrapError(fs.Rename(ctx, oldPath, newPath), "Rename")
}

// Move moves a file or directory. It tries Rename first, and falls back to Copy+Delete if Rename fails.
func Move(ctx context.Context, fs FileSystem, src, dst string, progChan chan<- Progress) error {
	if src == "" || dst == "" {
		return WrapError(fmt.Errorf("empty path"), "Move")
	}
	err := fs.Rename(ctx, src, dst)
	if err == nil {
		if progChan != nil {
			select {
			case progChan <- Progress{Percent: 1.0, Label: "Moved " + fs.Base(src)}:
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
		return WrapError(fmt.Errorf("move partially successful: items copied to destination but failed to remove from source: %w", err), "Move")
	}

	return nil
}

// Copy copies a file or directory recursively from src to dst within the same filesystem.
func Copy(ctx context.Context, fs FileSystem, src, dst string, progChan chan<- Progress) error {
	if src == "" || dst == "" {
		return WrapError(fmt.Errorf("empty path"), "Copy")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := fs.Lstat(ctx, src)
	if err != nil {
		return WrapError(err, "Copy")
	}

	if info.IsDir() {
		return WrapError(copyDir(ctx, fs, src, dst, progChan), "Copy")
	}
	return WrapError(copyFile(ctx, fs, src, dst, progChan), "Copy")
}

func copyFile(ctx context.Context, fs FileSystem, src, dst string, progChan chan<- Progress) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	out, err := fs.Create(ctx, dst)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := fs.Open(ctx, src)
	if err != nil {
		out.Close()
		fs.RemoveAll(ctx, dst) // Clean up partial file
		return err
	}
	defer in.Close()

	info, err := fs.Stat(ctx, src)
	if err != nil {
		return err
	}

	if progChan != nil {
		// For small files, just report start/end. For large files, we use a custom writer.
		if info.Size() < 1024*1024 { // 1MB
			select {
			case progChan <- Progress{Percent: 0, Label: "Copying " + fs.Base(src) + "..."}:
			default:
			}
			_, err = io.Copy(out, in)
			select {
			case progChan <- Progress{Percent: 1.0, Label: "Copied " + fs.Base(src)}:
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
		return err
	}

	return fs.Chmod(ctx, dst, info.Mode())
}

type progressWriter struct {
	io.Writer
	Total    int64
	Current  int64
	Label    string
	progChan chan<- Progress
	lastUpd  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	pw.Current += int64(n)
	if pw.progChan != nil && time.Since(pw.lastUpd) > 100*time.Millisecond {
		pw.lastUpd = time.Now()
		select {
		case pw.progChan <- Progress{
			Percent: float64(pw.Current) / float64(pw.Total),
			Label:   pw.Label + "...",
		}:
		default:
		}
	}
	if pw.Current == pw.Total && pw.progChan != nil {
		select {
		case pw.progChan <- Progress{
			Percent: 1.0,
			Label:   pw.Label + " (Done)",
		}:
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
	return n, err
}

func copyDir(ctx context.Context, fs FileSystem, src, dst string, progChan chan<- Progress) error {
	return copyDirRecursive(ctx, fs, src, dst, make(map[string]bool), progChan)
}

func copyDirRecursive(ctx context.Context, fs FileSystem, src, dst string, visited map[string]bool, progChan chan<- Progress) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	absSrc, err := fs.Abs(src)
	if err != nil {
		return err
	}
	if visited[absSrc] {
		return nil // Skip already visited path to avoid loops
	}
	visited[absSrc] = true

	info, err := fs.Stat(ctx, src)
	if err != nil {
		return err
	}

	if err := fs.MkdirAll(ctx, dst, info.Mode()); err != nil {
		return err
	}

	entries, err := fs.ReadDir(ctx, src)
	if err != nil {
		return err
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
