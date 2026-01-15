package tui

import (
	"fm/internal/tui/context"
	"fm/internal/tui/handlers"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// App implements tea.Model
type App struct {
	Model *context.Model
}

// NewApp creates a new Bubble Tea application
func NewApp(m *context.Model) *App {
	return &App{Model: m}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return a.Initialize()
}

// Update handles messages and returns commands
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := handlers.HandleUpdate(a.Model, msg)
	return a, cmd
}

// View renders the application UI
func (a *App) View() string {
	return view.Render(a.Model)
}
