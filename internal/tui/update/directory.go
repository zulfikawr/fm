package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleLoadedItems handles directory loading completion messages
func HandleLoadedItems(m *state.Model, msg commands.LoadedItemsMsg) (tea.Cmd, bool) {
	return actions.FinalizeDirectoryLoad(m, msg)
}
