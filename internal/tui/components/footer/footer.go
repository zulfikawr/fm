package footer

import (
	"strings"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"fm/internal/sshutil"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
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
	ModeConfirming
	ModeHostConfirm
	ModeMessage
	ModeSettings
)

// Props contains all data needed to render the footer
type Props struct {
	Mode  Mode
	Width int

	// Progress
	ProgressLabel   string
	ProgressPercent float64

	// Inputs
	ActiveInput textinput.Model
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
	ActionType     constants.ActionType
	ClipboardCount int
	ConflictDst    string
	HostConfirmReq *sshutil.HostConfirmRequest

	Styles      theme.Stylesheet
	PromptCache map[string]string
}

// Render renders the complete footer based on the current mode
func Render(props Props) string {
	switch props.Mode {
	case ModeProgress:
		return renderProgress(props)
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth, ModeConfirming, ModeHostConfirm:
		return renderPrompts(props)
	case ModeMessage:
		return renderMessage(props)
	case ModeSettings:
		return renderSettingsFooter(props)
	default:
		return renderNormalFooter(props)
	}
}

// buildNormalFooterParts builds all parts for normal footer display
func buildNormalFooterParts(props Props) []string {
	var parts []string

	// Pagination - always show
	pagination := renderPaginationInfo(PaginationInfo{
		Current: props.Cursor,
		Total:   props.TotalItems,
		Width:   props.Width,
	}, props.Styles)
	if pagination != "" {
		parts = append(parts, pagination)
	}

	// Permission info
	permission := renderPermissionInfo(props.FilteredItems, props.Cursor, props.Styles)
	if permission != "" {
		parts = append(parts, permission)
	}

	return parts
}

// assembleFooterContent assembles footer parts with proper spacing
func assembleFooterContent(parts []string, width int, styles theme.Stylesheet) string {
	if len(parts) == 0 {
		return ""
	}

	dimStyle := styles.DimCol.Inherit(styles.Footer).UnsetPadding().UnsetWidth()
	spacer := dimStyle.Render(" | ")
	content := strings.Join(parts, spacer)

	// Truncate if too long
	if lipgloss.Width(content) > width-2 {
		maxWidth := width - 5 // Leave room for "..."
		if maxWidth < 0 {
			maxWidth = 0
		}
		// Simple truncation
		if len(content) > maxWidth {
			content = content[:maxWidth] + "..."
		}
	}

	return content
}
