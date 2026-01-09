package tui

import (
	"fmt"
	"path/filepath"
	"time"

	"filemanager/internal/files"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

type clearMsg struct{}

// WatchEventMsg is sent when a file system event occurs in the watched directory.
type WatchEventMsg struct {
	Event fsnotify.Event
	Err   error
}

func (m *Model) setMsg(msg string) {
	m.msg = msg
	m.msgTime = time.Now()
}

func (m *Model) watchDir() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
			return WatchEventMsg{Event: event}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return nil
			}
			return WatchEventMsg{Err: err}
		}
	}
}

// Update handles incoming events and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
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
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "enter":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
		}

		m.searchInput, cmd = m.searchInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		m.applyFilter()
		if m.cursor >= len(m.filteredItems) {
			m.cursor = 0
			m.offset = 0
		}
		return m, tea.Batch(cmds...)
	}

	if m.renaming {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.renaming = false
				m.renameInput.Blur()
				return m, nil
			case "enter":
				newName := m.renameInput.Value()
				if newName != "" {
					selected := m.filteredItems[m.cursor]
					oldPath := selected.Path
					newPath := filepath.Join(m.path, newName)
					if err := files.Rename(oldPath, newPath); err != nil {
						m.setMsg("Rename failed: " + err.Error())
					} else {
						cmds = append(cmds, m.reload())
					}
				}
				m.renaming = false
				m.renameInput.Blur()
				return m, tea.Batch(cmds...)
			}
		}
		m.renameInput, cmd = m.renameInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	if m.confirming {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if m.actionType == "delete" {
					cmds = append(cmds, m.performDelete()...)
				} else if m.actionType == "paste" {
					cmds = append(cmds, m.performPaste()...)
				}
				m.confirming = false
				return m, tea.Batch(cmds...)
			case "n", "N", "esc":
				m.confirming = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.watcher.Close()
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else if m.cfg.WrapNavigation && len(m.filteredItems) > 0 {
				m.cursor = len(m.filteredItems) - 1
			}

			if m.cursor < m.offset {
				m.offset = m.cursor
			} else if m.cursor >= m.offset+m.getViewportHeight() {
				m.offset = m.cursor - m.getViewportHeight() + 1
			}

		case "down", "j":
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			} else if m.cfg.WrapNavigation && len(m.filteredItems) > 0 {
				m.cursor = 0
			}

			viewportHeight := m.getViewportHeight()
			if m.cursor >= m.offset+viewportHeight {
				m.offset = m.cursor - viewportHeight + 1
			} else if m.cursor < m.offset {
				m.offset = m.cursor
			}

		case "enter", "right", "l":
			if len(m.filteredItems) == 0 {
				break
			}
			selected := m.filteredItems[m.cursor]

			if selected.IsUp {
				m.cursorMemory[m.path] = m.cursor
				m.offsetMemory[m.path] = m.offset
				m.path = filepath.Dir(m.path)
				cmds = append(cmds, m.reload())
			} else if selected.IsDir {
				m.cursorMemory[m.path] = m.cursor
				m.offsetMemory[m.path] = m.offset
				m.path = filepath.Join(m.path, selected.Name)
				cmds = append(cmds, m.reload())
			}

		case "backspace", "left", "h":
			m.cursorMemory[m.path] = m.cursor
			m.offsetMemory[m.path] = m.offset
			m.path = filepath.Dir(m.path)
			cmds = append(cmds, m.reload())

		case " ":
			if len(m.filteredItems) > 0 {
				idx := m.cursor
				if !m.filteredItems[idx].IsUp {
					m.filteredItems[idx].Selected = !m.filteredItems[idx].Selected
					for i := range m.items {
						if m.items[i].Path == m.filteredItems[idx].Path {
							m.items[i].Selected = m.filteredItems[idx].Selected
							break
						}
					}
				}
			}

			// Update selectMode based on whether anything is selected
			anySelected := false
			for _, item := range m.items {
				if item.Selected {
					anySelected = true
					break
				}
			}
			m.selectMode = anySelected

		case "s":
			m.sortMode = (m.sortMode + 1) % 7
			cmds = append(cmds, m.reload())

		case "/":
			m.searching = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			m.applyFilter()
			return m, textinput.Blink

		case "c":
			m.actionType = "copy"
			m.clipboard = []string{}
			for _, item := range m.items {
				if item.Selected {
					m.clipboard = append(m.clipboard, item.Path)
				}
			}
			if len(m.clipboard) == 0 && len(m.filteredItems) > 0 {
				sel := m.filteredItems[m.cursor]
				if !sel.IsUp {
					m.clipboard = append(m.clipboard, sel.Path)
				}
			}
			m.setMsg(fmt.Sprintf("Copied %d items to clipboard", len(m.clipboard)))

		case "v":
			if len(m.clipboard) > 0 {
				if m.cfg.ConfirmOperations {
					m.confirming = true
					m.actionType = "paste"
				} else {
					cmds = append(cmds, m.performPaste()...)
				}
			}

		case "r":
			if len(m.filteredItems) > 0 {
				sel := m.filteredItems[m.cursor]
				if !sel.IsUp {
					m.renaming = true
					m.renameInput.Focus()
					m.renameInput.SetValue(sel.Name)
					return m, textinput.Blink
				}
			}

		case "d":
			selectedCount := 0
			for _, item := range m.items {
				if item.Selected {
					selectedCount++
				}
			}
			if selectedCount > 0 || (len(m.filteredItems) > 0 && !m.filteredItems[m.cursor].IsUp) {
				if m.cfg.ConfirmOperations {
					m.confirming = true
					m.actionType = "delete"
				} else {
					cmds = append(cmds, m.performDelete()...)
				}
			}

		case ".":
			m.settingsOpen = true
			m.settingsCursor = 0
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

func (m *Model) handleSettingsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	const numSettings = 11 // Hidden, CaseSensitive, Confirmations, Wrapping, Git, ShowSize, SizeFormat, ShowDate, DateFormat, ShowHeader, Theme

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", ".", "q":
			m.settingsOpen = false
			m.cfg.Save()
			return m, m.reload()
		case "up", "k":
			if m.settingsCursor > 0 {
				m.settingsCursor--
				// Skip disabled settings
				if (m.settingsCursor == 7 && !m.cfg.ShowSize) || (m.settingsCursor == 9 && !m.cfg.ShowDateModified) {
					if m.settingsCursor > 0 {
						m.settingsCursor--
					}
				}
			}
		case "down", "j":
			if m.settingsCursor < numSettings-1 {
				m.settingsCursor++
				// Skip disabled settings
				if (m.settingsCursor == 7 && !m.cfg.ShowSize) || (m.settingsCursor == 9 && !m.cfg.ShowDateModified) {
					if m.settingsCursor < numSettings-1 {
						m.settingsCursor++
					}
				}
			}
		case "enter", "right", "l", " ":
			m.toggleSetting(m.settingsCursor)
			m.cfg.Save()
		case "left", "h":
			m.toggleSettingPrev(m.settingsCursor)
			m.cfg.Save()
		}
	}
	return m, nil
}

func (m *Model) toggleSetting(idx int) {
	switch idx {
	case 0:
		m.cfg.ShowHidden = !m.cfg.ShowHidden
	case 1:
		m.cfg.CaseSensitive = !m.cfg.CaseSensitive
	case 2:
		m.cfg.ConfirmOperations = !m.cfg.ConfirmOperations
	case 3:
		m.cfg.WrapNavigation = !m.cfg.WrapNavigation
	case 4:
		m.cfg.ShowHeader = !m.cfg.ShowHeader
	case 5:
		m.cfg.EnableGit = !m.cfg.EnableGit
	case 6:
		m.cfg.ShowSize = !m.cfg.ShowSize
	case 7:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex + 1) % len(files.SizeFormats)
		}
	case 8:
		m.cfg.ShowDateModified = !m.cfg.ShowDateModified
	case 9:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex + 1) % len(files.DateFormats)
		}
	case 10:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex + 1) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	}
}

func (m *Model) toggleSettingPrev(idx int) {
	switch idx {
	case 7:
		if m.cfg.ShowSize {
			m.cfg.SizeFormatIndex = (m.cfg.SizeFormatIndex - 1 + len(files.SizeFormats)) % len(files.SizeFormats)
		}
	case 9:
		if m.cfg.ShowDateModified {
			m.cfg.DateFormatIndex = (m.cfg.DateFormatIndex - 1 + len(files.DateFormats)) % len(files.DateFormats)
		}
	case 10:
		m.cfg.ThemeIndex = (m.cfg.ThemeIndex - 1 + len(Themes)) % len(Themes)
		m.styles = NewStylesheet(Themes[m.cfg.ThemeIndex])
		m.updateThemeStyles()
	default:
		m.toggleSetting(idx)
	}
}

func (m *Model) performDelete() []tea.Cmd {
	var targets []string
	for _, item := range m.items {
		if item.Selected {
			targets = append(targets, item.Path)
		}
	}
	if len(targets) == 0 && len(m.filteredItems) > 0 {
		sel := m.filteredItems[m.cursor]
		if !sel.IsUp {
			targets = append(targets, sel.Path)
		}
	}

	for _, t := range targets {
		files.Delete(t)
	}
	m.setMsg(fmt.Sprintf("Deleted %d items", len(targets)))
	return []tea.Cmd{m.reload()}
}

func (m *Model) performPaste() []tea.Cmd {
	for _, src := range m.clipboard {
		dst := filepath.Join(m.path, filepath.Base(src))
		files.Copy(src, dst)
	}
	m.setMsg(fmt.Sprintf("Pasted %d items", len(m.clipboard)))
	return []tea.Cmd{m.reload()}
}

func (m *Model) getViewportHeight() int {
	h := m.height - 2
	if h < 5 {
		return 5
	}
	return h
}
