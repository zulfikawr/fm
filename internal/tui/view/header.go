package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/header"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func renderHeader(m *context.Model, layout context.Layout) string {
	styles := m.Display.Styles

	gitBranch := m.Git.Branch
	gitModified := m.Git.Modified
	gitStaged := m.Git.Staged
	gitUntracked := m.Git.Untracked

	// Hide git status when in analyze mode as requested
	if m.UI.AnalyzeOpen {
		gitBranch = ""
		gitModified = 0
		gitStaged = 0
		gitUntracked = 0
	}

	return header.Render(header.Props{
		Width:         layout.Width,
		Path:          m.Navigation.Path,
		Separator:     m.FS.Separator(),
		RemoteStr:     formatRemoteStr(m),
		RootOverride:  formatArchiveRoot(m),
		GitBranch:     gitBranch,
		GitModified:   gitModified,
		GitStaged:     gitStaged,
		GitUntracked:  gitUntracked,
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
