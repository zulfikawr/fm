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
func Delete(path string) error {
	return os.RemoveAll(path)
}

// Trash moves a file or directory to the system trash.
func Trash(path string) error {
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
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// Copy copies a file or directory recursively from src to dst.
func Copy(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
