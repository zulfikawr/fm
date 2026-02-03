//go:build !windows

package files

import (
	"os"
	"syscall"
)

// GetDeviceID extracts the device ID from FileInfo on Unix systems
func GetDeviceID(info os.FileInfo) uint64 {
	if info == nil {
		return 0
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return uint64(stat.Dev)
	}
	return 0
}
