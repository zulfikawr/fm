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
	ModeConfirming
	ModeHostConfirm
	ModeAlert
)

// Props contains all data needed to render any message or prompt
type Props struct {
	Mode  Mode
	Width int

	// Inputs
	ActiveInput ui.Input
	AltMode     bool

	// Status
	RemoteConnected bool
	Message         string

	// Confirming
	ActionType           constants.ActionType
	ClipboardCount       int
	ConflictDst          string
	ConflictPendingCount int
	HostConfirmReq       *ssh.HostConfirmRequest

	Style       theme.Stylesheet
	PromptCache map[string]string
}

// Render renders the appropriate message or prompt based on mode
func Render(props Props) string {
	switch props.Mode {
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth, ModeFuzzySearch, ModeZip, ModeUnzip:
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
