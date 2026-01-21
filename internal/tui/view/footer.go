package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/footer"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func renderFooter(m *context.Model, layout context.Layout) string {
	styles := m.Display.Styles

	return footer.Render(footer.Props{
		Mode:                 determineFooterMode(m),
		Width:                layout.Width,
		ProgressLabel:        m.Operations.Progress.Label,
		ProgressPercent:      m.Operations.Progress.Percent,
		ActiveInput:          m.Inputs.ActiveInput,
		AltMode:              m.Inputs.AltMode,
		RemoteConnected:      !m.FS.IsLocal(),
		Message:              m.Message.Text,
		SortMode:             m.Display.SortMode,
		Cursor:               m.Navigation.Cursor,
		TotalItems:           len(m.Navigation.FilteredItems),
		SelectedCount:        m.Navigation.SelectedCount,
		Items:                m.Navigation.Items,
		FilteredItems:        m.Navigation.FilteredItems,
		SettingsCursor:       m.Settings.Cursor,
		ActionType:           m.Operations.ActionType,
		ClipboardCount:       len(m.Operations.Clipboard.Paths),
		ConflictDst:          m.Operations.Conflict.Destination,
		ConflictPendingCount: len(m.Operations.Conflict.PendingItems),
		HostConfirmReq:       m.Remote.HostConfirmReq,
		LatestVersion:        m.UI.LatestVersion,
		Styles:               styles,
		PromptCache:          m.UI.PromptCache,
	})
}

func determineFooterMode(m *context.Model) footer.Mode {
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
	if m.UI.SettingsOpen {
		return footer.ModeSettings
	}
	if m.UI.LogOpen {
		return footer.ModeLog
	}
	if m.UI.ClipboardOpen {
		return footer.ModeClipboard
	}
	return footer.ModeNormal
}
