package conflict

import (
	"context"
	"fm/internal/files/core"
	"fmt"
)

type defaultResolver struct{}

// NewResolver returns a default implementation of Resolver
func NewResolver() Resolver {
	return &defaultResolver{}
}

func (r *defaultResolver) Resolve(ctx context.Context, fs core.FileSystem, src, dst string, policy Policy) (string, bool, error) {
	// Check for same file
	sAbs, _ := fs.Abs(src)
	dAbs, _ := fs.Abs(dst)
	if sAbs == dAbs && sAbs != "" {
		return "", false, fmt.Errorf("source and destination are the same: %s", src)
	}

	// Check for conflict
	dInfo, err := fs.Stat(ctx, dst)
	if err != nil {
		// No conflict
		return dst, false, nil
	}

	switch policy {
	case Overwrite:
		// Check if we are trying to overwrite a directory with a file or vice-versa
		sInfo, sErr := fs.Stat(ctx, src)
		if sErr == nil {
			if sInfo.IsDir() != dInfo.IsDir() {
				// Type mismatch: must remove the destination first
				return dst, false, nil
			}
		}
		return dst, false, nil
	case Skip:
		return "", false, nil // Return empty string to indicate skip
	case Rename:
		newName, err := GenerateUniqueName(ctx, fs, dst)
		return newName, true, err
	case Ask:
		fallthrough
	default:
		return "", false, &ConflictError{
			Source:      src,
			Destination: dst,
		}
	}
}
