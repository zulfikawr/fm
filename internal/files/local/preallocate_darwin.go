//go:build darwin

package local

// preallocate is a no-op on Darwin (macOS)
// Darwin doesn't support the fallocate syscall
func preallocate(path string, size int64) error {
	return nil
}
