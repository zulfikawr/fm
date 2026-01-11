//go:build windows

package files

// getStatInfo returns nil on Windows since stat information is not available
func getStatInfo(sys interface{}) *statInfo {
	return nil
}

type statInfo struct {
	dev    uint64
	ino    uint64
	blocks int64
}
