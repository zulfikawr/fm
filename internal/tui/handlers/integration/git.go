package integration

import (
	"github.com/zulfikawr/fm/internal/files/core"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleGit handles git-related messages
func HandleGit(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case messages.GitStatusMsg:
		return applyGitStatus(m, msg)
	}
	return nil
}

func applyGitStatus(m *tui_context.Model, msg messages.GitStatusMsg) tea.Cmd {
	if msg.Path != m.Navigation.Path {
		return nil
	}
	m.Git.Branch = msg.Branch

	// Cache git status for this directory
	m.Cache.GitStatusCache.Put(msg.Path, msg.Statuses)

	for i, item := range m.Navigation.Items {
		if status, ok := msg.Statuses[item.Name]; ok {
			m.Navigation.Items[i].GitStatus = status
		}
	}

	// Handle ghost entries (deleted in git but not on disk)
	seen := make(map[string]bool)
	for _, item := range m.Navigation.Items {
		seen[item.Name] = true
	}

	for name, status := range msg.Statuses {
		if !seen[name] && status == "D" {
			m.Navigation.Items = append(m.Navigation.Items, core.Item{
				Name:      name,
				Path:      m.FS.Join(m.Navigation.Path, name),
				IsDir:     false,
				GitStatus: "D",
				IsGhost:   true,
				Size:      0,
			})
		}
	}

	// We might need to re-apply filter here if searching
	return nil
}
