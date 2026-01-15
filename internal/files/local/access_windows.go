//go:build windows

package local

import (
	"os"
)

// isReadOnly checks if a path is on a read-only filesystem or is otherwise unwritable on Windows
func isReadOnly(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// On Windows, the FileMode's read-only bit is generally accurate for current user access
	// when combined with standard file attributes.
	return info.Mode().Perm()&0o200 == 0, nil
}
