package commands

import (
	"context"

	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/ops"

	tuierrors "fm/internal/tui/errors"

	tea "github.com/charmbracelet/bubbletea"
)

// DeleteItems returns a command to delete the specified targets.
func DeleteItems(fs files.FileSystem, targets []string, useTrash bool) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), constants.FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, target := range targets {
				select {
				case <-ctx.Done():
					// Timeout is a transient error
					err := tuierrors.TransientError("delete items", "Operation timed out", 3).
						WithContext("target", target).
						WithContext("targets_remaining", len(targets)-i)
					return ErrorMsg{Err: err}
				default:
				}

				if !useTrash {
					select {
					case progChan <- files.Progress{
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

// PasteItems returns a command to copy items from sources to destDir.
func PasteItems(fs files.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), constants.FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					err := tuierrors.TransientError("paste items", "Operation timed out", 3).
						WithContext("source", src).
						WithContext("destination", destDir).
						WithContext("items_remaining", len(sources)-i)
					return ErrorMsg{Err: err}
				default:
				}

				dst := fs.Join(destDir, fs.Base(src))

				// Check for conflict - channel will be closed by defer
				if _, err := fs.Stat(ctx, dst); err == nil {
					return ConflictMsg{
						Src:          src,
						Dst:          dst,
						PendingItems: sources[i+1:],
						IsMove:       false,
					}
				}

				select {
				case progChan <- files.Progress{
					Percent: float64(i) / float64(len(sources)),
					Label:   "Copying " + fs.Base(src) + "...",
				}:
				default:
				}

				if err := ops.Copy(ctx, fs, src, dst, progChan); err != nil {
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

// MoveItems returns a command to move items from sources to destDir.
func MoveItems(fs files.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), constants.FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					err := tuierrors.TransientError("move items", "Operation timed out", 3).
						WithContext("source", src).
						WithContext("destination", destDir).
						WithContext("items_remaining", len(sources)-i)
					return ErrorMsg{Err: err}
				default:
				}

				dst := fs.Join(destDir, fs.Base(src))

				// Check for conflict - channel will be closed by defer
				if _, err := fs.Stat(ctx, dst); err == nil {
					return ConflictMsg{
						Src:          src,
						Dst:          dst,
						PendingItems: sources[i+1:],
						IsMove:       true,
					}
				}

				select {
				case progChan <- files.Progress{
					Percent: float64(i) / float64(len(sources)),
					Label:   "Moving " + fs.Base(src) + "...",
				}:
				default:
				}

				if err := ops.Move(ctx, fs, src, dst, progChan); err != nil {
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
func OverwriteItem(fs files.FileSystem, src, dst string, isMove bool) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		ListenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), constants.FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			var err error
			if isMove {
				err = ops.Move(ctx, fs, src, dst, progChan)
			} else {
				err = ops.Copy(ctx, fs, src, dst, progChan)
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
func ListenToProgress(progChan chan files.Progress) tea.Cmd {
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
