package tui

import (
	"fmt"

	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleRenaming(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
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
				newPath := m.fs.Join(m.path, newName)
				if err := files.Rename(m.fs, oldPath, newPath); err != nil {
					m.LogError(err, "Rename")
				} else {
					cmd = m.reload()
					m.LogInfo(fmt.Sprintf("Renamed %s to %s", selected.Name, newName))
				}
			}
			m.renaming = false
			m.renameInput.Blur()
			return m, cmd
		}
	}
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

func (m *Model) handleConfirming(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			var cmds []tea.Cmd
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

	if len(targets) == 0 {
		return nil
	}

	m.loading = true
	m.setMsg(fmt.Sprintf("Deleting %d items...", len(targets)))
	return []tea.Cmd{deleteItems(m.fs, targets, m.cfg.UseTrash), m.reload()}
}

func (m *Model) performPaste() []tea.Cmd {
	if len(m.clipboard) == 0 {
		return nil
	}

	m.loading = true
	m.setMsg(fmt.Sprintf("Pasting %d items...", len(m.clipboard)))
	return []tea.Cmd{pasteItems(m.fs, m.clipboard, m.path), m.reload()}
}
