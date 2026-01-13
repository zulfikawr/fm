package ops

import (
	"context"
	"fmt"

	"fm/internal/files/core"
	"fm/internal/files/errors"
)

// Rename moves or renames a file or directory.
func Rename(ctx context.Context, fs core.FileSystem, oldPath, newPath string) error {
	if oldPath == "" || newPath == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
	}
	// Validate new filename component
	if err := ValidateFileName(fs.Base(newPath)); err != nil {
		return errors.WrapErrorWithPath(err, "Rename", newPath)
	}
	return errors.WrapErrorWithPath(fs.Rename(ctx, oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}
