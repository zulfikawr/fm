package ops

import (
	"context"
	"fmt"

	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/files/errors"
)

// Rename moves or renames a file or directory.
func Rename(ctx context.Context, fs core.FileSystem, oldPath, newPath string, policy conflict.Policy) error {
	if oldPath == "" || newPath == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("empty path"), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
	}
	// Validate new filename component
	if err := ValidateFileName(fs.Base(newPath)); err != nil {
		return errors.WrapErrorWithPath(err, "Rename", newPath)
	}

	resolver := conflict.NewResolver()
	resolvedPath, _, err := resolver.Resolve(ctx, fs, oldPath, newPath, policy)
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.OpType = "rename"
			return cerr
		}
		return err
	}

	if resolvedPath == "" {
		return nil // Skip
	}
	newPath = resolvedPath

	return errors.WrapErrorWithPath(fs.Rename(ctx, oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}
