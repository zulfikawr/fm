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
			item := m.filteredItems[idx]
			if !item.IsUp {
				newSelected := !item.Selected
				m.filteredItems[idx].Selected = newSelected

				// Update original items and selectedPaths map
				for i := range m.items {
					if m.items[i].Path == item.Path {
						m.items[i].Selected = newSelected
						if newSelected {
							m.selectedPaths[item.Path] = true
						} else {
							delete(m.selectedPaths, item.Path)
						}
						break
					}
				}
			}
		}

		// Update selectMode based on whether anything is selected
		m.selectMode = len(m.selectedPaths) > 0

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
		m.clipboardCut = false
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
		cmds = append(cmds, m.setMsg(fmt.Sprintf("Copied %d items to clipboard", len(m.clipboard))))

	case "x":
		if m.readOnly {
			return []tea.Cmd{m.setMsg("Error: Read-only filesystem")}
		}
		m.clipboardCut = true
		m.actionType = "cut"
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
		cmds = append(cmds, m.setMsg(fmt.Sprintf("Cut %d items to clipboard", len(m.clipboard))))

	case "v":
		if m.readOnly {
			return []tea.Cmd{m.setMsg("Error: Read-only filesystem")}
		}
		if len(m.clipboard) > 0 {
			if m.cfg.ConfirmOperations {
				m.confirming = true
				m.actionType = "paste"
			} else {
				cmds = append(cmds, m.performPaste()...)
			}
		}

	case "r":
		if m.readOnly {
			return []tea.Cmd{m.setMsg("Error: Read-only filesystem")}
		}
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
		if m.readOnly {
			return []tea.Cmd{m.setMsg("Error: Read-only filesystem")}
		}
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
		m.settingsOffset = 0
	}

	return cmds
}
