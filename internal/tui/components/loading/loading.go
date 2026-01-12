package loading

import (
	"fmt"

	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Props contains all data needed to render the loading state
type Props struct {
	Width   int
	Height  int
	Message string
	Spinner spinner.Model
	Styles  theme.Stylesheet
}

// Render renders the loading state with a spinner
func Render(props Props) string {
	// Combine spinner and message
	content := fmt.Sprintf("%s %s", props.Spinner.View(), props.Message)

	return lipgloss.NewStyle().
		Height(props.Height).
		Width(props.Width).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

// NewSpinner creates a new theme-aware spinner
func NewSpinner(themeColors theme.Theme) spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(themeColors.Dir)
	return s
}

// UpdateSpinner updates the spinner and returns the new model and command
func UpdateSpinner(s spinner.Model, msg tea.Msg) (spinner.Model, tea.Cmd) {
	var cmd tea.Cmd
	s, cmd = s.Update(msg)
	return s, cmd
}

// UpdateSpinnerTheme updates the spinner colors based on the current theme
func UpdateSpinnerTheme(s *spinner.Model, themeColors theme.Theme) {
	s.Style = lipgloss.NewStyle().Foreground(themeColors.Dir)
}
