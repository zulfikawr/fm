package conflict

import (
	"context"
	"os"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

type defaultResolver struct{}

// NewResolver returns a default implementation of Resolver
func NewResolver() Resolver {
	return &defaultResolver{}
}

func (r *defaultResolver) Resolve(ctx context.Context, fs core.FileSystem, src, dst string, policy Policy) (string, bool, error) {
	// Check for same file (only if src is provided)
	if src != "" {
		sAbs, err := fs.Abs(src)
		if err != nil {
			return "", false, err
		}
		dAbs, err := fs.Abs(dst)
		if err != nil {
			return "", false, err
		}
		if sAbs == dAbs && sAbs != "" {
			return "", false, &errors.ValidationError{
				Field:   "destination",
				Value:   dst,
				Message: "source and destination are the same",
			}
		}
	}

	// Check for conflict
	dInfo, err := fs.Stat(ctx, dst)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no such file") {
			// No conflict
			return dst, false, nil
		}
		return "", false, err
	}

	switch policy {
	case Overwrite:
		// Check if we are trying to overwrite a directory with a file or vice-versa
		if src != "" {
			sInfo, sErr := fs.Stat(ctx, src)
			if sErr == nil {
				if sInfo.IsDir() != dInfo.IsDir() {
					// Type mismatch: must remove the destination first
					return dst, false, nil
				}
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
