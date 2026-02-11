//go:build linux

package local

import (
	"os"

	"github.com/zulfikawr/fm/internal/logger"
	"golang.org/x/sys/unix"
)

func preallocate(path string, size int64) error {
	if size <= 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(f, "file during preallocate")

	// Mode 0 is the default allocation
	return unix.Fallocate(int(f.Fd()), 0, 0, size)
}
