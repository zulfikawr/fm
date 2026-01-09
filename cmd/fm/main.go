package main

import (
	"fmt"
	"os"
	"path/filepath"

	"filemanager/internal/config"
	"filemanager/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load config first to check for theme if help is requested
	cfg := config.Load()
	theme := tui.Themes[cfg.ThemeIndex]
	styles := tui.NewStylesheet(theme)

	// Check for help flags
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" {
			tui.PrintHelp(styles, theme.Name)
			os.Exit(0)
		}
	}

	// Default to current working directory
	path, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// If an argument is provided, use it as the starting path
	if len(os.Args) > 1 {
		argPath := os.Args[1]
		absPath, err := filepath.Abs(argPath)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			os.Exit(1)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Printf("Error accessing path: %v\n", err)
			os.Exit(1)
		}

		if !info.IsDir() {
			fmt.Printf("Error: %s is not a directory\n", argPath)
			os.Exit(1)
		}
		path = absPath
	}

	p := tea.NewProgram(tui.NewModel(path), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting file manager: %v\n", err)
		os.Exit(1)
	}
}
