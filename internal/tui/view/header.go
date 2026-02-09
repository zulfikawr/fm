package view

import (
	"github.com/zulfikawr/fm/internal/tui/components/header"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func renderHeader(m *context.Model, layout context.Layout) string {
	styles := m.Display.Styles

	gitBranch := m.Navigation.Git.Branch
	gitModified := m.Navigation.Git.Modified
	gitStaged := m.Navigation.Git.Staged
	gitUntracked := m.Navigation.Git.Untracked

	// Hide git status when in analyze mode
	if m.UI.ActiveView == context.ViewAnalyze {
		gitBranch = ""
		gitModified = 0
		gitStaged = 0
		gitUntracked = 0
	}

	return header.Render(header.Props{
		Width:        layout.Width,
		Path:         m.Navigation.Path,
		Separator:    m.FS.Separator(),
		RemoteStr:    formatRemoteStr(m),
		RootOverride: formatArchiveRoot(m),
		ReadOnly:     m.Display.ReadOnly,
		Tabs: header.TabInfo{
			Count:  len(m.Tabs),
			Active: m.ActiveTab,
		},
		Git: header.GitStatusInfo{
			Branch:    gitBranch,
			Modified:  gitModified,
			Staged:    gitStaged,
			Untracked: gitUntracked,
		},
		ActiveView: m.UI.ActiveView,
		Style:      styles,
	})
}
