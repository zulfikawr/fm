package ops

import (
	"context"
	"fmt"
	"time"

	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/files/errors"
)

// Move moves a file or directory. It tries Rename first, and falls back to Copy+Delete if Rename fails.
func Move(ctx context.Context, fs core.FileSystem, src, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	return CrossMove(ctx, fs, fs, src, dst, progChan, policy)
}

// CrossMove moves a file or directory between different filesystems.
func CrossMove(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	if src == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Move", "")
	}

	// Resolve conflict if any
	resolver := conflict.NewResolver()
	resolvedPath, isRenamed, err := resolver.Resolve(ctx, dstFS, src, dst, policy)
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.IsMove = true
			cerr.OpType = "move"
			return cerr
		}
		return err
	}

	if resolvedPath == "" {
		return nil // Skip
	}
	if resolvedPath == dst && policy == conflict.Overwrite {
		_ = dstFS.RemoveAll(ctx, dst)
	}
	dst = resolvedPath

	if progChan != nil && isRenamed {
		select {
		case progChan <- core.Progress{
			Percent: 0,
			Label:   fmt.Sprintf("Moving %s as %s...", srcFS.Base(src), dstFS.Base(dst)),
		}:
		default:
		}
	}

	// 1. Try atomic rename first if same FS
	if srcFS == dstFS {
		err := srcFS.Rename(ctx, src, dst)
		if err == nil {
			if progChan != nil {
				select {
				case progChan <- core.Progress{Percent: 1.0, Label: "Moved " + srcFS.Base(src)}:
				default:
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		}
	}

	// 2. Fallback for cross-device/FS moves: Copy then Delete
	if err := CrossCopy(ctx, srcFS, dstFS, src, dst, progChan, conflict.Overwrite); err != nil {
		_ = dstFS.RemoveAll(ctx, dst)
		return err
	}

	// 3. Verify the copy was successful
	if err := verifyCrossMove(ctx, srcFS, dstFS, src, dst); err != nil {
		_ = dstFS.RemoveAll(ctx, dst)
		return errors.WrapErrorWithPath(fmt.Errorf("move verification failed: %w", err), "CrossMove", src)
	}

	// 4. "Commit" the move by deleting the source
	if err := Delete(ctx, srcFS, src, nil); err != nil {
		return errors.WrapErrorWithPath(fmt.Errorf("move partially successful: items copied and verified but failed to remove from source: %w", err), "CrossMove", src)
	}

	return nil
}

// verifyCrossMove performs a basic check to ensure the source was copied correctly
func verifyCrossMove(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string) error {
	sInfo, err := srcFS.Lstat(ctx, src)
	if err != nil {
		return err
	}

	dInfo, err := dstFS.Lstat(ctx, dst)
	if err != nil {
		return err
	}

	if !sInfo.IsDir() && sInfo.Size() != dInfo.Size() {
		return fmt.Errorf("size mismatch: source %d, destination %d", sInfo.Size(), dInfo.Size())
	}

	return nil
}

// Deprecated: use CrossMove
func verifyMove(ctx context.Context, fs core.FileSystem, src, dst string) error {
	return verifyCrossMove(ctx, fs, fs, src, dst)
}
