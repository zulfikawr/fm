package main

import (
	"fmt"
	"os"
	"path/filepath"

	"filemanager/internal/config"
	"filemanager/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			printHelp(styles, theme.Name)
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

func printHelp(s tui.Stylesheet, themeName string) {
	title := s.Header.Render(" FM - Terminal File Manager ")

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	usage := lipgloss.JoinVertical(lipgloss.Left,
		s.DirCol.Render("\nUsage:"),
		fmt.Sprintf("  fm [path]      %s", dim.Render("Open fm in directory")),
		fmt.Sprintf("  fm -h, --help  %s", dim.Render("Show this help screen")),
	)

	keys := lipgloss.JoinVertical(lipgloss.Left,
		s.DirCol.Render("\nKeybindings:"),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Enter/→/l"), dim.Render("Open directory")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Backspace/←/h"), dim.Render("Parent directory")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Space"), dim.Render("Select file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("s"), dim.Render("Cycle sort mode")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("/"), dim.Render("Search")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("."), dim.Render("Settings & Themes")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("c"), dim.Render("Copy selection")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("v"), dim.Render("Paste clipboard")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("r"), dim.Render("Rename file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("d"), dim.Render("Delete file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Esc"), dim.Render("Unselect / Clear")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("q/Ctrl+C"), dim.Render("Quit")),
	)

	fmt.Println(title)
	fmt.Println(usage)
	fmt.Println(keys)
}
