package tui

import (
	"strings"
	"time"

	"fm/internal/config"
	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func tickDirSizeBatch() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return dirSizeBatchTickMsg{}
	})
}

type dirSizeBatchTickMsg struct{}

// Update implements the tea.Model interface.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case LoadedItemsMsg:
		m.loading = false
		if msg.Generation != m.pathGeneration {
			return m, nil
		}
		if msg.Err != nil {
			cmd := m.LogError(msg.Err, "failed to load directory")
			m.items = []files.Item{}

			// If we failed to load the current path, try going back
			if m.path == msg.Path {
				parent := m.fs.Dir(m.path)
				if parent != m.path {
					m.path = parent
					m.pathGeneration++
					return m, tea.Batch(cmd, m.reload())
				}
			}
			return m, cmd
		} else {
			m.err = nil
			m.items = msg.Items
			m.gitBranch = msg.GitBranch
			m.gitRoot = msg.GitRoot
			m.readOnly = msg.IsReadOnly

			// Restore selection
			for i := range m.items {
				if m.selectedPaths[m.items[i].Path] {
					m.items[i].Selected = true
				}
			}
		}
		m.selectMode = len(m.selectedPaths) > 0

		// Restore cursor/offset from tab or memory
		if m.activeTab < len(m.tabs) {
			tab := m.tabs[m.activeTab]
			if tab.path == m.path {
				m.cursor = tab.cursor
				m.offset = tab.offset
			}
		}

		if m.cursor == 0 {
			if val, ok := m.cursorMemory.Get(m.path); ok {
				m.cursor = val
			}
		}
		if m.offset == 0 {
			if val, ok := m.offsetMemory.Get(m.path); ok {
				m.offset = val
			}
		}

		var sizeCmds []tea.Cmd
		for i, item := range m.items {
			if item.IsDir && !item.IsUp {
				if entry, ok := m.dirSizeCache.Get(item.Path); ok {
					// Validate MTime
					if entry.MTime.Equal(item.MTime) {
						m.items[i].Size = entry.Size
					} else {
						sizeCmds = append(sizeCmds, calculateDirSize(m.fs, item.Path))
					}
				} else {
					sizeCmds = append(sizeCmds, calculateDirSize(m.fs, item.Path))
				}
			}
		}

		m.applyFilter()

		if m.cursor >= len(m.filteredItems) {
			m.cursor = 0
		}
		if m.offset >= len(m.filteredItems) {
			m.offset = 0
		}

		// Save the valid loaded state to active tab
		m.saveTabState()

		if m.fs.IsLocal() {
			if m.lastWatched != "" {
				m.watcher.Remove(m.lastWatched)
			}
			m.watcher.Add(m.path)
			m.lastWatched = m.path

			// Also watch .git directory if present
			if m.lastWatchedGit != "" {
				m.watcher.Remove(m.lastWatchedGit)
				m.lastWatchedGit = ""
			}
			if m.gitRoot != "" {
				gitPath := m.fs.Join(m.gitRoot, ".git")
				// We watch .git folder to detect index changes, etc.
				if err := m.watcher.Add(gitPath); err == nil {
					m.lastWatchedGit = gitPath
				}
			}
		}

		if m.cfg.EnableGit {
			// Check cache first for instant response
			if cachedStatus, ok := m.gitStatusCache[m.path]; ok {
				for i, item := range m.items {
					if status, exists := cachedStatus[item.Name]; exists {
						m.items[i].GitStatus = status
					}
				}
			}
			// Fetch in background to update cache
			cmds = append(cmds, fetchGitStatus(m.fs, m.path))
		}

		// If we have many sizes to calculate, start a batch tick
		if len(sizeCmds) > 0 {
			m.sizeTickActive = true
			cmds = append(cmds, tickDirSizeBatch())
		}

		return m, tea.Batch(append(cmds, append(sizeCmds, m.watchDir())...)...)

	case DirSizeMsg:
		m.pendingDirSizes[msg.Path] = msg.Size
		// We can't easily pass MTime through pendingDirSizes map without changing its type
		// but we can put it directly in cache here or update the map type.
		// Let's update the map type in model.go later. For now, just put in cache.
		m.dirSizeCache.Put(msg.Path, SizeCacheEntry{Size: msg.Size, MTime: msg.MTime})

		if !m.sizeTickActive {
			m.sizeTickActive = true
			return m, tickDirSizeBatch()
		}
		return m, nil

	case dirSizeBatchTickMsg:
		if len(m.pendingDirSizes) == 0 {
			// Keep ticking if we still have items being counted in current view
			stillCounting := false
			for _, item := range m.items {
				if item.IsDir && !item.IsUp && item.Size == -1 {
					stillCounting = true
					break
				}
			}
			if stillCounting {
				return m, tickDirSizeBatch()
			}
			m.sizeTickActive = false
			return m, nil
		}

		batch := m.pendingDirSizes
		m.pendingDirSizes = make(map[string]int64)

		needsResort := false
		for path, size := range batch {
			// Note: We already Put it in cache in DirSizeMsg handler above
			for i, item := range m.items {
				if item.Path == path {
					m.items[i].Size = size
					break
				}
			}
			// Update filtered items too
			for i, item := range m.filteredItems {
				if item.Path == path {
					m.filteredItems[i].Size = size
					break
				}
			}
		}

		if m.sortMode == files.SortSizeDesc || m.sortMode == files.SortSizeAsc {
			needsResort = true
		}

		if needsResort {
			m.applyFilter()
		}

		// Keep ticking if we still have items being counted (size == -1)
		stillCounting := false
		for _, item := range m.items {
			if item.IsDir && !item.IsUp && item.Size == -1 {
				stillCounting = true
				break
			}
		}

		if stillCounting || len(m.pendingDirSizes) > 0 {
			return m, tickDirSizeBatch()
		}
		m.sizeTickActive = false
		return m, nil

	case GitStatusMsg:
		if msg.Path != m.path {
			return m, nil
		}
		m.gitBranch = msg.Branch

		statusMap := msg.Statuses
		// Cache git status for this directory
		m.gitStatusCache[msg.Path] = statusMap

		for i, item := range m.items {
			if status, ok := statusMap[item.Name]; ok {
				m.items[i].GitStatus = status
			}
		}

		seen := make(map[string]bool)
		for _, item := range m.items {
			seen[item.Name] = true
		}
		for name, status := range statusMap {
			if !seen[name] && status == "D" {
				m.items = append(m.items, files.Item{
					Name:      name,
					Path:      m.fs.Join(m.path, name),
					IsDir:     false,
					GitStatus: "D",
					IsGhost:   true,
					Size:      0,
				})
			}
		}
		m.applyFilter()
		return m, nil

	case WatchEventMsg:
		path := msg.Event.Name
		isGit := strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") ||
			strings.Contains(path, "\\.git\\") || strings.HasSuffix(path, "\\.git")

		if msg.Err != nil {
			return m, m.watchDir()
		}

		if isGit {
			// Invalidate git status cache for current path if git repo changed
			delete(m.gitStatusCache, m.path)
			// Fetch status again but don't reload full directory list to avoid loops
			return m, tea.Batch(fetchGitStatus(m.fs, m.path), m.watchDir())
		}

		// Invalidate git status cache on regular file changes for accuracy
		delete(m.gitStatusCache, m.path)
		return m, tea.Batch(m.reload(), m.watchDir())

	case WatcherErrorMsg:
		// Attempt to restart watcher
		return m, tea.Batch(m.LogError(msg.Err, "File watcher error"), m.setMsg("Watcher error: restarting..."), m.restartWatcher())

	case WatcherClosedMsg:
		// Attempt to restart watcher
		return m, tea.Batch(m.setMsg("Watcher closed: restarting..."), m.restartWatcher())

	case ProgressMsg:
		m.showProgress = true
		m.progressLabel = msg.Label
		m.progressPercent = msg.Percent
		return m, listenToProgress(msg.Channel)

	case OperationFinishedMsg:
		m.showProgress = false
		m.progressLabel = ""
		for _, p := range msg.Paths {
			delete(m.processingItems, p)
			delete(m.selectedPaths, p)
		}
		m.selectMode = len(m.selectedPaths) > 0
		return m, nil

	case conflictMsg:
		m.loading = false
		m.confirming = true
		m.actionType = "conflict"
		m.conflictSrc = msg.Src
		m.conflictDst = msg.Dst
		m.pendingItems = msg.PendingItems
		m.clipboardCut = msg.IsMove
		return m, nil

	case clearMsg:
		m.ClearMsg()
		return m, nil

	case errMsg:
		// Clear all processing items on error for safety
		m.processingItems = make(map[string]bool)
		return m, m.LogError(msg.err, "Operation failed")

	case tea.KeyMsg:
		if !m.searching && !m.renaming && !m.confirming && !m.settingsOpen && msg.String() == "esc" {
			m.selectMode = false
			m.selectedPaths = make(map[string]bool)
			hasSelection := false
			for i := range m.items {
				if m.items[i].Selected {
					m.items[i].Selected = false
					hasSelection = true
				}
			}
			if hasSelection {
				m.applyFilter()
				return m, nil
			}
			m.msg = ""
			return m, nil
		}
	}

	if m.settingsOpen {
		return m.handleSettingsUpdate(msg)
	}

	if m.searching {
		return m.handleSearching(msg)
	}

	if m.renaming {
		return m.handleRenaming(msg)
	}

	if m.confirming {
		return m.handleConfirming(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle tab operations first
		switch msg.String() {
		case "alt+t":
			// Create new tab (max 9 tabs)
			if len(m.tabs) >= 9 {
				return m, m.setMsg("Tab limit reached (max 9 tabs)")
			}
			// Save current tab state BEFORE switching
			m.saveTabState()
			newTab := Tab{
				path:          m.path,
				sortMode:      files.SortDefault,
				selectedPaths: make(map[string]bool),
			}
			m.tabs = append(m.tabs, newTab)
			m.activeTab = len(m.tabs) - 1
			m.syncTabToModel()
			return m, m.reload()
		case "alt+1", "alt+2", "alt+3", "alt+4", "alt+5", "alt+6", "alt+7", "alt+8", "alt+9":
			// Only switch tabs if there are multiple tabs
			if len(m.tabs) > 1 {
				tabNum := int(msg.String()[4] - '0') // Extract number from "alt+N"
				if tabNum > 0 && tabNum <= len(m.tabs) {
					// Save current tab state BEFORE switching
					m.saveTabState()
					// Reset model cursor/offset so syncTabToModel can set new ones clearly
					m.cursor = 0
					m.offset = 0
					// Switch to new tab
					m.activeTab = tabNum - 1
					m.syncTabToModel()
					return m, m.reload()
				}
			}
			return m, nil
		case "alt+w":
			// Close current tab (only if more than one tab)
			if len(m.tabs) > 1 {
				// Remove current tab
				m.tabs = append(m.tabs[:m.activeTab], m.tabs[m.activeTab+1:]...)
				// Adjust active tab index
				if m.activeTab >= len(m.tabs) {
					m.activeTab = len(m.tabs) - 1
				}
				m.syncTabToModel()
				return m, m.reload()
			}
			return m, nil
		case "ctrl+c", "q":
			if m.fs.IsLocal() {
				m.watcher.Close()
			}
			m.dirSizeCache.Save(config.GetSizeCachePath())
			return m, tea.Quit
		}

		// Try navigation handlers
		navCmds := m.handleNavigation(msg)
		if len(navCmds) > 0 {
			cmds = append(cmds, navCmds...)
		}

		// Try action handlers
		actionCmds := m.handleAction(msg)
		if len(actionCmds) > 0 {
			cmds = append(cmds, actionCmds...)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update text input widths to be responsive
		maxInputWidth := m.width - 10
		if maxInputWidth < 20 {
			maxInputWidth = 20
		}
		if maxInputWidth > 80 {
			maxInputWidth = 80
		}
		m.searchInput.Width = maxInputWidth
		m.renameInput.Width = maxInputWidth

		// Update settings scroll if in settings mode to ensure proper display
		if m.settingsOpen {
			m.updateSettingsScroll()
		}

		// Force a re-render by returning the model
		return m, nil
	}

	if m.msg != "" {
		cmds = append(cmds, clearMessage())
	}

	return m, tea.Batch(cmds...)
}
