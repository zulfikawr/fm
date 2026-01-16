package errors

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
	Op   string
	Path string
	Err  error
	Msg  string
}

func (e *FileError) Error() string {
	pathInfo := ""
	if e.Path != "" {
		pathInfo = fmt.Sprintf(" at %s", e.Path)
	}

	if e.Msg != "" {
		return fmt.Sprintf("%s failed%s: %s", e.Op, pathInfo, e.Msg)
	}
	return fmt.Sprintf("%s failed%s: %v", e.Op, pathInfo, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

// UnsupportedOperationError indicates an operation is not supported
type UnsupportedOperationError struct {
	Op         string
	Filesystem string
}

func (e *UnsupportedOperationError) Error() string {
	return fmt.Sprintf("operation %s is not supported by %s filesystem", e.Op, e.Filesystem)
}

// ValidationError indicates invalid input
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s (value: %v): %s", e.Field, e.Value, e.Message)
}

// PermissionError indicates insufficient permissions
type PermissionError struct {
	Path      string
	Operation string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for %s on path %s", e.Operation, e.Path)
}

// WrapError converts system-level errors into user-friendly messages.
func WrapError(err error, op string) error {
	return WrapErrorWithPath(err, op, "")
}

// WrapErrorWithPath converts system-level errors into user-friendly messages with path context.
func WrapErrorWithPath(err error, op string, path string) error {
	if err == nil {
		return nil
	}

	// Don't double wrap if it's already a FileError and we don't have a path to add
	var fe *FileError
	if errors.As(err, &fe) {
		if path != "" && fe.Path == "" {
			fe.Path = path
		}
		return err
	}

	// Handle specialized errors
	var ue *UnsupportedOperationError
	if errors.As(err, &ue) {
		return &FileError{Op: op, Path: path, Err: err, Msg: err.Error()}
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		return &FileError{Op: op, Path: path, Err: err, Msg: err.Error()}
	}

	var pe *PermissionError
	if errors.As(err, &pe) {
		return &FileError{Op: op, Path: path, Err: err, Msg: "permission denied"}
	}

	// Handle context cancellation
	if errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
		return &FileError{Op: op, Path: path, Err: err, Msg: "timed out"}
	}

	if errors.Is(err, io.EOF) {
		return &FileError{Op: op, Path: path, Err: err, Msg: "unexpected end of file"}
	}

	// Detect "Not Found" (including executables)
	if errors.Is(err, os.ErrNotExist) ||
		strings.Contains(strings.ToLower(err.Error()), "no such file") ||
		strings.Contains(strings.ToLower(err.Error()), "executable file not found") ||
		strings.Contains(strings.ToLower(err.Error()), "not found") {
		return &FileError{Op: op, Path: path, Err: err, Msg: err.Error()}
	}

	// Detect "Already Exists"
	if errors.Is(err, os.ErrExist) || strings.Contains(strings.ToLower(err.Error()), "file exists") || strings.Contains(strings.ToLower(err.Error()), "already exists") {
		return &FileError{Op: op, Path: path, Err: err, Msg: "destination already exists"}
	}

	// Detect "File in Use" / Locked files
	if isLockedError(err) {
		return &FileError{Op: op, Path: path, Err: err, Msg: "file is currently in use by another process"}
	}

	// Detect Permission issues
	if errors.Is(err, os.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "permission denied") {
		return &FileError{Op: op, Path: path, Err: err, Msg: "permission denied"}
	}

	// Detect Disk Full
	if isDiskFullError(err) {
		return &FileError{Op: op, Path: path, Err: err, Msg: "disk is full"}
	}

	// Detect "Not a directory"
	if strings.Contains(strings.ToLower(err.Error()), "not a directory") {
		return &FileError{Op: op, Path: path, Err: err, Msg: "not a directory"}
	}

	// Detect "Is a directory"
	if strings.Contains(strings.ToLower(err.Error()), "is a directory") {
		return &FileError{Op: op, Path: path, Err: err, Msg: "cannot perform this operation on a directory"}
	}

	// Default: Keep original but prefix with operation
	return &FileError{Op: op, Path: path, Err: err}
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
