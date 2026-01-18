package footer

import (
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/components/views"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Mode represents the current footer display mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeProgress
	ModeSearching
	ModeRenaming
	ModeGoto
	ModeAuth
	ModeFuzzySearch
	ModeConfirming
	ModeHostConfirm
	ModeMessage
	ModeSettings
	ModeLog
	ModeClipboard
	ModeZip
	ModeUnzip
)

// Props contains all data needed to render the footer
type Props struct {
	Mode  Mode
	Width int

	// Progress
	ProgressLabel   string
	ProgressPercent float64

	// Inputs
	ActiveInput ui.Input
	AltMode     bool

	// Status
	RemoteConnected bool
	Message         string
	SortMode        sorting.SortMode
	Cursor          int
	TotalItems      int
	SelectedCount   int
	Items           []core.Item
	FilteredItems   []core.Item

	// Settings
	SettingsCursor int

	// Confirming
	ActionType           constants.ActionType
	ClipboardCount       int
	ConflictDst          string
	ConflictPendingCount int
	HostConfirmReq       *ssh.HostConfirmRequest

	Styles      theme.Stylesheet
	PromptCache map[string]string
}

// Render assembles the footer by delegating to mode-specific functions
func Render(props Props) string {
	switch props.Mode {
	case ModeProgress:
		return renderProgressFooter(props)
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth, ModeFuzzySearch, ModeZip, ModeUnzip, ModeConfirming, ModeHostConfirm:
		return renderPromptsFooter(props)
	case ModeMessage:
		return renderAlertFooter(props)
	case ModeSettings:
		return views.RenderSettingsFooter(props.Width, props.SettingsCursor, props.Styles)
	case ModeLog:
		return views.RenderLogsFooter(props.Width, props.Styles)
	case ModeClipboard:
		return views.RenderClipboardFooter(props.Width, props.ClipboardCount == 0, props.Styles)
	default:
		return renderStatsFooter(props)
	}
}

func renderPromptsFooter(props Props) string {
	msgMode := messages.ModeNone
	switch props.Mode {
	case ModeSearching:
		msgMode = messages.ModeSearching
	case ModeRenaming:
		msgMode = messages.ModeRenaming
	case ModeGoto:
		msgMode = messages.ModeGoto
	case ModeAuth:
		msgMode = messages.ModeAuth
	case ModeFuzzySearch:
		msgMode = messages.ModeFuzzySearch
	case ModeZip:
		msgMode = messages.ModeZip
	case ModeUnzip:
		msgMode = messages.ModeUnzip
	case ModeConfirming:
		msgMode = messages.ModeConfirming
	case ModeHostConfirm:
		msgMode = messages.ModeHostConfirm
	}

	return messages.Render(messages.Props{
		Mode:                 msgMode,
		Width:                props.Width,
		ActiveInput:          props.ActiveInput,
		AltMode:              props.AltMode,
		RemoteConnected:      props.RemoteConnected,
		ActionType:           props.ActionType,
		ClipboardCount:       props.ClipboardCount,
		ConflictDst:          props.ConflictDst,
		ConflictPendingCount: props.ConflictPendingCount,
		HostConfirmReq:       props.HostConfirmReq,
		Style:                props.Styles,
		PromptCache:          props.PromptCache,
	})
}

func renderAlertFooter(props Props) string {
	return messages.Render(messages.Props{
		Mode:    messages.ModeAlert,
		Width:   props.Width,
		Message: props.Message,
		Style:   props.Styles,
	})
}

func renderProgressFooter(props Props) string {
	return ui.ProgressBar(props.ProgressLabel, props.ProgressPercent, props.Width, props.Styles)
}
