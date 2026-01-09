package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleAction(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
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
		cmds = append(cmds, textinput.Blink)

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
				cmds = append(cmds, textinput.Blink)
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

	return cmds
}
