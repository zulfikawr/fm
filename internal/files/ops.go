package files

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Delete removes a file or directory recursively.
func Delete(fs FileSystem, path string) error {
	return fs.RemoveAll(path)
}

// Trash moves a file or directory to the system trash.
func Trash(fs FileSystem, path string) error {
	if !fs.IsLocal() {
		return fmt.Errorf("trash not supported on remote file systems")
	}

	switch runtime.GOOS {
	case "linux":
		return exec.Command("gio", "trash", path).Run()
	case "darwin":
		return exec.Command("osascript", "-e", `tell application "Finder" to delete POSIX file "`+path+`"`).Run()
	case "windows":
		path = filepath.Clean(path)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		op := "DeleteFile"
		if info.IsDir() {
			op = "DeleteDirectory"
		}

		cmd := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; [Microsoft.VisualBasic.FileIO.FileSystem]::%s('%s', 'OnlyErrorDialogs', 'SendToRecycleBin')`, op, path)
		return exec.Command("powershell", "-Command", cmd).Run()
	default:
		return fmt.Errorf("trash not supported on %s", runtime.GOOS)
	}
}

// Rename moves or renames a file or directory.
func Rename(fs FileSystem, oldPath, newPath string) error {
	return fs.Rename(oldPath, newPath)
}

// Copy copies a file or directory recursively from src to dst within the same filesystem.
func Copy(fs FileSystem, src, dst string) error {
	info, err := fs.Lstat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(fs, src, dst)
	}
	return copyFile(fs, src, dst)
}

func copyFile(fs FileSystem, src, dst string) error {
	out, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	info, err := fs.Stat(src)
	if err != nil {
		return err
	}
	return fs.Chmod(dst, info.Mode())
}

func copyDir(fs FileSystem, src, dst string) error {
	info, err := fs.Stat(src)
	if err != nil {
		return err
	}

	if err := fs.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := fs.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := fs.Join(src, entry.Name())
		dstPath := fs.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(fs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(fs, srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
