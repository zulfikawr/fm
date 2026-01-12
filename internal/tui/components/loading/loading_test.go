package loading

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestNewSpinner(t *testing.T) {
	themeColors := theme.Themes[0]

	spinner := NewSpinner(themeColors)

	if spinner.Spinner.FPS == 0 {
		t.Error("Expected spinner FPS to be set")
	}
}

func TestUpdateSpinner(t *testing.T) {
	themeColors := theme.Themes[0]
	spinner := NewSpinner(themeColors)

	// Get the tick command and execute it to get the message
	tickCmd := spinner.Tick
	msg := tickCmd()

	newSpinner, cmd := UpdateSpinner(spinner, msg)

	if cmd == nil {
		t.Error("Expected command to be returned for tick")
	}

	if newSpinner.Spinner.FPS == 0 {
		t.Error("Spinner should still be configured after update")
	}
}

func TestUpdateSpinnerTheme(t *testing.T) {
	initialTheme := theme.Themes[0]
	spinner := NewSpinner(initialTheme)

	// Change to a different theme
	newTheme := theme.Themes[1]
	UpdateSpinnerTheme(&spinner, newTheme)

	// Verify the function executes without panic - actual style comparison is difficult
	// since lipgloss styles don't expose their foreground color directly
	if spinner.Spinner.FPS == 0 {
		t.Error("Spinner should still be configured after theme update")
	}
}

func TestRender(t *testing.T) {
	themeColors := theme.Themes[0]
	spinner := NewSpinner(themeColors)

	props := Props{
		Width:   80,
		Height:  24,
		Message: "Loading directory...",
		Spinner: spinner,
		Styles:  theme.NewStylesheet(themeColors),
	}

	result := Render(props)

	if result == "" {
		t.Error("Expected non-empty render result")
	}

	if !strings.Contains(result, "Loading directory...") {
		t.Error("Expected message to appear in render output")
	}
}

func TestRender_EmptyMessage(t *testing.T) {
	themeColors := theme.Themes[0]
	spinner := NewSpinner(themeColors)

	props := Props{
		Width:   80,
		Height:  24,
		Message: "",
		Spinner: spinner,
		Styles:  theme.NewStylesheet(themeColors),
	}

	result := Render(props)

	if result == "" {
		t.Error("Expected non-empty render result even with empty message")
	}
}

func TestSpinnerTick(t *testing.T) {
	themeColors := theme.Themes[0]
	spinner := NewSpinner(themeColors)

	cmd := spinner.Tick
	msg := cmd()

	if msg == nil {
		t.Error("Expected message from tick command")
	}
}
