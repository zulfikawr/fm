package ops

import (
	"context"
	"fmt"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
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

// CopyMultiple copies multiple items from sources to destDir between different filesystems.
func CopyMultiple(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress, policy conflict.Policy, applyToAll bool) error {
	if err := ValidateWritable(ctx, dstFS, destDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := dstFS.Join(destDir, srcFS.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(ctx, dstFS, src, dst, policy)
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				cerr.PendingItems = sources[i:]
				cerr.OpType = "copy"
				return cerr
			}
			return err
		}

		if resolvedDst == "" {
			// If we skipped and not applying to all, reset policy for next items
			if !applyToAll {
				policy = conflict.Ask
			}
			continue // Skip
		}
		dst = resolvedDst

		if progChan != nil {
			label := "Copying " + srcFS.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Copying %s as %s...", srcFS.Base(src), dstFS.Base(dst))
			}
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   label,
			}:
			default:
			}
		}

		if err := CrossCopy(ctx, srcFS, dstFS, src, dst, progChan, conflict.Overwrite); err != nil {
			return err
		}

		// If we are not applying to all, reset policy to Ask after the first successful resolution
		if !applyToAll {
			policy = conflict.Ask
		}
	}
	return nil
}

// MoveMultiple moves multiple items from sources to destDir between different filesystems.
func MoveMultiple(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress, policy conflict.Policy, applyToAll bool) error {
	if err := ValidateWritable(ctx, dstFS, destDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := dstFS.Join(destDir, srcFS.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(ctx, dstFS, src, dst, policy)
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
			if !applyToAll {
				policy = conflict.Ask
			}
			continue // Skip
		}
		dst = resolvedDst

		if progChan != nil {
			label := "Moving " + srcFS.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Moving %s as %s...", srcFS.Base(src), dstFS.Base(dst))
			}
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   label,
			}:
			default:
			}
		}

		if err := CrossMove(ctx, srcFS, dstFS, src, dst, progChan, conflict.Overwrite); err != nil {
			return err
		}

		if !applyToAll {
			policy = conflict.Ask
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
