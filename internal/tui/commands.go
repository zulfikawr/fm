package tui

import (
	"context"
	"time"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

const (
	DirectoryLoadTimeout   = 30 * time.Second
	FileOperationTimeout   = 5 * time.Minute
	MessageDisplayDuration = 3 * time.Second
	SSHConnectionTimeout   = 5 * time.Second
)

// LoadedItemsMsg is sent when directory contents have been loaded asynchronously.
type LoadedItemsMsg struct {
	Generation  int
	Path        string
	Items       []files.Item
	GitStatuses map[string]string
	GitBranch   string
	GitRoot     string
	IsReadOnly  bool
	Err         error
}

// DirSizeMsg is sent when a directory size has been calculated.
type DirSizeMsg struct {
	Path  string
	Size  int64
	MTime time.Time
}

// DirSizesBatchMsg is sent when multiple directory sizes have been calculated.
type DirSizesBatchMsg map[string]int64

// GitStatusMsg is sent when git status has been fetched.
type GitStatusMsg struct {
	Path     string
	Statuses map[string]string
	Branch   string
}

// WatchEventMsg is sent when a file system event occurs in the watched directory.
type WatchEventMsg struct {
	Event fsnotify.Event
	Err   error
}

// WatcherErrorMsg is sent when the watcher encounters an error.
type WatcherErrorMsg struct {
	Err error
}

// WatcherClosedMsg is sent when the watcher is closed unexpectedly.
type WatcherClosedMsg struct{}

type errMsg struct{ err error }

type clearMsg struct{}

// conflictMsg is sent when a destination file already exists.
type conflictMsg struct {
	Src          string
	Dst          string
	PendingItems []string
	IsMove       bool
}

// OperationFinishedMsg is sent when a background file operation completes.
type OperationFinishedMsg struct {
	Paths []string
}

// ProgressMsg is sent to update the progress bar.
type ProgressMsg struct {
	Percent float64
	Label   string
	Channel chan files.Progress
}

// reload triggers an asynchronous reload of the current directory.
func (m *Model) reload() tea.Cmd {
	m.loading = true
	path := m.path
	generation := m.pathGeneration
	mode := m.sortMode
	showHidden := m.cfg.ShowHidden
	fs := m.fs

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), DirectoryLoadTimeout)
		defer cancel()

		items, err := files.Load(ctx, fs, path, mode, showHidden, nil)
		ro, _ := fs.IsReadOnly(ctx, path)
		gitRoot := ""
		if fs.IsLocal() {
			gitRoot = files.GetGitRoot(ctx, path)
		}
		return LoadedItemsMsg{
			Generation:  generation,
			Path:        path,
			Items:       items,
			GitStatuses: nil,
			GitBranch:   "",
			GitRoot:     gitRoot,
			IsReadOnly:  ro,
			Err:         err,
		}
	}
}

func calculateDirSize(fs files.FileSystem, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), DirectoryLoadTimeout)
		defer cancel()

		size := fs.GetDirSize(ctx, path)
		mtime := time.Now()
		if info, err := fs.Stat(ctx, path); err == nil {
			mtime = info.ModTime()
		}
		return DirSizeMsg{Path: path, Size: size, MTime: mtime}
	}
}

func fetchGitStatus(fs files.FileSystem, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), files.GitCommandTimeout)
		defer cancel()

		statuses, branch := fs.GetGitStatus(ctx, path)
		return GitStatusMsg{Path: path, Statuses: statuses, Branch: branch}
	}
}

func (m *Model) watchDir() tea.Cmd {
	if !m.fs.IsLocal() {
		return nil
	}
	return func() tea.Msg {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				// Watcher closed, attempt to restart
				return WatcherClosedMsg{}
			}
			return WatchEventMsg{Event: event}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				// Error channel closed, attempt to restart
				return WatcherClosedMsg{}
			}
			// Return error with restart flag
			return WatcherErrorMsg{Err: err}
		}
	}
}

func (m *Model) restartWatcher() tea.Cmd {
	return func() tea.Msg {
		// Close old watcher if it exists
		if m.watcher != nil {
			m.watcher.Close()
		}

		// Create new watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			m.LogError(err, "Failed to restart watcher")
			return errMsg{err}
		}

		m.watcher = watcher
		m.lastWatched = ""

		// Re-watch current directory
		if m.fs.IsLocal() && m.path != "" {
			if err := m.watcher.Add(m.path); err != nil {
				m.LogError(err, "Failed to watch directory")
				return errMsg{err}
			}
			m.lastWatched = m.path
		}

		// Resume watching
		return m.watchDir()()
	}
}

func (m *Model) setMsg(msg string) tea.Cmd {
	m.msg = msg
	m.msgTime = time.Now()
	return clearMessage()
}

func clearMessage() tea.Cmd {
	return tea.Tick(MessageDisplayDuration, func(time.Time) tea.Msg {
		return clearMsg{}
	})
}

func listenToProgress(progChan chan files.Progress) tea.Cmd {
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

func deleteItems(fs files.FileSystem, targets []string, useTrash bool) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, target := range targets {
				select {
				case <-ctx.Done():
					return errMsg{ctx.Err()}
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
					err = files.Trash(ctx, fs, target)
				} else {
					err = files.Delete(ctx, fs, target, nil)
				}
				if err != nil {
					return errMsg{err}
				}
			}
			return OperationFinishedMsg{Paths: targets}
		},
	)
}

func pasteItems(fs files.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					return errMsg{ctx.Err()}
				default:
				}

				dst := fs.Join(destDir, fs.Base(src))

				// Check for conflict
				if _, err := fs.Stat(ctx, dst); err == nil {
					return conflictMsg{
						Src:          src,
						Dst:          dst,
						PendingItems: sources[i+1:],
						IsMove:       false,
					}
				}

				// If it's a single big file, Copy will send updates to progChan.
				// If many files, we manually update for each file.
				select {
				case progChan <- files.Progress{
					Percent: float64(i) / float64(len(sources)),
					Label:   "Copying " + fs.Base(src) + "...",
				}:
				default:
				}

				if err := files.Copy(ctx, fs, src, dst, progChan); err != nil {
					return errMsg{err}
				}
			}
			return OperationFinishedMsg{Paths: sources}
		},
	)
}

func moveItems(fs files.FileSystem, sources []string, destDir string) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			for i, src := range sources {
				select {
				case <-ctx.Done():
					return errMsg{ctx.Err()}
				default:
				}

				dst := fs.Join(destDir, fs.Base(src))

				// Check for conflict
				if _, err := fs.Stat(ctx, dst); err == nil {
					return conflictMsg{
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

				if err := files.Move(ctx, fs, src, dst, progChan); err != nil {
					return errMsg{err}
				}
			}
			return OperationFinishedMsg{Paths: sources}
		},
	)
}

func overwriteItem(fs files.FileSystem, src, dst string, isMove bool) tea.Cmd {
	progChan := make(chan files.Progress, 100)
	return tea.Batch(
		listenToProgress(progChan),
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), FileOperationTimeout)
			defer cancel()
			defer close(progChan)

			var err error
			if isMove {
				err = files.Move(ctx, fs, src, dst, progChan)
			} else {
				err = files.Copy(ctx, fs, src, dst, progChan)
			}

			if err != nil {
				return errMsg{err}
			}
			return OperationFinishedMsg{Paths: []string{src, dst}}
		},
	)
}
