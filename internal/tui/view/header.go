package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/header"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func renderHeader(m *context.Model, layout context.Layout) string {
	styles := m.Display.Styles

	return header.Render(header.Props{
		Width:         layout.Width,
		Path:          m.Navigation.Path,
		Separator:     m.FS.Separator(),
		RemoteStr:     formatRemoteStr(m),
		RootOverride:  formatArchiveRoot(m),
		GitBranch:     m.Git.Branch,
		ReadOnly:      m.Display.ReadOnly,
		TabCount:      len(m.Tabs),
		ActiveTab:     m.ActiveTab,
		SettingsOpen:  m.UI.SettingsOpen,
		HelpOpen:      m.UI.HelpOpen,
		LogOpen:       m.UI.LogOpen,
		ClipboardOpen: m.UI.ClipboardOpen,
		Style:         styles,
	})
}
