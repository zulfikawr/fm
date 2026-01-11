package tui

import (
	"context"
	"fmt"
	"time"

	"fm/internal/files"

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
				// Validate filename
				if err := files.ValidateFileName(newName); err != nil {
					cmd := m.LogError(err, "Invalid filename")
					m.renaming = false
					m.renameInput.Blur()
					return m, cmd
				}

				selected := m.filteredItems[m.cursor]
				oldPath := selected.Path

				if m.processingItems[oldPath] {
					return m, m.setMsg("Error: Item is currently being processed")
				}

				newPath := m.fs.Join(m.path, newName)

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := files.Rename(ctx, m.fs, oldPath, newPath); err != nil {
					cmd = m.LogError(err, "Rename")
				} else {
					cmd = tea.Batch(m.reload(), m.LogInfo(fmt.Sprintf("Renamed %s to %s", selected.Name, newName)))
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
			switch m.actionType {
			case "delete":
				cmds = append(cmds, m.performDelete()...)
			case "paste":
				cmds = append(cmds, m.performPaste()...)
			case "conflict":
				return m.handleConflict("overwrite")
			}
			m.confirming = false
			return m, tea.Batch(cmds...)
		case "n", "N", "esc":
			if m.actionType == "conflict" {
				return m.handleConflict("skip")
			}
			m.confirming = false
			return m, nil
		case "r", "R":
			if m.actionType == "conflict" {
				return m.handleConflict("rename")
			}
		}
	}
	return m, nil
}

func (m *Model) handleConflict(choice string) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch choice {
	case "overwrite":
		cmds = append(cmds, overwriteItem(m.fs, m.conflictSrc, m.conflictDst, m.clipboardCut))
	case "skip":
		// Remove from processingItems since we're skipping it
		delete(m.processingItems, m.conflictSrc)
	case "rename":
		// Auto-rename
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ext := ""
		base := m.conflictDst
		for i := len(m.conflictDst) - 1; i >= 0 && m.conflictDst[i] != '/'; i-- {
			if m.conflictDst[i] == '.' {
				ext = m.conflictDst[i:]
				base = m.conflictDst[:i]
				break
			}
		}

		newName := ""
		for i := 1; ; i++ {
			newName = fmt.Sprintf("%s (%d)%s", base, i, ext)
			if _, err := m.fs.Stat(ctx, newName); err != nil {
				break
			}
		}
		cmds = append(cmds, overwriteItem(m.fs, m.conflictSrc, newName, m.clipboardCut))
	}

	m.confirming = false
	m.actionType = ""

	// Continue with pending items if any
	if len(m.pendingItems) > 0 {
		m.loading = true
		if m.clipboardCut {
			cmds = append(cmds, moveItems(m.fs, m.pendingItems, m.path))
		} else {
			cmds = append(cmds, pasteItems(m.fs, m.pendingItems, m.path))
		}
	} else {
		// If no pending items, reload will happen after the overwriteItem/moveItems completes
		// via OperationFinishedMsg (though currently we just reload here if skipping)
		if choice == "skip" {
			cmds = append(cmds, m.reload())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) checkAndMarkProcessing(paths []string) bool {
	for _, p := range paths {
		if m.processingItems[p] {
			return false
		}
	}
	for _, p := range paths {
		m.processingItems[p] = true
	}
	return true
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

	if !m.checkAndMarkProcessing(targets) {
		return []tea.Cmd{m.setMsg("Error: Some items are already being processed")}
	}

	m.loading = true
	return []tea.Cmd{m.setMsg(fmt.Sprintf("Deleting %d items...", len(targets))), deleteItems(m.fs, targets, m.cfg.UseTrash), m.reload()}
}

func (m *Model) performPaste() []tea.Cmd {
	if len(m.clipboard) == 0 {
		return nil
	}

	if !m.checkAndMarkProcessing(m.clipboard) {
		return []tea.Cmd{m.setMsg("Error: Some items are already being processed")}
	}

	m.loading = true
	if m.clipboardCut {
		cmds := []tea.Cmd{
			m.setMsg(fmt.Sprintf("Moving %d items...", len(m.clipboard))),
			moveItems(m.fs, m.clipboard, m.path),
			m.reload(),
		}
		m.clipboard = []string{}
		m.clipboardCut = false
		return cmds
	}
	return []tea.Cmd{m.setMsg(fmt.Sprintf("Pasting %d items...", len(m.clipboard))), pasteItems(m.fs, m.clipboard, m.path), m.reload()}
}
