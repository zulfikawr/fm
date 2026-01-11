package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// FileError represents a wrapped file system error with context.
type FileError struct {
	Op  string
	Err error
	Msg string
}

func (e *FileError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s failed: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("%s failed: %v", e.Op, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

// WrapError converts system-level errors into user-friendly messages.
func WrapError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Don't double wrap
	var fe *FileError
	if errors.As(err, &fe) {
		return err
	}

	// Handle context cancellation
	if errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
		return &FileError{Op: op, Err: err, Msg: "timed out"}
	}

	if errors.Is(err, io.EOF) {
		return &FileError{Op: op, Err: err, Msg: "unexpected end of file"}
	}

	// Detect "Not Found"
	if errors.Is(err, os.ErrNotExist) || strings.Contains(strings.ToLower(err.Error()), "no such file") {
		return &FileError{Op: op, Err: err, Msg: "file or directory does not exist"}
	}

	// Detect "Already Exists"
	if errors.Is(err, os.ErrExist) || strings.Contains(strings.ToLower(err.Error()), "file exists") || strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return &FileError{Op: op, Err: err, Msg: "destination already exists"}
	}

	// Detect "File in Use" / Locked files
	if isLockedError(err) {
		return &FileError{Op: op, Err: err, Msg: "file is currently in use by another process"}
	}

	// Detect Permission issues
	if errors.Is(err, os.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "permission denied") {
		return &FileError{Op: op, Err: err, Msg: "permission denied"}
	}

	// Detect Disk Full
	if isDiskFullError(err) {
		return &FileError{Op: op, Err: err, Msg: "disk is full"}
	}

	// Detect "Not a directory"
	if strings.Contains(strings.ToLower(err.Error()), "not a directory") {
		return &FileError{Op: op, Err: err, Msg: "not a directory"}
	}

	// Detect "Is a directory"
	if strings.Contains(strings.ToLower(err.Error()), "is a directory") {
		return &FileError{Op: op, Err: err, Msg: "cannot perform this operation on a directory"}
	}

	// Default: Keep original but prefix with operation
	return &FileError{Op: op, Err: err}
}

func isLockedError(err error) bool {
	msg := strings.ToLower(err.Error())
	if runtime.GOOS == "windows" {
		// Common Windows error messages for locked files
		return strings.Contains(msg, "being used by another process") ||
			strings.Contains(msg, "sharing violation") ||
			strings.Contains(msg, "requested operation cannot be performed on a file with a user-mapped section open")
	}
	// On Unix/Linux/macOS, EBUSY (text file busy) is the equivalent
	return strings.Contains(msg, "text file busy") || strings.Contains(msg, "resource busy")
}

func isDiskFullError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no space left on device") ||
		strings.Contains(msg, "disk is full") ||
		strings.Contains(msg, "insufficient space")
}
