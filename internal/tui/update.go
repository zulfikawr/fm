package tui

import (
	"strings"
	"time"

	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming events and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case LoadedItemsMsg:
		m.loading = false
		if msg.Path != m.path {
			return m, nil
		}
		if msg.Err != nil {
			m.LogError(msg.Err, "Loading directory")
			m.items = []files.Item{}
		} else {
			m.err = nil
			m.items = msg.Items
			m.gitBranch = msg.GitBranch
		}
		m.selectMode = false

		if m.cursor == 0 {
			if val, ok := m.cursorMemory[m.path]; ok {
				m.cursor = val
			}
		}
		if m.offset == 0 {
			if val, ok := m.offsetMemory[m.path]; ok {
				m.offset = val
			}
		}

		for i, item := range m.items {
			if item.IsDir && !item.IsUp {
				if size, ok := m.dirSizeCache[item.Path]; ok {
					m.items[i].Size = size
				} else {
					cmds = append(cmds, calculateDirSize(m.fs, item.Path))
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

		if m.fs.IsLocal() {
			if m.lastWatched != "" {
				m.watcher.Remove(m.lastWatched)
			}
			m.watcher.Add(m.path)
			m.lastWatched = m.path
		}

		if m.cfg.EnableGit {
			cmds = append(cmds, fetchGitStatus(m.fs, m.path))
		}

		return m, tea.Batch(append(cmds, m.watchDir())...)

	case DirSizeMsg:
		m.dirSizeCache[msg.Path] = msg.Size
		for i, item := range m.items {
			if item.Path == msg.Path {
				m.items[i].Size = msg.Size
				break
			}
		}
		m.applyFilter()
		return m, nil

	case GitStatusMsg:
		if msg.Path != m.path {
			return m, nil
		}
		m.gitBranch = msg.Branch

		statusMap := msg.Statuses
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
		if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") ||
			strings.Contains(path, "\\.git\\") || strings.HasSuffix(path, "\\.git") {
			return m, m.watchDir()
		}
		if msg.Err != nil {
			return m, m.watchDir()
		}
		return m, tea.Batch(m.reload(), m.watchDir())

	case clearMsg:
		m.ClearMsg()
		return m, nil

	case errMsg:
		m.LogError(msg.err, "Operation failed")
		return m, nil

	case tea.KeyMsg:
		if !m.searching && !m.renaming && !m.confirming && !m.settingsOpen && msg.String() == "esc" {
			m.selectMode = false
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
		switch msg.String() {
		case "ctrl+c", "q":
			if m.fs.IsLocal() {
				m.watcher.Close()
			}
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
	}

	if m.msg != "" {
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearMsg{}
		}))
	}

	return m, tea.Batch(cmds...)
}
