package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"fm/internal/config"
	"fm/internal/files"
	"fm/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
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
			return fmt.Errorf("invalid remote format, use user@host[:path]")
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
				return fmt.Errorf("reading password: %w", err)
			}

			// Zero password bytes after use
			password := string(bytePw)
			for i := range bytePw {
				bytePw[i] = 0
			}
			bytePw = nil

			fs, err = files.NewSftpFS(host, user, password, "")
			// Clear password string from memory
			password = ""

			if err != nil {
				return fmt.Errorf("connection failed: %w", err)
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
				return fmt.Errorf("resolving path: %w", err)
			}

			info, err := os.Stat(absPath)
			if err != nil {
				return fmt.Errorf("accessing path: %w", err)
			}

			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", argPath)
			}
			startPath = absPath
		} else {
			startPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
		}
	}

	m := tui.NewModel(fs, startPath)
	defer m.Close()

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("running file manager: %w", err)
	}

	return nil
}
