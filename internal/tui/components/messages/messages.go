package messages

import (
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Mode represents the type of message/interaction in the footer
type Mode int

const (
	ModeNone Mode = iota
	ModeSearching
	ModeRenaming
	ModeGoto
	ModeAuth
	ModeFuzzySearch
	ModeZip
	ModeUnzip
	ModeCreate
	ModeConflictRename
	ModeKeybinding
	ModeConfirming
	ModeHostConfirm
	ModeAlert
)

// InputContext holds text input state for messages
type InputContext struct {
	Active      ui.Input
	AltMode     bool
	PromptCache map[string]string
}

// ConfirmContext holds data for confirmation messages
type ConfirmContext struct {
	ActionType     constants.ActionType
	ClipboardCount int
	ClipboardPaths []string
	ConflictDst    string
	ConflictCount  int
	HostConfirmReq *ssh.HostConfirmRequest
	LatestVersion  string
}

// Props contains all data needed to render any message or prompt
type Props struct {
	Mode  Mode
	Width int
	Style theme.Stylesheet

	// Categorized sub-props
	Input   InputContext
	Confirm ConfirmContext

	// Single fields
	RemoteConnected bool
	Message         string
}

// Render renders the appropriate message or prompt based on mode
func Render(props Props) string {
	switch props.Mode {
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth, ModeFuzzySearch, ModeZip, ModeUnzip, ModeCreate, ModeConflictRename, ModeKeybinding:
		return RenderInputPrompt(props)
	case ModeConfirming:
		return RenderConfirmationPrompt(props)
	case ModeHostConfirm:
		return RenderHostConfirmPrompt(props)
	case ModeAlert:
		return RenderAlert(props)
	default:
		return ""
	}
}
