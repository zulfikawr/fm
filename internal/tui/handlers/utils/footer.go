package utils

import (
	"strings"

	"github.com/zulfikawr/fm/internal/tui/components/footer"
	"github.com/zulfikawr/fm/internal/tui/context"
)

// DetermineFooterMode calculates the current footer display mode
func DetermineFooterMode(m *context.Model) footer.Mode {
	// Quit confirmation has absolute precedence
	if strings.HasPrefix(m.Message.Text, "press [") && strings.HasSuffix(m.Message.Text, "] again to close") {
		return footer.ModeMessage
	}

	if m.UI.InputActive {
		switch m.Inputs.Mode {
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
		}
	}
	if m.UI.HostConfirm {
		return footer.ModeHostConfirm
	}
	if m.UI.Confirming {
		return footer.ModeConfirming
	}
	if m.Operations.Progress.Visible {
		return footer.ModeProgress
	}
	if m.Message.Text != "" {
		return footer.ModeMessage
	}

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

	return footer.ModeNormal
}

// GetPromptLength returns the visual length of the input prompt for the current mode
func GetPromptLength(m *context.Model) int {
	switch m.Inputs.Mode {
	case context.InputGoto:
		if m.Inputs.AltMode {
			return 16 // "Go to (Remote): "
		}
		return 15 // "Go to (Local): "
	case context.InputAuth:
		if m.Inputs.AltMode {
			return 6 // "Path: "
		}
		return 10 // "Password: "
	case context.InputSearch:
		return 8 // "Filter: "
	case context.InputFuzzySearch:
		return 8 // "Search: "
	case context.InputRename:
		return 8 // "Rename: "
	case context.InputZip:
		return 10 // "Zip name: "
	case context.InputUnzip:
		return 10 // "Unzip to: "
	case context.InputCreate:
		if m.Inputs.AltMode {
			return 17 // "Create (Folder): "
		}
		return 15 // "Create (File): "
	case context.InputConflictRename:
		return 10 // "New name: "
	case context.InputKeybinding:
		return 8 // "Bind: "
	}
	return 0
}
