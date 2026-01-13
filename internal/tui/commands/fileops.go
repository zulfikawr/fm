package commands

import (
	"context"

	"fm/internal/files/core"
	"fm/internal/files/ops"

	tuierrors "fm/internal/tui/errors"

	tea "github.com/charmbracelet/bubbletea"
)

// DeleteItems returns a command to delete the specified targets.
func DeleteItems(ctx context.Context, fs core.FileSystem, targets []string, useTrash bool) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)

			for i, target := range targets {
				select {
				case <-ctx.Done():
					return OperationFinishedMsg{Paths: targets}
				default:
				}

				if !useTrash {
					select {
					case progChan <- core.Progress{
						Percent: float64(i) / float64(len(targets)),
						Label:   "Deleting " + fs.Base(target) + "...",
					}:
					default:
					}
				}

				var err error
				if useTrash {
					err = ops.Trash(ctx, fs, target)
				} else {
					err = ops.Delete(ctx, fs, target, nil)
				}
				if err != nil {
					// Wrap delete error with context
					sysErr := tuierrors.SystemError("delete file", err).
						WithContext("path", target).
						WithContext("use_trash", useTrash)
					return ErrorMsg{Err: sysErr}
				}
			}
			return OperationFinishedMsg{Paths: targets}
		},
	)
}

// PasteItems returns a command to copy items from srcFS to dstFS.
func PasteItems(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					return OperationFinishedMsg{Paths: sources}
				default:
				}

				dst := dstFS.Join(destDir, srcFS.Base(src))

				// Check for conflict
				if _, err := dstFS.Stat(ctx, dst); err == nil {
					return ConflictMsg{
						Src:          src,
						Dst:          dst,
						PendingItems: sources[i+1:],
						IsMove:       false,
					}
				}

				select {
				case progChan <- core.Progress{
					Percent: float64(i) / float64(len(sources)),
					Label:   "Copying " + srcFS.Base(src) + "...",
				}:
				}

				if err := ops.CrossCopy(ctx, srcFS, dstFS, src, dst, progChan); err != nil {
					sysErr := tuierrors.SystemError("copy file", err).
						WithContext("from", src).
						WithContext("to", dst)
					return ErrorMsg{Err: sysErr}
				}
			}
			return OperationFinishedMsg{Paths: sources}
		},
	)
}

// MoveItems returns a command to move items from srcFS to dstFS.
func MoveItems(ctx context.Context, srcFS, dstFS core.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					return OperationFinishedMsg{Paths: sources}
				default:
				}

				dst := dstFS.Join(destDir, srcFS.Base(src))

				// Check for conflict
				if _, err := dstFS.Stat(ctx, dst); err == nil {
					return ConflictMsg{
						Src:          src,
						Dst:          dst,
						PendingItems: sources[i+1:],
						IsMove:       true,
					}
				}

				select {
				case progChan <- core.Progress{
					Percent: float64(i) / float64(len(sources)),
					Label:   "Moving " + srcFS.Base(src) + "...",
				}:
				}

				if err := ops.CrossMove(ctx, srcFS, dstFS, src, dst, progChan); err != nil {
					sysErr := tuierrors.SystemError("move file", err).
						WithContext("from", src).
						WithContext("to", dst)
					return ErrorMsg{Err: sysErr}
				}
			}
			return OperationFinishedMsg{Paths: sources}
		},
	)
}

// OverwriteItem returns a command to overwrite a file by copying or moving.
func OverwriteItem(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, isMove bool) tea.Cmd {
	progChan := make(chan core.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			defer close(progChan)

			var err error
			if isMove {
				err = ops.CrossMove(ctx, srcFS, dstFS, src, dst, progChan)
			} else {
				err = ops.CrossCopy(ctx, srcFS, dstFS, src, dst, progChan)
			}

			if err != nil {
				operation := "copy file"
				if isMove {
					operation = "move file"
				}
				sysErr := tuierrors.SystemError(operation, err).
					WithContext("from", src).
					WithContext("to", dst).
					WithContext("overwrite", true)
				return ErrorMsg{Err: sysErr}
			}
			return OperationFinishedMsg{Paths: []string{src, dst}}
		},
	)
}

// ListenToProgress returns a command that listens for progress updates on a channel.
func ListenToProgress(progChan chan core.Progress) tea.Cmd {
	return func() tea.Msg {
		prog, ok := <-progChan
		if !ok {
			return nil
		}
		return ProgressMsg{
			Percent: prog.Percent,
			Label:   prog.Label,
			Channel: progChan,
		}
	}
}
