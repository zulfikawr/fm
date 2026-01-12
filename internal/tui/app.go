package tui

import (
	"fm/internal/files"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
	"fm/internal/tui/update"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// App wraps the state.Model to implement tea.Model.
type App struct {
	*state.Model
}

// NewApp creates a new Bubble Tea application with the given initial path.
func NewApp(fs files.FileSystem, initialPath string) *App {
	m := NewModel(fs, initialPath)
	return &App{m}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return a.ModelInit()
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return a, update.Update(a.Model, msg)
}

// View implements tea.Model.
func (a *App) View() string {
	return a.ModelView()
}

// ModelView renders the application UI.
func (a *App) ModelView() string {
	s := a.ViewState()
	styles := theme.GetStylesheet(a.Config.ThemeIndex)
	return view.Render(&s, styles)
}

// ViewState constructs the current view state from the model.
func (a *App) ViewState() view.ViewState {
	return view.GetViewState(a.Model)
}
