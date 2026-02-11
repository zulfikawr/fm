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
	if m.UI.ActiveView == context.ViewSettings {
		items := app.BuildFullSettingList(m)
		settingsItems = make([]views.SettingHelpItem, len(items))
		for i, item := range items {
			settingsItems[i] = views.SettingHelpItem{HelpText: item.HelpText}
		}
	}

	return footer.Render(footer.Props{
		Mode:       utils.DetermineFooterMode(m),
		ActiveView: m.UI.ActiveView,
		Width:      layout.Width,
		Styles:     styles,
		Model:      m,
		Progress: footer.ProgressProps{
			Label:   m.Operations.Progress.Label,
			Percent: m.Operations.Progress.Percent,
		},
		Input: footer.InputContext{
			Active:      m.Inputs.ActiveInput,
			AltMode:     m.Inputs.AltMode,
			PromptCache: m.UI.PromptCache,
		},
		Status: footer.StatusInfo{
			Connected:     !m.FS.IsLocal(),
			Message:       m.Message.Text,
			SortMode:      m.Display.SortMode,
			ShowRAM:       m.Config.UI.ShowRAMUsage,
			Cursor:        m.Navigation.Cursor,
			TotalItems:    len(m.Navigation.FilteredItems),
			SelectedCount: m.Navigation.SelectedCount(),
			Items:         m.Navigation.Items,
			FilteredItems: m.Navigation.FilteredItems,
			TrashCount:    len(m.Trash.Items),
		},
		Confirm: footer.ConfirmContext{
			ActionType:     m.Operations.ActionType,
			ClipboardCount: len(m.Operations.Clipboard.Paths),
			ClipboardPaths: m.Operations.Clipboard.Paths,
			ConflictDst:    m.Operations.Conflict.Destination,
			ConflictCount:  len(m.Operations.Conflict.PendingItems),
			HostReq:        m.Navigation.Remote.HostConfirmReq,
			LatestVersion:  m.UI.LatestVersion,
			PendingValue:   m.Operations.PendingOp.Value,
		},
		SettingsCursor: m.Settings.Cursor,
		SettingsItems:  settingsItems,
	})
}
