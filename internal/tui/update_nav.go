package tui

import (
	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleNavigation(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
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
			m.path = m.fs.Dir(m.path)
			cmds = append(cmds, m.reload())
		} else if selected.IsDir {
			m.cursorMemory[m.path] = m.cursor
			m.offsetMemory[m.path] = m.offset
			m.path = m.fs.Join(m.path, selected.Name)
			cmds = append(cmds, m.reload())
		} else {
			// Handle file opening
			if msg.String() == "enter" {
				if !m.fs.IsLocal() {
					m.setMsg("Opening remote files not supported yet")
					break
				}
				execCmd, isTerminal, err := files.GetOpenCmd(selected.Path, m.cfg.EditorIndex)
				if err != nil {
					m.setMsg("Error: " + err.Error())
					break
				}

				if isTerminal {
					return []tea.Cmd{tea.ExecProcess(execCmd, func(err error) tea.Msg {
						if err != nil {
							return errMsg{err}
						}
						return nil
					})}
				} else {
					if err := execCmd.Start(); err != nil {
						m.setMsg("Error: " + err.Error())
					}
				}
			}
		}

	case "backspace", "left", "h":
		m.cursorMemory[m.path] = m.cursor
		m.offsetMemory[m.path] = m.offset
		m.path = m.fs.Dir(m.path)
		cmds = append(cmds, m.reload())
	}

	return cmds
}
