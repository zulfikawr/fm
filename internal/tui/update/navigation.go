package update

import (
	"fm/internal/files/ops"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleNavigation handles navigation key events
func HandleNavigation(msg tea.KeyMsg, m *state.Model) []tea.Cmd {
	var cmds []tea.Cmd

	viewportHeight := m.Display.ViewportHeight
	if viewportHeight <= 0 {
		// Fallback if not yet calculated
		s := view.GetViewState(m)
		viewportHeight = view.GetViewportHeight(&s)
	}

	switch msg.String() {
	case "up", "k":
		if m.Navigation.Cursor > 0 {
			m.Navigation.Cursor--
		} else if m.Config.WrapNavigation && len(m.Navigation.FilteredItems) > 0 {
			m.Navigation.Cursor = len(m.Navigation.FilteredItems) - 1
		}
		m.Navigation.Offset = scroll(m.Navigation.Cursor, m.Navigation.Offset, viewportHeight)

	case "down", "j":
		if m.Navigation.Cursor < len(m.Navigation.FilteredItems)-1 {
			m.Navigation.Cursor++
		} else if m.Config.WrapNavigation && len(m.Navigation.FilteredItems) > 0 {
			m.Navigation.Cursor = 0
		}
		m.Navigation.Offset = scroll(m.Navigation.Cursor, m.Navigation.Offset, viewportHeight)

	case "enter", "right", "l":
		if len(m.Navigation.FilteredItems) == 0 {
			break
		}
		selected := m.Navigation.FilteredItems[m.Navigation.Cursor]

		if selected.IsDir {
			cmds = append(cmds, actions.NavigateToSelected(m))
		} else {
			// Handle file opening
			if msg.String() == "enter" {
				if !m.FS.IsLocal() {
					return []tea.Cmd{commands.SetMsg(m, "Opening remote files not supported yet")}
				}
				execCmd, isTerminal, err := ops.GetOpenCmd(selected.Path, m.Config.EditorIndex)
				if err != nil {
					return []tea.Cmd{commands.SetMsg(m, "Error: "+err.Error())}
				}

				if isTerminal {
					return []tea.Cmd{tea.ExecProcess(execCmd, func(err error) tea.Msg {
						if err != nil {
							return commands.ErrorMsg{Err: err}
						}
						return nil
					})}
				} else {
					if err := execCmd.Start(); err != nil {
						cmds = append(cmds, commands.SetMsg(m, "Error: "+err.Error()))
					}
				}
			}
		}

	case "backspace", "left", "h":
		cmds = append(cmds, actions.NavigateToParent(m))
	}

	return cmds
}

// scroll calculates the new offset to ensure cursor is visible
func scroll(cursor, offset, viewportHeight int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+viewportHeight {
		return cursor - viewportHeight + 1
	}
	return offset
}
