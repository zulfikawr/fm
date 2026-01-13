//go:build unix

package local

import (
	"golang.org/x/sys/unix"
)

// isReadOnly checks if a path is on a read-only filesystem or is otherwise unwritable
func isReadOnly(path string) (bool, error) {
	// W_OK checks for write permission
	err := unix.Access(path, unix.W_OK)
	if err == nil {
		return false, nil
	}

	// If we get EROFS, it's definitely a read-only filesystem
	if err == unix.EROFS {
		return true, nil
	}

	// If it's a permission error, it might be ACLs or simple permission bits,
	// but for the UI we treat it as "Read Only"
	if err == unix.EACCES {
		return true, nil
	}

	return false, err
}
