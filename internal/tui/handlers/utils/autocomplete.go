package utils

import (
	"context"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/ops"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
)

// FetchCompletions returns a command that fetches path completions
func FetchCompletions(ctx context.Context, fs core.FileSystem, currentDir, input string) tea.Cmd {
	return func() tea.Msg {
		completions := ops.GetPathCompletions(ctx, fs, currentDir, input)
		return messages.CompletionsMsg{Completions: completions}
	}
}

// UpdateSearchSuggestion updates the active input's suggestion based on the current filtered list
func UpdateSearchSuggestion(m *tuictx.Model) {
	input := m.Inputs.ActiveInput.Value()
	if input == "" {
		m.Inputs.ActiveInput.Suggestion = ""
		return
	}

	inputLower := strings.ToLower(input)
	for _, item := range m.Navigation.Items {
		if item.IsUp {
			continue
		}
		nameLower := strings.ToLower(item.Name)
		if strings.HasPrefix(nameLower, inputLower) {
			m.Inputs.ActiveInput.Suggestion = item.Name
			return
		}
	}
	m.Inputs.ActiveInput.Suggestion = ""
}
