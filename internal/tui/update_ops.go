package tui

import (
	"fmt"
	"path/filepath"

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
				newPath := filepath.Join(m.path, newName)
				if err := files.Rename(oldPath, newPath); err != nil {
					m.setMsg("Rename failed: " + err.Error())
				} else {
					cmd = m.reload()
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

	errs := 0
	for _, t := range targets {
		if err := files.Trash(t); err != nil {
			errs++
			m.setMsg(fmt.Sprintf("Error trashing %s: %v", filepath.Base(t), err))
		}
	}
	if errs == 0 {
		m.setMsg(fmt.Sprintf("Trashed %d items", len(targets)))
	}
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
