package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/footer"
	"github.com/zulfikawr/fm/internal/tui/components/views"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
)

func renderFooter(m *context.Model, layout context.Layout) string {
	styles := m.Display.Styles

	// Build settings items for footer help text
	var settingsItems []views.SettingHelpItem
	if m.UI.SettingsOpen {
		items := app.BuildFullSettingList(m)
		settingsItems = make([]views.SettingHelpItem, len(items))
		for i, item := range items {
			settingsItems[i] = views.SettingHelpItem{HelpText: item.HelpText}
		}
	}

	return footer.Render(footer.Props{
		Mode:                 utils.DetermineFooterMode(m),
		Width:                layout.Width,
		ProgressLabel:        m.Operations.Progress.Label,
		ProgressPercent:      m.Operations.Progress.Percent,
		ActiveInput:          m.Inputs.ActiveInput,
		AltMode:              m.Inputs.AltMode,
		RemoteConnected:      !m.FS.IsLocal(),
		Message:              m.Message.Text,
		SortMode:             m.Display.SortMode,
		ShowRAMUsage:         m.Config.ShowRAMUsage,
		Cursor:               m.Navigation.Cursor,
		TotalItems:           len(m.Navigation.FilteredItems),
		SelectedCount:        m.Navigation.SelectedCount,
		Items:                m.Navigation.Items,
		FilteredItems:        m.Navigation.FilteredItems,
		SettingsCursor:       m.Settings.Cursor,
		SettingsItems:        settingsItems,
		ActionType:           m.Operations.ActionType,
		ClipboardCount:       len(m.Operations.Clipboard.Paths),
		ClipboardPaths:       m.Operations.Clipboard.Paths,
		ClipboardOpen:        m.UI.ClipboardOpen,
		HelpOpen:             m.UI.HelpOpen,
		ConflictDst:          m.Operations.Conflict.Destination,
		ConflictPendingCount: len(m.Operations.Conflict.PendingItems),
		HostConfirmReq:       m.Remote.HostConfirmReq,
		LatestVersion:        m.UI.LatestVersion,
		Styles:               styles,
		PromptCache:          m.UI.PromptCache,
	})
}
