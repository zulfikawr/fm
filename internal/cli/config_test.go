package cli

import (
	"flag"
	"io"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/config"
)

func TestParseConfig(t *testing.T) {
	t.Run("config subcommand", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"config"})
		if !args.IsConfig {
			t.Error("Expected IsConfig to be true")
		}
		if args.ConfigReset || args.ConfigInit {
			t.Error("Expected flags to be false")
		}
	})

	t.Run("config --reset", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"config", "--reset"})
		if !args.IsConfig {
			t.Error("Expected IsConfig to be true")
		}
		if !args.ConfigReset {
			t.Error("Expected ConfigReset to be true")
		}
	})

	t.Run("config init", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"config", "init"})
		if !args.IsConfig {
			t.Error("Expected IsConfig to be true")
		}
		if !args.ConfigInit {
			t.Error("Expected ConfigInit to be true")
		}
	})
}

func TestRunConfig(t *testing.T) {
	// Mock stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	_, w, _ := os.Pipe()
	os.Stdout = w

	t.Run("showConfig", func(t *testing.T) {
		args := &Args{IsConfig: true}
		err := RunConfig(args)
		if err != nil {
			t.Errorf("RunConfig failed: %v", err)
		}
	})

	t.Run("resetConfig", func(t *testing.T) {
		args := &Args{IsConfig: true, ConfigReset: true}
		err := RunConfig(args)
		if err != nil {
			t.Errorf("RunConfig reset failed: %v", err)
		}

		// Verify it actually reset (vim is default editor)
		cfg := config.Load()
		if cfg.EditorIndex != 0 {
			t.Errorf("Expected editor index 0, got %d", cfg.EditorIndex)
		}
	})
}

func TestConfigInitModel(t *testing.T) {
	m := ConfigInitModel{
		config: config.DefaultConfig(),
	}

	// Test navigation
	m.cursor = 0
	m.step = 0 // Theme selection

	// Test update "down"
	m, _ = updateModel(m, "down")
	if m.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", m.cursor)
	}

	// Test update "enter" to next step
	m, _ = updateModel(m, "enter")
	if m.step != 1 {
		t.Errorf("Expected step 1, got %d", m.step)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor reset to 0, got %d", m.cursor)
	}

	// Test through all steps
	m, _ = updateModel(m, "enter") // Icons
	if m.step != 2 {
		t.Errorf("Expected step 2, got %d", m.step)
	}

	m, _ = updateModel(m, "enter") // Mouse
	if m.step != 3 {
		t.Errorf("Expected step 3, got %d", m.step)
	}

	m, _ = updateModel(m, "enter") // Editor
	if m.step != 4 {
		t.Errorf("Expected step 4, got %d", m.step)
	}

	if m.quitting {
		t.Error("Should not be quitting before final enter")
	}

	m, _ = updateModel(m, "enter") // Save & Quit
	if !m.quitting {
		t.Error("Expected quitting to be true after final enter")
	}
}

// Helper to test model updates without tea.Msg boiler plate
func updateModel(m ConfigInitModel, key string) (ConfigInitModel, tea.Cmd) {
	var msg tea.KeyMsg
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "j":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	case "k":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	case "q":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	newModel, cmd := m.Update(msg)
	return newModel.(ConfigInitModel), cmd
}
