//go:build unix

package trash

import (
	"os"
	"syscall"
)

// getOwnership extracts UID and GID from FileInfo (Unix-specific).
func getOwnership(info os.FileInfo) (uid, gid int) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Uid), int(stat.Gid)
	}
	return 0, 0
}
