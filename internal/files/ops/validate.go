package ops

import (
	"context"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

// ValidatePath checks if a path is safe and doesn't contain traversal attempts
func ValidatePath(fs core.FileSystem, basePath, targetName string) error {
	if err := ValidateFileName(targetName); err != nil {
		return err
	}

	// Additional check: ensure the resolved path is within base
	if basePath != "" {
		testPath := fs.Join(basePath, targetName)

		// Ensure the result is still under basePath
		relPath, err := fs.Rel(basePath, testPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return &errors.ValidationError{
				Field:   "path",
				Value:   targetName,
				Message: "invalid path: attempts to escape base directory",
			}
		}
	}

	return nil
}

// ValidateSafePath checks if an arbitrary path is safe (doesn't try to escape via ..)
func ValidateSafePath(path string) error {
	if path == "" {
		return &errors.ValidationError{
			Field:   "path",
			Value:   path,
			Message: "path is empty",
		}
	}

	return nil
}

// ValidateFileName checks if a filename is valid for rename/create operations
func ValidateFileName(name string) error {
	if name == "" || name == "." || name == ".." {
		return &errors.ValidationError{
			Field:   "filename",
			Value:   name,
			Message: "cannot be empty, '.' or '..'",
		}
	}

	if len(name) > constants.MaxFilenameLength {
		return &errors.ValidationError{
			Field:   "filename",
			Value:   name,
			Message: "filename too long",
		}
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return &errors.ValidationError{
			Field:   "filename",
			Value:   name,
			Message: "filename contains invalid characters",
		}
	}

	return nil
}

// ValidateSearchQuery ensures the search string doesn't contain dangerous characters
// that could be misused if search is later expanded to shell commands.
func ValidateSearchQuery(query string) error {
	if strings.ContainsAny(query, "`$;") {
		return &errors.ValidationError{
			Field:   "query",
			Value:   query,
			Message: "search query contains invalid characters",
		}
	}
	return nil
}

// ValidateWritable checks if a path is on a writable filesystem
func ValidateWritable(ctx context.Context, fs core.FileSystem, path string) error {
	ro, err := fs.IsReadOnly(ctx, path)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ValidateWritable", path)
	}
	if ro {
		return &errors.PermissionError{
			Path:      path,
			Operation: "write",
		}
	}
	return nil
}
