//go:build windows

package files

import (
	"os"
)

// GetDeviceID is a placeholder for Windows where Dev ID is handled differently
func GetDeviceID(info os.FileInfo) uint64 {
	return 0
}
