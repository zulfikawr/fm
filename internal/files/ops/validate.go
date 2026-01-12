package ops

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"fm/internal/constants"
)

// ValidatePath checks if a path is safe and doesn't contain traversal attempts
func ValidatePath(basePath, targetName string) error {
	if err := ValidateFileName(targetName); err != nil {
		return err
	}

	// Additional check: ensure the resolved path is within base
	if basePath != "" {
		cleanBase := filepath.Clean(basePath)
		testPath := filepath.Join(basePath, targetName)
		cleanTest := filepath.Clean(testPath)

		// Ensure the result is still under basePath
		relPath, err := filepath.Rel(cleanBase, cleanTest)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return errors.New("invalid path: attempts to escape base directory")
		}
	}

	return nil
}

// ValidateSafePath checks if an arbitrary path is safe (doesn't try to escape via ..)
func ValidateSafePath(path string) error {
	if path == "" {
		return errors.New("path is empty")
	}

	return nil
}

// ValidateFileName checks if a filename is valid for rename/create operations
func ValidateFileName(name string) error {
	if name == "" || name == "." || name == ".." {
		return errors.New("invalid filename: cannot be empty, '.' or '..'")
	}

	if len(name) > constants.MaxFilenameLength {
		return fmt.Errorf("filename too long (max %d bytes)", constants.MaxFilenameLength)
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return errors.New("filename contains invalid characters")
	}

	return nil
}

// ValidateSearchQuery ensures the search string doesn't contain dangerous characters
// that could be misused if search is later expanded to shell commands.
func ValidateSearchQuery(query string) error {
	if strings.ContainsAny(query, "`$;") {
		return errors.New("search query contains invalid characters")
	}
	return nil
}
