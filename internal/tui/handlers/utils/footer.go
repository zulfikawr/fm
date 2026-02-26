package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/tui/components/footer"
	"github.com/zulfikawr/fm/internal/tui/context"
)

// DetermineFooterMode calculates the current footer display mode
func DetermineFooterMode(m *context.Model) footer.Mode {
	// 1. System Overrides (Highest Priority)
	if isQuitConfirmation(m) {
		return footer.ModeMessage
	}
	if m.Operations.Progress.Visible {
		return footer.ModeProgress
	}

	// 2. Interactive States
	if m.UI.InputActive {
		return getInputFooterMode(m.Inputs.Mode)
	}
	if m.UI.HostConfirm {
		return footer.ModeHostConfirm
	}
	if m.UI.Confirming {
		return footer.ModeConfirming
	}

	// 3. Status Messages
	if m.Message.Text != "" {
		return footer.ModeMessage
	}

	// 4. View-specific Overlays
	switch m.UI.ActiveView {
	case context.ViewSettings:
		return footer.ModeSettings
	case context.ViewHelp:
		return footer.ModeHelp
	case context.ViewLogs:
		return footer.ModeLog
	case context.ViewClipboard:
		return footer.ModeClipboard
	case context.ViewTrash:
		return footer.ModeTrash
	case context.ViewAnalyze:
		return footer.ModeAnalyze
	}

	// 5. Default
	return footer.ModeNormal
}

// isQuitConfirmation checks if the model is currently showing a quit warning
func isQuitConfirmation(m *context.Model) bool {
	return strings.HasPrefix(m.Message.Text, "Press [") && strings.HasSuffix(m.Message.Text, "] again to close")
}

// getInputFooterMode maps InputMode to footer.Mode
func getInputFooterMode(mode context.InputMode) footer.Mode {
	switch mode {
	case context.InputSearch:
		return footer.ModeSearching
	case context.InputRename:
		return footer.ModeRenaming
	case context.InputGoto:
		return footer.ModeGoto
	case context.InputAuth:
		return footer.ModeAuth
	case context.InputFuzzySearch:
		return footer.ModeFuzzySearch
	case context.InputZip:
		return footer.ModeZip
	case context.InputUnzip:
		return footer.ModeUnzip
	case context.InputCreate:
		return footer.ModeCreate
	case context.InputConflictRename:
		return footer.ModeConflictRename
	case context.InputKeybinding:
		return footer.ModeKeybinding
	default:
		return footer.ModeNormal
	}
}

// GetPromptLength returns the visual length of the input prompt for the current mode
func GetPromptLength(m *context.Model) int {
	prompt := context.GetPromptText(m.Inputs.Mode, m.Inputs.AltMode)
	return lipgloss.Width(prompt)
}
