package update

import (
	"fmt"

	"fm/internal/constants"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleAction handles action key events
func HandleAction(msg tea.KeyMsg, m *state.Model) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
	case " ":
		filtered := m.Navigation.FilteredItems
		if len(filtered) > 0 {
			idx := m.Navigation.Cursor
			item := filtered[idx]
			if !item.IsUp {
				newSelected := !item.Selected
				filtered[idx].Selected = newSelected

				// Update original items and selectedPaths map
				items := m.Navigation.Items
				for i := range items {
					if items[i].Path == item.Path {
						items[i].Selected = newSelected
						if newSelected {
							m.Operations.SelectedPaths[item.Path] = true
							m.Navigation.SelectedCount++
						} else {
							delete(m.Operations.SelectedPaths, item.Path)
							m.Navigation.SelectedCount--
						}
						break
					}
				}
			}
		}

		// Update selectMode based on whether anything is selected
		m.UI.SelectMode = len(m.Operations.SelectedPaths) > 0

	case "s":
		m.Display.SortMode = (m.Display.SortMode + 1) % 7
		cmds = append(cmds, actions.Reload(m))

	case "/":
		return []tea.Cmd{actions.OpenPrompt(m, state.InputSearch, "")}

	case "c":
		m.Operations.Clipboard.IsCut = false
		m.Operations.ActionType = constants.ActionCopy
		m.Operations.Clipboard.Paths = []string{}

		clipboard := []string{}
		for _, item := range m.Navigation.Items {
			if item.Selected {
				clipboard = append(clipboard, item.Path)
			}
		}
		if len(clipboard) == 0 && len(m.Navigation.FilteredItems) > 0 {
			sel := m.Navigation.FilteredItems[m.Navigation.Cursor]
			if !sel.IsUp {
				clipboard = append(clipboard, sel.Path)
			}
		}
		m.Operations.Clipboard.Paths = clipboard
		cmds = append(cmds, commands.SetMsg(m, fmt.Sprintf("Copied %d items to clipboard", len(clipboard))))

	case "x":
		if m.Display.ReadOnly {
			return []tea.Cmd{commands.SetMsg(m, "Error: Read-only filesystem")}
		}
		m.Operations.Clipboard.IsCut = true
		m.Operations.ActionType = constants.ActionCut

		clipboard := []string{}
		for _, item := range m.Navigation.Items {
			if item.Selected {
				clipboard = append(clipboard, item.Path)
			}
		}
		if len(clipboard) == 0 && len(m.Navigation.FilteredItems) > 0 {
			sel := m.Navigation.FilteredItems[m.Navigation.Cursor]
			if !sel.IsUp {
				clipboard = append(clipboard, sel.Path)
			}
		}
		m.Operations.Clipboard.Paths = clipboard
		cmds = append(cmds, commands.SetMsg(m, fmt.Sprintf("Cut %d items to clipboard", len(clipboard))))

	case "v":
		if m.Display.ReadOnly {
			return []tea.Cmd{commands.SetMsg(m, "Error: Read-only filesystem")}
		}
		if len(m.Operations.Clipboard.Paths) > 0 {
			if m.Config.ConfirmOperations {
				m.UI.Confirming = true
				m.Operations.ActionType = constants.ActionPaste
			} else {
				cmds = append(cmds, PerformPaste(m)...)
			}
		}

	case "r":
		if m.Display.ReadOnly {
			return []tea.Cmd{commands.SetMsg(m, "Error: Read-only filesystem")}
		}
		if len(m.Navigation.FilteredItems) > 0 {
			sel := m.Navigation.FilteredItems[m.Navigation.Cursor]
			if !sel.IsUp {
				return []tea.Cmd{actions.OpenPrompt(m, state.InputRename, sel.Name)}
			}
		}

	case "d":
		if m.Display.ReadOnly {
			return []tea.Cmd{commands.SetMsg(m, "Error: Read-only filesystem")}
		}
		selectedCount := 0
		for _, item := range m.Navigation.Items {
			if item.Selected {
				selectedCount++
			}
		}
		if selectedCount > 0 || (len(m.Navigation.FilteredItems) > 0 && !m.Navigation.FilteredItems[m.Navigation.Cursor].IsUp) {
			if m.Config.ConfirmOperations {
				m.UI.Confirming = true
				m.Operations.ActionType = constants.ActionDelete
			} else {
				cmds = append(cmds, PerformDelete(m)...)
			}
		}

	case ".":
		m.UI.SettingsOpen = true
		m.Settings.Cursor = 0
		m.Settings.Offset = 0

	case "g":
		return []tea.Cmd{actions.OpenPrompt(m, state.InputGoto, m.Navigation.Path)}
	}

	return cmds
}
