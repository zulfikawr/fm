package nav

import (
	"context"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/archive"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/listing"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/logger"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"
)

// Reload triggers an asynchronous reload of the current directory.
func Reload(m *tui_context.Model, silent bool) tea.Cmd {
	path := m.Navigation.Path
	gen := m.Navigation.PathGen
	fs := m.FS
	gs := m.GS
	ctx := m.Context
	mode := m.Display.SortMode
	showHidden := m.Config.UI.ShowHidden
	sizeFormatIdx := m.Config.UI.SizeFormatIndex
	dateFormatIdx := m.Config.UI.DateFormatIndex

	if items, ok := m.Cache.ItemCache.Get(path); ok {
		m.Navigation.Items = items
		ApplyFilter(m)

		if val, ok := m.Cache.CursorMemory.Get(path); ok {
			m.Navigation.Cursor = val
		} else {
			m.Navigation.Cursor = 0
		}
		if val, ok := m.Cache.OffsetMemory.Get(path); ok {
			m.Navigation.Offset = val
		} else {
			m.Navigation.Offset = 0
		}
		SyncOffset(m)

		m.UI.Loading = false
		silent = true
	} else {
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
	}

	if !silent {
		m.UI.Loading = true
	}

	if len(m.Navigation.Items) == 0 && !core.IsRoot(fs, path) {
		m.Navigation.Items = []core.Item{{
			Name:  "↑ ..",
			IsDir: true,
			State: core.ItemState{
				IsUp:      true,
				SearchKey: "..",
			},
			Path: core.GetParent(fs, path),
		}}
		ApplyFilter(m)
	}

	if m.Navigation.Git.CancelFunc != nil {
		m.Navigation.Git.CancelFunc()
	}

	loadSkeletonCmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, constants.DirectoryLoadTimeout)
		defer cancel()

		var gitStatuses map[string]string
		var gitRoot string
		var branch string
		var modified, staged, untracked int

		if gs.IsEnabled() {
			if root, ok := m.Cache.GitRootCache.Get(path); ok {
				gitRoot = root
			} else {
				gitRoot = gs.GetRoot(ctx, path)
				if gitRoot != "" {
					m.Cache.GitRootCache.Put(path, gitRoot)
				}
			}
			gitStatuses, branch, modified, staged, untracked = gs.GetStatus(ctx, path)
		}

		items, err := listing.LoadSkeleton(ctx, listing.LoadOptions{
			FS:          fs,
			Path:        path,
			ShowHidden:  showHidden,
			GitStatuses: gitStatuses,
		})
		if err != nil {
			return messages.LoadedItemsMsg{Generation: gen, Path: path, Err: err, GitRoot: gitRoot}
		}

		for i := range items {
			items[i].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
		}

		sorting.SortItems(items, mode, true)

		return messages.PartialItemsMsg{
			Generation: gen,
			Path:       path,
			Items:      items,
			GitRoot:    gitRoot,
			Branch:     branch,
			Modified:   modified,
			Staged:     staged,
			Untracked:  untracked,
		}
	}

	_, gitCancel := context.WithCancel(ctx)
	m.Navigation.Git.CancelFunc = gitCancel

	return loadSkeletonCmd
}

func HandlePartialItems(m *tui_context.Model, msg messages.PartialItemsMsg) tea.Cmd {
	if msg.Generation != m.Navigation.PathGen {
		return nil
	}

	m.UI.Loading = false
	m.Navigation.Items = msg.Items
	m.Navigation.FilteredItems = msg.Items
	m.Navigation.Git.Branch = msg.Branch
	m.Navigation.Git.Root = msg.GitRoot
	m.Navigation.Git.Modified = msg.Modified
	m.Navigation.Git.Staged = msg.Staged
	m.Navigation.Git.Untracked = msg.Untracked

	if val, ok := m.Cache.CursorMemory.Get(m.Navigation.Path); ok {
		m.Navigation.Cursor = val
	}
	if val, ok := m.Cache.OffsetMemory.Get(m.Navigation.Path); ok {
		m.Navigation.Offset = val
	}
	SyncOffset(m)

	return fetchMetadata(m)
}

func fetchMetadata(m *tui_context.Model) tea.Cmd {
	path := m.Navigation.Path
	gen := m.Navigation.PathGen
	fs := m.FS
	items := m.Navigation.Items
	offset := m.Navigation.Offset
	height := m.Display.ViewportHeight
	if height <= 0 {
		height = 20
	}

	sizeFormatIdx := m.Config.UI.SizeFormatIndex
	dateFormatIdx := m.Config.UI.DateFormatIndex

	// Capture current Git state to pass through
	gitBranch := m.Navigation.Git.Branch
	gitRoot := m.Navigation.Git.Root
	gitMod := m.Navigation.Git.Modified
	gitStaged := m.Navigation.Git.Staged
	gitUntracked := m.Navigation.Git.Untracked

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		buffer := 10
		start := offset - buffer
		if start < 0 {
			start = 0
		}
		end := offset + height + buffer
		if end > len(items) {
			end = len(items)
		}

		updatedItems := make([]core.Item, len(items))
		copy(updatedItems, items)

		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10)

		for i := range updatedItems {
			idx := i
			if updatedItems[idx].State.IsUp || updatedItems[idx].Display.IsGhost || updatedItems[idx].State.HasMetadata {
				continue
			}

			if idx < start || idx >= end {
				continue
			}

			g.Go(func() error {
				info, err := fs.Stat(ctx, updatedItems[idx].Path)
				if err == nil {
					updatedItems[idx] = core.NewItem(info, updatedItems[idx].Path, updatedItems[idx].Display.GitStatus)
					if updatedItems[idx].IsDir {
						listing.EnrichMetadata(ctx, fs, &updatedItems[idx])
					}
					updatedItems[idx].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
				} else {
					updatedItems[idx].State.HasMetadata = true
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			logger.Debugf("Error fetching priority metadata: %v", err)
		}

		g, ctx = errgroup.WithContext(ctx)
		g.SetLimit(10)

		for i := range updatedItems {
			idx := i
			if updatedItems[idx].State.IsUp || updatedItems[idx].Display.IsGhost || updatedItems[idx].State.HasMetadata {
				continue
			}

			g.Go(func() error {
				info, err := fs.Stat(ctx, updatedItems[idx].Path)
				if err == nil {
					updatedItems[idx] = core.NewItem(info, updatedItems[idx].Path, updatedItems[idx].Display.GitStatus)
					if updatedItems[idx].IsDir {
						listing.EnrichMetadata(ctx, fs, &updatedItems[idx])
					}
					updatedItems[idx].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
				} else {
					updatedItems[idx].State.HasMetadata = true
				}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			logger.Debugf("Error fetching remaining metadata: %v", err)
		}

		ro, err := fs.IsReadOnly(ctx, path)
		if err != nil {
			logger.Debugf("Failed to check if path %s is read-only: %v", path, err)
		}

		return messages.LoadedItemsMsg{
			Generation: gen,
			Path:       path,
			Items:      updatedItems,
			IsReadOnly: ro,
			GitBranch:  gitBranch,
			GitRoot:    gitRoot,
			Modified:   gitMod,
			Staged:     gitStaged,
			Untracked:  gitUntracked,
		}
	}
}

func FinalizeDirectoryLoad(m *tui_context.Model, msg messages.LoadedItemsMsg) tea.Cmd {
	m.UI.Loading = false
	if msg.Generation != m.Navigation.PathGen {
		return nil
	}

	if msg.Err != nil {
		m.Navigation.Items = []core.Item{}
		m.Navigation.FilteredItems = []core.Item{}

		if m.Navigation.Path == msg.Path {
			if !core.IsRoot(m.FS, m.Navigation.Path) {
				m.Navigation.Path = core.GetParent(m.FS, m.Navigation.Path)
				m.Navigation.PathGen++
				return Reload(m, false)
			}
		}
		return func() tea.Msg { return messages.ErrorMsg{Err: msg.Err} }
	}

	m.Navigation.Items = msg.Items
	ApplyFilter(m)
	m.Navigation.Git.Branch = msg.GitBranch
	m.Navigation.Git.Root = msg.GitRoot
	m.Navigation.Git.Modified = msg.Modified
	m.Navigation.Git.Staged = msg.Staged
	m.Navigation.Git.Untracked = msg.Untracked
	m.Display.ReadOnly = msg.IsReadOnly

	if !msg.Cached && msg.Err == nil {
		m.Cache.ItemCache.Put(msg.Path, msg.Items)
	}

	if val, ok := m.Cache.CursorMemory.Get(msg.Path); ok {
		m.Navigation.Cursor = val
	}
	if val, ok := m.Cache.OffsetMemory.Get(msg.Path); ok {
		m.Navigation.Offset = val
	}

	if len(m.Navigation.FilteredItems) > 0 {
		if m.Navigation.Cursor >= len(m.Navigation.FilteredItems) {
			m.Navigation.Cursor = len(m.Navigation.FilteredItems) - 1
		}
		if m.Navigation.Cursor < 0 {
			m.Navigation.Cursor = 0
		}
	} else {
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
	}

	SyncOffset(m)
	return func() tea.Msg { return messages.WatchDirMsg{} }
}

func EnterArchive(m *tui_context.Model, selected core.Item) tea.Cmd {
	m.UI.Loading = true
	archivePath := selected.Path

	return func() tea.Msg {
		afs, err := archive.NewArchiveFS(archivePath)
		if err != nil {
			return messages.ErrorMsg{Err: err}
		}
		return messages.ArchiveEnteredMsg{
			FS:         afs,
			ParentFS:   m.FS,
			ParentPath: m.Navigation.Path,
		}
	}
}
