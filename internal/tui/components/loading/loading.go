package loading

import (
	"fmt"

	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// Props contains all data needed to render the loading state
type Props struct {
	Width   int
	Height  int
	Message string
	Spinner ui.Spinner
	Style   theme.Stylesheet
}

// Render renders the loading state with a spinner
func Render(props Props) string {
	content := fmt.Sprintf("%s %s", props.Spinner.View(), props.Message)

	return lipgloss.NewStyle().
		Height(props.Height).
		Width(props.Width).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}
