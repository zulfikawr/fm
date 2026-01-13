//go:build windows

package local

func preallocate(path string, size int64) error {
	// Fallback for non-linux systems (no-op for now)
	return nil
}
