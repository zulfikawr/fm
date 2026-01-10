package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"filemanager/internal/config"
	"filemanager/internal/files"
	"filemanager/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func main() {
	// Load config first to check for theme
	cfg := config.Load()
	theme := tui.Themes[cfg.ThemeIndex]
	styles := tui.NewStylesheet(theme)

	// Define flags
	var remote string
	flag.StringVar(&remote, "remote", "", "Remote address (user@host[:path])")
	flag.StringVar(&remote, "r", "", "Remote address (shorthand)")

	// Custom Usage
	flag.Usage = func() {
		tui.PrintHelp(styles, theme.Name)
	}

	flag.Parse()

	var fs files.FileSystem
	var startPath string
	var err error

	if remote != "" {
		// Parse remote string: user@host[:path]
		if !strings.Contains(remote, "@") {
			fmt.Println("Error: Invalid remote format. Use user@host[:path]")
			os.Exit(1)
		}

		parts := strings.SplitN(remote, "@", 2)
		user := parts[0]
		rest := parts[1]

		var host string
		if strings.Contains(rest, ":") {
			parts2 := strings.SplitN(rest, ":", 2)
			host = parts2[0]
			startPath = parts2[1]
		} else {
			host = rest
			startPath = "."
		}

		fmt.Printf("Connecting to %s@%s...\n", user, host)

		// Try connecting without password first (Agent/Key)
		fs, err = files.NewSftpFS(host, user, "", "")
		if err != nil {
			// If failed, prompt for password
			fmt.Print("Password: ")
			bytePw, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println() // Newline after input
			if err != nil {
				fmt.Printf("Error reading password: %v\n", err)
				os.Exit(1)
			}

			fs, err = files.NewSftpFS(host, user, string(bytePw), "")
			if err != nil {
				fmt.Printf("Connection failed: %v\n", err)
				os.Exit(1)
			}
		}

		if startPath == "." || startPath == "" {
			startPath, _ = fs.GetHomeDir()
		}

	} else {
		// Local File System
		fs = &files.LocalFS{}

		// Determine start path from arguments (non-flag)
		args := flag.Args()
		if len(args) > 0 {
			argPath := args[0]

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
			startPath = absPath
		} else {
			startPath, err = os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
		}
	}

	m := tui.NewModel(fs, startPath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	defer m.Close()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting file manager: %v\n", err)
		os.Exit(1)
	}
}
