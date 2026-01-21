package nav

import (
	tui_context "github.com/zulfikawr/fm/internal/tui/context"

	tea "github.com/charmbracelet/bubbletea"
)

// NavigateBack moves to the previous path in history
func NavigateBack(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.BackHistory) == 0 {
		return nil
	}

	m.Navigation.ForwardHistory = append(m.Navigation.ForwardHistory, m.Navigation.Path)

	prevPath := m.Navigation.BackHistory[len(m.Navigation.BackHistory)-1]
	m.Navigation.BackHistory = m.Navigation.BackHistory[:len(m.Navigation.BackHistory)-1]

	return NavigateToPathInternal(m, prevPath)
}

// NavigateForward moves forward in history
func NavigateForward(m *tui_context.Model) tea.Cmd {
	if len(m.Navigation.ForwardHistory) == 0 {
		return nil
	}

	m.Navigation.BackHistory = append(m.Navigation.BackHistory, m.Navigation.Path)

	nextPath := m.Navigation.ForwardHistory[len(m.Navigation.ForwardHistory)-1]
	m.Navigation.ForwardHistory = m.Navigation.ForwardHistory[:len(m.Navigation.ForwardHistory)-1]

	return NavigateToPathInternal(m, nextPath)
}
