package main

import (
	"flag"
	"fmt"
	"os"

	"fm/internal/bootstrap"
	"fm/internal/config"
	"fm/internal/tui"
	"fm/internal/tui/help"
	"fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load config first to check for theme
	cfg := config.Load()
	t := theme.Themes[cfg.ThemeIndex]
	styles := theme.NewStylesheet(t)

	// Define flags
	var remoteStr string
	flag.StringVar(&remoteStr, "remote", "", "Remote address (user@host[:path] or ssh-alias)")
	flag.StringVar(&remoteStr, "r", "", "Remote address (shorthand)")

	// Custom Usage
	flag.Usage = func() {
		help.Print(styles, t.Name)
	}

	flag.Parse()

	a, err := bootstrap.InitializeApp(remoteStr, flag.Args())
	if err != nil {
		return err
	}
	defer tui.Close(a.Model)

	// Ensure cleanup happens even on panic
	defer func() {
		if r := recover(); r != nil {
			tui.Close(a.Model)
			panic(r) // Re-panic after cleanup
		}
	}()

	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("running file manager: %w", err)
	}

	return nil
}
