//go:build unix

package local

import (
	"syscall"
)

// getStatInfo extracts stat information from os.FileInfo on Unix systems
func getStatInfo(sys interface{}) *statInfo {
	if sys == nil {
		return nil
	}
	if stat, ok := sys.(*syscall.Stat_t); ok {
		return &statInfo{
			dev:    uint64(stat.Dev),
			ino:    uint64(stat.Ino),
			blocks: stat.Blocks,
		}
	}
	return nil
}

type statInfo struct {
	dev    uint64
	ino    uint64
	blocks int64
}
