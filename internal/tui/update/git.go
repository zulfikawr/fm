package update

import (
	"fm/internal/files"
	"fm/internal/tui/commands"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleGitMsg delegates git-related messages to specialized handlers
func HandleGitMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.GitStatusMsg:
		HandleGitStatus(m, msg)
		filter.Apply(m)
		return nil
	}
	return nil
}

// HandleGitStatus updates the model with the received git status
func HandleGitStatus(m *state.Model, msg commands.GitStatusMsg) {
	if msg.Path != m.Navigation.Path {
		return
	}
	m.Git.Branch = msg.Branch

	statusMap := msg.Statuses
	// Cache git status for this directory
	m.Cache.GitStatusCache[msg.Path] = statusMap

	for i, item := range m.Navigation.Items {
		if status, ok := statusMap[item.Name]; ok {
			m.Navigation.Items[i].GitStatus = status
		}
	}

	seen := make(map[string]bool)
	for _, item := range m.Navigation.Items {
		seen[item.Name] = true
	}
	for name, status := range statusMap {
		if !seen[name] && status == "D" {
			m.Navigation.Items = append(m.Navigation.Items, files.Item{
				Name:      name,
				Path:      m.FS.Join(m.Navigation.Path, name),
				IsDir:     false,
				GitStatus: "D",
				IsGhost:   true,
				Size:      0,
			})
		}
	}
}
