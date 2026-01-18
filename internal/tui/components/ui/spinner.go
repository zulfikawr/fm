package ui

import (
	"time"

	"github.com/zulfikawr/fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TickMsg is sent when the spinner should advance to the next frame
type TickMsg struct {
	Time time.Time
}

// Spinner represents a custom theme-aware spinner
type Spinner struct {
	Frames   []string
	Index    int
	Interval time.Duration
	Style    lipgloss.Style
}

var (
	// DefaultFrames is the "Dot" spinner from bubbles
	DefaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// NewSpinner creates a new theme-aware spinner
func NewSpinner(styles theme.Stylesheet) Spinner {
	return Spinner{
		Frames:   DefaultFrames,
		Interval: 100 * time.Millisecond,
		Style:    styles.DirCol,
	}
}

// Tick returns a command that waits for the next spinner tick
func Tick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

// Update handles the tick message and advances the spinner
func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		s.Index = (s.Index + 1) % len(s.Frames)
		return s, Tick(s.Interval)
	}
	return s, nil
}

// View renders the current frame of the spinner
func (s Spinner) View() string {
	if len(s.Frames) == 0 {
		return ""
	}
	if s.Index >= len(s.Frames) {
		s.Index = 0
	}
	return s.Style.Render(s.Frames[s.Index])
}

// Start returns a command to start the spinner animation
func (s Spinner) Start() tea.Cmd {
	return Tick(s.Interval)
}
