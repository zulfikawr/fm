package ops

import (
	"context"
	"fmt"

	"fm/internal/files/conflict"
	"fm/internal/files/core"
)

// DeleteMultiple removes multiple files or directories recursively.
func DeleteMultiple(ctx context.Context, fs core.FileSystem, paths []string, useTrash bool, progChan chan<- core.Progress) error {
	if len(paths) > 0 {
		if err := ValidateWritable(ctx, fs, fs.Dir(paths[0])); err != nil {
			return err
		}
	}
	for i, path := range paths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if progChan != nil && !useTrash {
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(paths)),
				Label:   "Deleting " + fs.Base(path) + "...",
			}:
			default:
			}
		}

		var err error
		if useTrash {
			err = Trash(ctx, fs, path)
		} else {
			err = Delete(ctx, fs, path, nil)
		}
		if err != nil {
			return err
		}
	}

	if progChan != nil && len(paths) > 0 {
		label := ""
		if len(paths) == 1 {
			label = fmt.Sprintf("Deleted %s", fs.Base(paths[0]))
		} else {
			label = fmt.Sprintf("Deleted %d items", len(paths))
		}
		select {
		case progChan <- core.Progress{
			Percent: 1.0,
			Label:   label,
		}:
		default:
		}
	}

	return nil
}

// CopyMultiple copies multiple items from sources to destDir.
func CopyMultiple(ctx context.Context, fs core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress, policy conflict.Policy) error {
	if err := ValidateWritable(ctx, fs, destDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := fs.Join(destDir, fs.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(ctx, fs, src, dst, policy)
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				cerr.PendingItems = sources[i:]
				cerr.OpType = "copy"
				return cerr
			}
			return err
		}

		if resolvedDst == "" {
			continue // Skip
		}
		dst = resolvedDst

		if progChan != nil {
			label := "Copying " + fs.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Copying %s as %s...", fs.Base(src), fs.Base(dst))
			}
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   label,
			}:
			default:
			}
		}

		if err := Copy(ctx, fs, src, dst, progChan, conflict.Overwrite); err != nil {
			return err
		}
	}
	return nil
}

// MoveMultiple moves multiple items from sources to destDir.
func MoveMultiple(ctx context.Context, fs core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress, policy conflict.Policy) error {
	if err := ValidateWritable(ctx, fs, destDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := fs.Join(destDir, fs.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(ctx, fs, src, dst, policy)
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				cerr.PendingItems = sources[i:]
				cerr.IsMove = true
				cerr.OpType = "move"
				return cerr
			}
			return err
		}

		if resolvedDst == "" {
			continue // Skip
		}
		dst = resolvedDst

		if progChan != nil {
			label := "Moving " + fs.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Moving %s as %s...", fs.Base(src), fs.Base(dst))
			}
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   label,
			}:
			default:
			}
		}

		if err := Move(ctx, fs, src, dst, progChan, conflict.Overwrite); err != nil {
			return err
		}
	}
	return nil
}

// CheckAndMarkProcessing checks if any of the paths are currently being processed
// and marks them as processing if none are. Returns false if any path is already processing.
func CheckAndMarkProcessing(processing map[string]bool, paths []string) bool {
	for _, p := range paths {
		if processing[p] {
			return false
		}
	}
	for _, p := range paths {
		processing[p] = true
	}
	return true
}
