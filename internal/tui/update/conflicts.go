package update

import (
	"fm/internal/constants"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"
)

// HandleConflict handles file operation conflict messages
func HandleConflict(m *state.Model, msg commands.ConflictMsg) {
	m.UI.Loading = false
	m.UI.Confirming = true
	m.Operations.ActionType = constants.ActionConflict
	m.Operations.Conflict.Source = msg.Src
	m.Operations.Conflict.Destination = msg.Dst
	m.Operations.Conflict.PendingItems = msg.PendingItems
	m.Operations.Clipboard.IsCut = msg.IsMove
}
