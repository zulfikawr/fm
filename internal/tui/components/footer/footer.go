package footer

import (
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/components/views"
	"github.com/zulfikawr/fm/internal/tui/context"
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
	ModeHelp
	ModeLog
	ModeClipboard
	ModeTrash
	ModeZip
	ModeUnzip
	ModeCreate
	ModeConflictRename
	ModeKeybinding
	ModeAnalyze
)

// ProgressProps holds progress bar data
type ProgressProps struct {
	Label   string
	Percent float64
}

// InputContext holds text input state
type InputContext struct {
	Active      ui.Input
	AltMode     bool
	PromptCache map[string]string
}

// StatusInfo holds general status and navigation data
type StatusInfo struct {
	Connected     bool
	Message       string
	SortMode      sorting.SortMode
	ShowRAM       bool
	Cursor        int
	TotalItems    int
	SelectedCount int
	Items         []core.Item
	FilteredItems []core.Item
	TrashCount    int
}

// ConfirmContext holds data for confirmation prompts
type ConfirmContext struct {
	ActionType     constants.ActionType
	ClipboardCount int
	ClipboardPaths []string
	ConflictDst    string
	ConflictCount  int
	HostReq        *ssh.HostConfirmRequest
	LatestVersion  string
}

// Props contains all data needed to render the footer
type Props struct {
	Mode       Mode
	ActiveView context.ViewMode
	Width      int
	Styles     theme.Stylesheet
	Model      *context.Model

	// Categorized sub-props
	Progress ProgressProps
	Input    InputContext
	Status   StatusInfo
	Confirm  ConfirmContext

	// View-specific data
	SettingsCursor int
	SettingsItems  []views.SettingHelpItem
}

// Render assembles the footer by delegating to mode-specific functions
func Render(props Props) string {
	switch props.Mode {
	case ModeProgress:
		return renderProgressFooter(props)
	case ModeAnalyze:
		return views.RenderAnalyzeFooter(props.Width, props.Styles)
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth, ModeFuzzySearch, ModeZip, ModeUnzip, ModeCreate, ModeConflictRename, ModeKeybinding, ModeConfirming, ModeHostConfirm:
		return renderPromptsFooter(props)
	case ModeMessage:
		return renderAlertFooter(props)
	default:
		// If NO prompt/progress mode is active, check active view
		switch props.ActiveView {
		case context.ViewSettings:
			return views.RenderSettingsFooter(props.Width, props.SettingsCursor, props.SettingsItems, props.Styles)
		case context.ViewHelp:
			return views.RenderHelpFooter(props.Width, props.Styles)
		case context.ViewLogs:
			return views.RenderLogsFooter(props.Width, props.Styles)
		case context.ViewClipboard:
			return views.RenderClipboardFooter(props.Width, props.Confirm.ClipboardCount == 0, props.Styles)
		case context.ViewTrash:
			return views.RenderTrashFooter(props.Width, props.Status.TrashCount == 0, props.Styles)
		default:
			return renderStatsFooter(props)
		}
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
	case ModeCreate:
		msgMode = messages.ModeCreate
	case ModeConflictRename:
		msgMode = messages.ModeConflictRename
	case ModeKeybinding:
		msgMode = messages.ModeKeybinding
	case ModeConfirming:
		msgMode = messages.ModeConfirming
	case ModeHostConfirm:
		msgMode = messages.ModeHostConfirm
	}

	return messages.Render(messages.Props{
		Mode:  msgMode,
		Width: props.Width,
		Style: props.Styles,
		Input: messages.InputContext{
			Active:      props.Input.Active,
			AltMode:     props.Input.AltMode,
			PromptCache: props.Input.PromptCache,
		},
		RemoteConnected: props.Status.Connected,
		Confirm: messages.ConfirmContext{
			ActionType:     props.Confirm.ActionType,
			ClipboardCount: props.Confirm.ClipboardCount,
			ClipboardPaths: props.Confirm.ClipboardPaths,
			ConflictDst:    props.Confirm.ConflictDst,
			ConflictCount:  props.Confirm.ConflictCount,
			HostConfirmReq: props.Confirm.HostReq,
			LatestVersion:  props.Confirm.LatestVersion,
		},
	})
}

func renderAlertFooter(props Props) string {
	return messages.Render(messages.Props{
		Mode:    messages.ModeAlert,
		Width:   props.Width,
		Message: props.Status.Message,
		Style:   props.Styles,
	})
}

func renderProgressFooter(props Props) string {
	return ui.ProgressBar(ui.ProgressProps{
		Label:   props.Progress.Label,
		Percent: props.Progress.Percent,
		Width:   props.Width,
		Styles:  props.Styles,
	})
}
