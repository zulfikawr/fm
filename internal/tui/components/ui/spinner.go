package ui

import (
	"time"

	"fm/internal/tui/theme"

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
func (m Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.Index = (m.Index + 1) % len(m.Frames)
		return m, Tick(m.Interval)
	}
	return m, nil
}

// View renders the current frame of the spinner
func (m Spinner) View() string {
	if len(m.Frames) == 0 {
		return ""
	}
	if m.Index >= len(m.Frames) {
		m.Index = 0
	}
	return m.Style.Render(m.Frames[m.Index])
}

// Start returns a command to start the spinner animation
func (m Spinner) Start() tea.Cmd {
	return Tick(m.Interval)
}
