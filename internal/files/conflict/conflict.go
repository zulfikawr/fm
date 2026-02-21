// Package conflict handles file operation conflicts and resolution strategies.
// It provides policies for handling existing files (overwrite, skip, rename, ask)
// and includes security validation for path traversal attacks.
package conflict

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
)

// Policy defines how to handle existing destination files
type Policy int

const (
	Ask Policy = iota
	Overwrite
	Skip
	Rename
)

// ConflictError is returned when a destination file already exists and policy is Ask
type ConflictError struct {
	Source       string
	Destination  string
	PendingItems []string
	IsMove       bool
	OpType       string // copy, move, zip, unzip
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("destination already exists: %s", e.Destination)
}

// ResolveOptions encapsulates data for conflict resolution
type ResolveOptions struct {
	Src    string
	Dst    string
	Policy Policy
}

// Resolver handles conflict resolution logic
type Resolver interface {
	Resolve(ctx context.Context, fs core.FileSystem, opts ResolveOptions) (string, bool, error)
}

// ValidateSecurePath ensures that a target path is securely contained within a base directory.
// This prevents ZipSlip and other directory traversal attacks.
func ValidateSecurePath(fs core.FileSystem, baseDir, targetPath string) (string, error) {
	// Normalize path separators and clean the target path
	targetPath = filepath.ToSlash(targetPath)
	targetPath = filepath.Clean(targetPath)

	// Block absolute paths
	if filepath.IsAbs(targetPath) {
		return "", &errors.ValidationError{
			Field:   "targetPath",
			Value:   targetPath,
			Message: "Security block: Absolute paths not allowed",
		}
	}

	fullPath := fs.Join(baseDir, targetPath)
	cleanBase, err := fs.Abs(baseDir)
	if err != nil {
		return "", errors.WrapError(err, "ValidateSecurePath")
	}

	// Resolve symlinks to prevent symlink-based traversal
	cleanTarget, err := fs.Abs(fullPath)
	if err != nil {
		return "", errors.WrapError(err, "ValidateSecurePath")
	}

	// Additional check: evaluate symlinks if local filesystem
	if fs.IsLocal() {
		if resolved, err := filepath.EvalSymlinks(cleanTarget); err == nil {
			cleanTarget = resolved
		}
	}

	rel, err := fs.Rel(cleanBase, cleanTarget)
	if err != nil {
		return "", errors.WrapError(err, "ValidateSecurePath")
	}

	// Block any path that tries to escape the base directory
	if strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, fs.Separator()) || filepath.IsAbs(rel) {
		return "", &errors.ValidationError{
			Field:   "targetPath",
			Value:   targetPath,
			Message: "Security block: Cannot write files outside the destination directory",
		}
	}

	return cleanTarget, nil
}

// GenerateUniqueName generates a unique filename by appending a suffix if it exists.
// e.g., "file.txt" -> "file (1).txt" -> "file (2).txt"
func GenerateUniqueName(ctx context.Context, fs core.FileSystem, path string) (string, error) {
	if info, err := fs.Stat(ctx, path); err != nil {
		return path, nil // Path doesn't exist, it's unique
	} else if info == nil {
		logger.Warnf("GenerateUniqueName: stat returned nil info for path %q", path)
	}

	ext := fs.Ext(path)
	base := path[:len(path)-len(ext)]

	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if info, err := fs.Stat(ctx, newName); err != nil {
			return newName, nil
		} else if info == nil {
			logger.Warnf("GenerateUniqueName: stat returned nil info for candidate name %q", newName)
		}
		// Safety break to prevent infinite loop in case of errors other than NotExist
		if i > 10000 {
			return "", &errors.FileError{
				Op:   "GenerateUniqueName",
				Path: path,
				Msg:  "Rename failed: Could not automatically generate a unique name after multiple attempts",
			}
		}
	}
}
