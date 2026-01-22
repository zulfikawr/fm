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

func (r *defaultResolver) Resolve(ctx context.Context, fs core.FileSystem, opts ResolveOptions) (string, bool, error) {
	// Check for same file (only if src is provided)
	if opts.Src != "" {
		sAbs, err := fs.Abs(opts.Src)
		if err != nil {
			return "", false, err
		}
		dAbs, err := fs.Abs(opts.Dst)
		if err != nil {
			return "", false, err
		}
		if sAbs == dAbs && sAbs != "" {
			return "", false, &errors.ValidationError{
				Field:   "destination",
				Value:   opts.Dst,
				Message: "source and destination are the same",
			}
		}
	}

	// Check for conflict
	dInfo, err := fs.Stat(ctx, opts.Dst)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no such file") {
			// No conflict
			return opts.Dst, false, nil
		}
		return "", false, err
	}

	switch opts.Policy {
	case Overwrite:
		// Check if we are trying to overwrite a directory with a file or vice-versa
		if opts.Src != "" {
			sInfo, sErr := fs.Stat(ctx, opts.Src)
			if sErr == nil {
				if sInfo.IsDir() != dInfo.IsDir() {
					// Type mismatch: must remove the destination first
					return opts.Dst, false, nil
				}
			}
		}
		return opts.Dst, false, nil
	case Skip:
		return "", false, nil // Return empty string to indicate skip
	case Rename:
		newName, err := GenerateUniqueName(ctx, fs, opts.Dst)
		return newName, true, err
	case Ask:
		fallthrough
	default:
		return "", false, &ConflictError{
			Source:      opts.Src,
			Destination: opts.Dst,
		}
	}
}
