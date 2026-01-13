package ops

import (
	"context"
	"fm/internal/files/core"
)

// ConflictError is returned when a destination file already exists
type ConflictError struct {
	Source       string
	Destination  string
	PendingItems []string
	IsMove       bool
}

func (e *ConflictError) Error() string {
	return "destination already exists: " + e.Destination
}

// DeleteMultiple removes multiple files or directories recursively.
func DeleteMultiple(ctx context.Context, fs core.FileSystem, paths []string, useTrash bool, progChan chan<- core.Progress) error {
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
	return nil
}

// Paste copies multiple items from sources to destDir.
func Paste(ctx context.Context, fs core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress) error {
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := fs.Join(destDir, fs.Base(src))

		// Check for conflict
		if _, err := fs.Stat(ctx, dst); err == nil {
			return &ConflictError{
				Source:       src,
				Destination:  dst,
				PendingItems: sources[i+1:],
				IsMove:       false,
			}
		}

		if progChan != nil {
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   "Copying " + fs.Base(src) + "...",
			}:
			default:
			}
		}

		if err := Copy(ctx, fs, src, dst, progChan); err != nil {
			return err
		}
	}
	return nil
}

// MoveMultiple moves multiple items from sources to destDir.
func MoveMultiple(ctx context.Context, fs core.FileSystem, sources []string, destDir string, progChan chan<- core.Progress) error {
	for i, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dst := fs.Join(destDir, fs.Base(src))

		// Check for conflict
		if _, err := fs.Stat(ctx, dst); err == nil {
			return &ConflictError{
				Source:       src,
				Destination:  dst,
				PendingItems: sources[i+1:],
				IsMove:       true,
			}
		}

		if progChan != nil {
			select {
			case progChan <- core.Progress{
				Percent: float64(i) / float64(len(sources)),
				Label:   "Moving " + fs.Base(src) + "...",
			}:
			default:
			}
		}

		if err := Move(ctx, fs, src, dst, progChan); err != nil {
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
