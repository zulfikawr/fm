package tui

import (
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
			m.err = msg.Err
			m.items = []files.Item{}
		} else {
			m.err = nil
			m.items = msg.Items
			m.gitBranch = msg.GitBranch
		}
		m.selectMode = false
		m.applyFilter()

		if val, ok := m.cursorMemory[m.path]; ok {
			m.cursor = val
			if m.cursor >= len(m.filteredItems) {
				m.cursor = 0
			}
		} else {
			m.cursor = 0
		}
		if val, ok := m.offsetMemory[m.path]; ok {
			m.offset = val
			if m.offset >= len(m.filteredItems) {
				m.offset = 0
			}
		} else {
			m.offset = 0
		}

		if m.lastWatched != "" {
			m.watcher.Remove(m.lastWatched)
		}
		m.watcher.Add(m.path)
		m.lastWatched = m.path
		return m, m.watchDir()

	case WatchEventMsg:
		if msg.Err != nil {
			return m, m.watchDir()
		}
		return m, tea.Batch(m.reload(), m.watchDir())

	case clearMsg:
		if time.Since(m.msgTime) >= 5*time.Second {
			m.msg = ""
		}
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
			m.watcher.Close()
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
