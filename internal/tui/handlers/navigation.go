package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	fileerrors "fm/internal/files/errors"
	"fm/internal/files/listing"
	"fm/internal/files/local"
	"fm/internal/files/ops"
	"fm/internal/files/sorting"
	"fm/internal/tui/components/file"
	tui_context "fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"
)

// HandleNavigation handles navigation-related messages
func HandleNavigation(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleNavKeys(m, msg)
	case PartialItemsMsg:
		return handlePartialItems(m, msg)
	case LoadedItemsMsg:
		return finalizeDirectoryLoad(m, msg)
	}
	return nil
}

func handlePartialItems(m *tui_context.Model, msg PartialItemsMsg) tea.Cmd {
	if msg.Generation != m.Navigation.PathGen {
		return nil
	}

	m.UI.Loading = false
	m.Navigation.Items = msg.Items
	m.Navigation.FilteredItems = msg.Items

	// Restore cursor/offset from memory if available
	if val, ok := m.Cache.CursorMemory.Get(m.Navigation.Path); ok {
		m.Navigation.Cursor = val
	}
	if val, ok := m.Cache.OffsetMemory.Get(m.Navigation.Path); ok {
		m.Navigation.Offset = val
	}
	syncOffset(m)

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

	sizeFormatIdx := m.Config.SizeFormatIndex
	dateFormatIdx := m.Config.DateFormatIndex

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 1. Determine priority range (visible + buffer)
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

		// Create a limited concurrency group for fetching
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(10) // Limit concurrent SFTP requests to be polite

		// 2. Fetch for priority range first
		for i := range updatedItems {
			idx := i
			// Skip if already has metadata or is special item
			if updatedItems[idx].IsUp || updatedItems[idx].IsGhost || updatedItems[idx].HasMetadata {
				continue
			}

			// Prioritize visible items
			if idx < start || idx >= end {
				continue
			}

			g.Go(func() error {
				info, err := fs.Stat(ctx, updatedItems[idx].Path)
				if err == nil {
					updatedItems[idx] = core.NewItem(info, updatedItems[idx].Path, updatedItems[idx].GitStatus)
					if updatedItems[idx].IsDir {
						enrichDirectoryDate(ctx, fs, &updatedItems[idx])
					}
					updatedItems[idx].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
				} else {
					updatedItems[idx].HasMetadata = true
				}
				return nil
			})
		}

		_ = g.Wait()

		// 3. Fetch the rest in the background
		g, ctx = errgroup.WithContext(ctx)
		g.SetLimit(10)

		for i := range updatedItems {
			idx := i
			if updatedItems[idx].IsUp || updatedItems[idx].IsGhost || updatedItems[idx].HasMetadata {
				continue
			}

			g.Go(func() error {
				info, err := fs.Stat(ctx, updatedItems[idx].Path)
				if err == nil {
					updatedItems[idx] = core.NewItem(info, updatedItems[idx].Path, updatedItems[idx].GitStatus)
					if updatedItems[idx].IsDir {
						enrichDirectoryDate(ctx, fs, &updatedItems[idx])
					}
					updatedItems[idx].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
				} else {
					updatedItems[idx].HasMetadata = true
				}
				return nil
			})
		}

		_ = g.Wait()

		ro, _ := fs.IsReadOnly(ctx, path)

		return LoadedItemsMsg{
			Generation: gen,
			Path:       path,
			Items:      updatedItems,
			IsReadOnly: ro,
		}
	}
}

func enrichDirectoryDate(ctx context.Context, fs core.FileSystem, item *core.Item) {
	entries, err := fs.ReadDirEntries(ctx, item.Path)
	if err != nil {
		return
	}

	maxMTime := item.MTime
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			if info.ModTime().After(maxMTime) {
				maxMTime = info.ModTime()
			}
		}
	}
	item.MTime = maxMTime
}

func handleNavKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	// Don't handle nav keys if an input is active or a modal view is open
	if m.UI.InputActive || m.UI.SettingsOpen || m.UI.LogOpen || m.UI.ClipboardOpen {
		return nil
	}

	key := msg.String()

	// Tab management shortcuts
	if strings.HasPrefix(key, "alt+") {
		if len(key) == 5 && key[4] >= '1' && key[4] <= '9' {
			tabNum := int(key[4] - '0')
			return switchTab(m, tabNum)
		}
		switch key {
		case "alt+t":
			return createTab(m)
		case "alt+w":
			return closeTab(m)
		}
	}

	switch key {
	case "up", "k":
		moveCursor(m, -1)
	case "down", "j":
		moveCursor(m, 1)
	case "enter", "right", "l":
		return navigateToSelected(m)
	case "backspace", "left", "h":
		return navigateToParent(m)
	case " ":
		toggleSelection(m)
		return nil
	case "/":
		m.StartInput(tui_context.InputSearch)
		return m.Inputs.ActiveInput.FocusCmd()
	case "s":
		m.Display.SortMode = m.Display.SortMode.Next()
		sorting.SortItems(m.Navigation.Items, m.Display.SortMode, true)
		ApplyFilter(m)
		return nil
	case "g":
		m.StartInput(tui_context.InputGoto)
		m.Inputs.ActiveInput.SetValue(m.Navigation.Path)
		return m.Inputs.ActiveInput.FocusCmd()
	case "alt+/":
		m.StartInput(tui_context.InputFuzzySearch)
		return m.Inputs.ActiveInput.FocusCmd()
	case "[":
		return NavigateBack(m)
	case "]":
		return NavigateForward(m)
	}
	return nil
}

func toggleSelection(m *tui_context.Model) {
	if len(m.Navigation.FilteredItems) == 0 {
		return
	}
	cursor := m.Navigation.Cursor
	if cursor >= len(m.Navigation.FilteredItems) {
		return
	}
	item := &m.Navigation.FilteredItems[cursor]
	if item.IsUp {
		return
	}

	item.Selected = !item.Selected
	m.Navigation.ToggleSelection(item.Path)

	m.UI.SelectMode = m.Navigation.SelectedCount > 0
}

// Tab management functions

func saveTabState(m *tui_context.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		t := &m.Tabs[m.ActiveTab]
		t.FS = m.FS
		t.Path = m.Navigation.Path
		t.Items = m.Navigation.Items
		t.FilteredItems = m.Navigation.FilteredItems
		t.Cursor = m.Navigation.Cursor
		t.Offset = m.Navigation.Offset
		t.SortMode = m.Display.SortMode
		t.GitBranch = m.Git.Branch
		t.GitRoot = m.Git.Root
		t.Searching = m.UI.InputActive && m.Inputs.Mode == tui_context.InputSearch
		t.SearchQuery = m.Inputs.ActiveInput.Value()
		t.SelectMode = m.UI.SelectMode
		t.RemoteUser = m.Remote.User
		t.RemoteHost = m.Remote.Host
		t.BackHistory = make([]string, len(m.Navigation.BackHistory))
		copy(t.BackHistory, m.Navigation.BackHistory)
		t.ForwardHistory = make([]string, len(m.Navigation.ForwardHistory))
		copy(t.ForwardHistory, m.Navigation.ForwardHistory)
		t.SelectedPaths = make(map[string]bool)
		for k, v := range m.Navigation.SelectedPaths {
			t.SelectedPaths[k] = v
		}
	}
}

func syncTabToModel(m *tui_context.Model) tea.Cmd {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		t := m.Tabs[m.ActiveTab]
		m.FS = t.FS
		m.Navigation.Path = t.Path
		m.Navigation.Items = t.Items
		m.Navigation.FilteredItems = t.FilteredItems
		m.Navigation.Cursor = t.Cursor
		m.Navigation.Offset = t.Offset
		m.Display.SortMode = t.SortMode
		m.Git.Branch = t.GitBranch
		m.Git.Root = t.GitRoot
		m.Remote.User = t.RemoteUser
		m.Remote.Host = t.RemoteHost
		m.Navigation.BackHistory = make([]string, len(t.BackHistory))
		copy(m.Navigation.BackHistory, t.BackHistory)
		m.Navigation.ForwardHistory = make([]string, len(t.ForwardHistory))
		copy(m.Navigation.ForwardHistory, t.ForwardHistory)

		var cmd tea.Cmd
		if t.Searching {
			m.UI.InputActive = true
			m.Inputs.Mode = tui_context.InputSearch
			cmd = m.Inputs.ActiveInput.FocusCmd()
		} else {
			m.UI.InputActive = false
			m.Inputs.Mode = tui_context.InputNone
			m.Inputs.ActiveInput.Blur()
		}

		m.Inputs.ActiveInput.SetValue(t.SearchQuery)
		m.UI.SelectMode = t.SelectMode
		m.Navigation.SelectedPaths = make(map[string]bool)
		for k, v := range t.SelectedPaths {
			m.Navigation.SelectedPaths[k] = v
		}
		m.Navigation.PathGen++
		return cmd
	}
	return nil
}

func createTab(m *tui_context.Model) tea.Cmd {
	if len(m.Tabs) >= 9 {
		return SetMsg(m, "Tab limit reached (max 9 tabs)")
	}
	saveTabState(m)
	m.AddTab(m.Navigation.Path)
	m.ActiveTab = len(m.Tabs) - 1
	cmd := syncTabToModel(m)
	return tea.Batch(cmd, Reload(m, false))
}

func switchTab(m *tui_context.Model, tabNum int) tea.Cmd {
	if tabNum > 0 && tabNum <= len(m.Tabs) {
		saveTabState(m)
		m.ActiveTab = tabNum - 1
		cmd := syncTabToModel(m)
		return tea.Batch(cmd, Reload(m, false))
	}
	return nil
}

func closeTab(m *tui_context.Model) tea.Cmd {
	if m.CloseActiveTab() {
		cmd := syncTabToModel(m)
		return tea.Batch(cmd, Reload(m, false))
	}
	return nil
}

// ApplyFilter filters the navigation items based on current search query
func ApplyFilter(m *tui_context.Model) {
	if m.UI.InputActive && m.Inputs.Mode == tui_context.InputSearch {
		m.Navigation.FilterQuery = strings.ToLower(m.Inputs.ActiveInput.Value())
	}
	file.ApplyFilter(m)
	syncOffset(m)
}

func moveCursor(m *tui_context.Model, delta int) {
	items := m.Navigation.FilteredItems
	if len(items) == 0 {
		return
	}

	newCursor := m.Navigation.Cursor + delta
	if newCursor < 0 {
		if m.Config.WrapNavigation {
			newCursor = len(items) - 1
		} else {
			newCursor = 0
		}
	} else if newCursor >= len(items) {
		if m.Config.WrapNavigation {
			newCursor = 0
		} else {
			newCursor = len(items) - 1
		}
	}

	m.Navigation.Cursor = newCursor
	syncOffset(m)

	// Save to memory immediately to prevent jumps during background reloads
	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)
}

func moveCursorToStart(m *tui_context.Model) {
	m.Navigation.Cursor = 0
	syncOffset(m)
	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)
}

func moveCursorToEnd(m *tui_context.Model) {
	if len(m.Navigation.FilteredItems) > 0 {
		m.Navigation.Cursor = len(m.Navigation.FilteredItems) - 1
	}
	syncOffset(m)
	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)
}

func syncOffset(m *tui_context.Model) {
	if m.Display.ViewportHeight == 0 {
		return
	}

	cursor := m.Navigation.Cursor
	offset := m.Navigation.Offset
	height := m.Display.ViewportHeight

	if cursor < offset {
		m.Navigation.Offset = cursor
	} else if cursor >= offset+height {
		m.Navigation.Offset = cursor - height + 1
	}
}

func navigateToSelected(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.FilteredItems) == 0 {
		return nil
	}
	selected := m.Navigation.FilteredItems[m.Navigation.Cursor]

	if selected.IsUp {
		return navigateToParent(m)
	}

	if selected.IsDir {
		if err := ops.ValidatePath(m.FS, m.Navigation.Path, selected.Name); err != nil {
			return SetErrMsg(m, "Security: "+err.Error())
		}
		if !selected.CanRead {
			return SetErrMsg(m, "Access Denied: Permission denied")
		}
		return NavigateToPath(m, m.FS.Join(m.Navigation.Path, selected.Name))
	}

	// File opening logic
	if !m.FS.IsLocal() {
		return SetErrMsg(m, "Opening remote files not supported yet")
	}

	execCmd, isTerminal, err := ops.GetOpenCmd(m.FS, selected.Path, m.Config.EditorIndex)
	if err != nil {
		return LogError(m, fileerrors.WrapError(err, "Open"), "Open")
	}

	if isTerminal {
		return tea.ExecProcess(execCmd, func(err error) tea.Msg {
			if err != nil {
				return ErrorMsg{Err: fileerrors.WrapError(err, "Open")}
			}
			return nil
		})
	} else {
		if err := execCmd.Start(); err != nil {
			return LogError(m, fileerrors.WrapError(err, "Open"), "Open")
		}
		return nil
	}
}

func navigateToParent(m *tui_context.Model) tea.Cmd {
	parent := m.FS.Dir(m.Navigation.Path)
	return NavigateToPath(m, parent)
}

// NavigateBack moves to the previous path in history
func NavigateBack(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.BackHistory) == 0 {
		return nil
	}

	// Save current to ForwardHistory
	m.Navigation.ForwardHistory = append(m.Navigation.ForwardHistory, m.Navigation.Path)

	// Pop from BackHistory
	prevPath := m.Navigation.BackHistory[len(m.Navigation.BackHistory)-1]
	m.Navigation.BackHistory = m.Navigation.BackHistory[:len(m.Navigation.BackHistory)-1]

	return navigateToPathInternal(m, prevPath, true)
}

// NavigateForward moves forward in history
func NavigateForward(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.ForwardHistory) == 0 {
		return nil
	}

	// Save current to BackHistory
	m.Navigation.BackHistory = append(m.Navigation.BackHistory, m.Navigation.Path)

	// Pop from ForwardHistory
	nextPath := m.Navigation.ForwardHistory[len(m.Navigation.ForwardHistory)-1]
	m.Navigation.ForwardHistory = m.Navigation.ForwardHistory[:len(m.Navigation.ForwardHistory)-1]

	return navigateToPathInternal(m, nextPath, true)
}

// navigateToPathInternal is a helper for history navigation that skips history pushing
func navigateToPathInternal(m *tui_context.Model, path string, isHistory bool) tea.Cmd {
	// Clean and validate path
	info, err := m.FS.Stat(m.Context, path)
	if err != nil {
		return LogError(m, fileerrors.WrapError(err, "Stat"), "Navigate")
	}

	if !info.IsDir() {
		return LogError(m, fmt.Errorf("not a directory"), "Navigate")
	}

	// Save current state to cache
	m.Cache.CursorMemory.Put(m.Navigation.Path, m.Navigation.Cursor)
	m.Cache.OffsetMemory.Put(m.Navigation.Path, m.Navigation.Offset)

	// Memory Cleanup
	oldParent := m.FS.Dir(m.Navigation.Path)
	m.Cache.ItemCache.Unprotect(oldParent)

	// Update path
	m.Navigation.Path = path
	m.Navigation.PathGen++

	// Pin new parent
	newParent := m.FS.Dir(path)
	m.Cache.ItemCache.Protect(newParent)

	// Clear selection on navigation
	m.ClearSelection()

	return Reload(m, false)
}

// SwitchToLocal switches the current filesystem back to local
func SwitchToLocal(m *tui_context.Model, path string) tea.Cmd {
	if m.FS.IsLocal() {
		return NavigateToPath(m, path)
	}

	// Only close the current FS if it's not being used by any other tab
	isShared := false
	for i, t := range m.Tabs {
		if i == m.ActiveTab {
			continue
		}
		if t.FS == m.FS {
			isShared = true
			break
		}
	}

	if !isShared {
		m.FS.Close()
	}

	m.FS = local.NewLocalFS()

	targetPath := path
	if targetPath == "" {
		home, err := m.FS.GetHomeDir()
		if err == nil {
			targetPath = home
		} else {
			targetPath = "/"
		}
	}

	// Memory Cleanup
	oldParent := m.FS.Dir(m.Navigation.Path)
	m.Cache.ItemCache.Unprotect(oldParent)

	m.Navigation.Path = targetPath
	m.Navigation.PathGen++

	// Pin parent
	newParent := m.FS.Dir(targetPath)
	m.Cache.ItemCache.Protect(newParent)

	m.Navigation.Offset = 0
	m.Navigation.FilterQuery = ""
	m.Inputs.ActiveInput.Reset()

	return Reload(m, false)
}

// NavigateToPath handles navigation to a specific directory path
func NavigateToPath(m *tui_context.Model, path string) tea.Cmd {
	// Update history: Push current path to BackHistory and clear ForwardHistory
	if m.Navigation.Path != "" {
		if len(m.Navigation.BackHistory) == 0 || m.Navigation.BackHistory[len(m.Navigation.BackHistory)-1] != m.Navigation.Path {
			m.Navigation.BackHistory = append(m.Navigation.BackHistory, m.Navigation.Path)
			if len(m.Navigation.BackHistory) > 100 {
				m.Navigation.BackHistory = m.Navigation.BackHistory[1:]
			}
		}
	}
	m.Navigation.ForwardHistory = nil

	return navigateToPathInternal(m, path, false)
}
// Reload triggers an asynchronous reload of the current directory.
// If silent is true, it won't show the loading spinner.
func Reload(m *tui_context.Model, silent bool) tea.Cmd {
	path := m.Navigation.Path
	gen := m.Navigation.PathGen
	fs := m.FS
	gs := m.GS
	ctx := m.Context
	mode := m.Display.SortMode
	showHidden := m.Config.ShowHidden
	sizeFormatIdx := m.Config.SizeFormatIndex
	dateFormatIdx := m.Config.DateFormatIndex

	// 1. Synchronous Cache Check (The "Instant" Fix)
	// If we have it in cache, populate immediately to avoid flicker
	if items, ok := m.Cache.ItemCache.Get(path); ok {
		m.Navigation.Items = items
		ApplyFilter(m)

		// Restore cursor/offset from memory if available
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
		syncOffset(m)

		m.UI.Loading = false
		silent = true // Force background refresh to be silent
	} else {
		// If NOT in cache, we need to reset the cursor for the upcoming new list
		m.Navigation.Cursor = 0
		m.Navigation.Offset = 0
	}

	if !silent {
		m.UI.Loading = true
	}

	// 2. Immediate Skeleton (The "Stable UI" Fix)
	// If the list is currently empty (not a cache hit), at least add the ".." entry
	// so the UI doesn't completely disappear.
	if len(m.Navigation.Items) == 0 && path != "/" && path != fs.Separator() {
		m.Navigation.Items = []core.Item{{
			Name:      "↑ ..",
			IsDir:     true,
			IsUp:      true,
			SearchKey: "..",
			Path:      fs.Dir(path),
		}}
		ApplyFilter(m)
	}

	// Cancel previous git operation if running
	if m.Git.CancelFunc != nil {
		m.Git.CancelFunc()
	}

	var cmds []tea.Cmd

	// 2. Load Skeleton (names only) quickly
	loadSkeletonCmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, constants.DirectoryLoadTimeout)
		defer cancel()

		var gitStatuses map[string]string
		var gitRoot string

		if gs.IsEnabled() {
			if root, ok := m.Cache.GitRootCache.Get(path); ok {
				gitRoot = root
			} else {
				gitRoot = gs.GetRoot(ctx, path)
				if gitRoot != "" {
					m.Cache.GitRootCache.Put(path, gitRoot)
				}
			}
			gitStatuses, _ = gs.GetStatus(ctx, path)
		}

		items, err := listing.LoadSkeleton(ctx, fs, path, showHidden, gitStatuses)
		if err != nil {
			return LoadedItemsMsg{Generation: gen, Path: path, Err: err}
		}

		// Pre-calculate formatted strings for items that already have metadata
		for i := range items {
			items[i].UpdateFormatting(sizeFormatIdx, dateFormatIdx)
		}

		// Immediate Sort of skeleton
		sorting.SortItems(items, mode, true)

		return PartialItemsMsg{
			Generation: gen,
			Path:       path,
			Items:      items,
		}
	}

	// 3. Setup cancellation for background tasks
	_, gitCancel := context.WithCancel(ctx)
	m.Git.CancelFunc = gitCancel

	cmds = append(cmds, loadSkeletonCmd)
	return tea.Batch(cmds...)
}

// WatchDirAction starts watching the current directory
func WatchDirAction(m *tui_context.Model) tea.Cmd {
	if m.Navigation.Path == "" {
		return nil
	}

	// If we switched between local and remote, reset listening state
	if m.FS.IsLocal() == m.Watcher.IsRemote {
		m.Watcher.IsListening = false
		m.Watcher.IsRemote = !m.FS.IsLocal()
	}

	if !m.FS.IsLocal() {
		if m.Watcher.IsListening {
			return nil
		}
		m.Watcher.IsListening = true
		return WatchRemoteDir()
	}

	if m.Watcher.Watcher == nil {
		return nil
	}

	// Update the watched directory if it changed
	if m.Watcher.LastWatched != m.Navigation.Path {
		if m.Watcher.LastWatched != "" {
			_ = m.Watcher.Watcher.Remove(m.Watcher.LastWatched)
		}
		if err := m.Watcher.Watcher.Add(m.Navigation.Path); err != nil {
			return nil
		}
		m.Watcher.LastWatched = m.Navigation.Path
	}

	if m.Watcher.IsListening {
		return nil
	}

	m.Watcher.IsListening = true
	return WatchDir(m.Watcher.Watcher)
}

func finalizeDirectoryLoad(m *tui_context.Model, msg LoadedItemsMsg) tea.Cmd {
	m.UI.Loading = false
	if msg.Generation != m.Navigation.PathGen {
		return nil
	}

	if msg.Err != nil {
		cmd := SetErrMsg(m, "Failed to load directory: "+msg.Err.Error())
		m.Navigation.Items = []core.Item{}
		m.Navigation.FilteredItems = []core.Item{}

		// If we failed to load the current path, try going back
		if m.Navigation.Path == msg.Path {
			parent := m.FS.Dir(m.Navigation.Path)
			if parent != m.Navigation.Path {
				m.Navigation.Path = parent
				m.Navigation.PathGen++
				return tea.Batch(cmd, Reload(m, false))
			}
		}
		return cmd
	}

	m.Navigation.Items = msg.Items
	ApplyFilter(m)
	m.Git.Branch = msg.GitBranch
	m.Git.Root = msg.GitRoot
	m.Display.ReadOnly = msg.IsReadOnly

	// Store in cache for next time
	if !msg.Cached && msg.Err == nil {
		m.Cache.ItemCache.Put(msg.Path, msg.Items)
	}

	// Restore cursor/offset from memory if available
	if val, ok := m.Cache.CursorMemory.Get(msg.Path); ok {
		m.Navigation.Cursor = val
	}
	if val, ok := m.Cache.OffsetMemory.Get(msg.Path); ok {
		m.Navigation.Offset = val
	}

	// Bounds check cursor and offset
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

	syncOffset(m)
	return WatchDirAction(m)
}
