package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"fm/internal/config"
	"fm/internal/files"
	"fm/internal/files/local"
	remotefs "fm/internal/files/remote"
	"fm/internal/sshutil"
	"fm/internal/tui"
	"fm/internal/tui/help"
	"fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	// Ensure directory exists
	_ = os.MkdirAll(sshDir, 0700)
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		_ = os.WriteFile(knownHostsPath, []byte{}, 0600)
	}

	cb, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			// Host not found in known_hosts
			fmt.Printf("The authenticity of host '%s' can't be established.\n", hostname)
			fmt.Printf("%s key fingerprint is %s.\n", key.Type(), ssh.FingerprintSHA256(key))
			fmt.Print("Are you sure you want to continue connecting (y/n)? ")

			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				// Add to known_hosts
				return sshutil.AddToKnownHosts(hostname, remote, key)
			}
		}
		return err
	}, nil
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

	var fs files.FileSystem
	var startPath string
	var err error

	if remoteStr != "" {
		host := remoteStr
		user := ""
		keyPath := ""
		password := ""

		// Check SSH config first
		sshConfigs, _ := sshutil.ParseSSHConfig()
		if cfg, ok := sshConfigs[remoteStr]; ok {
			host = cfg.HostName
			if host == "" {
				host = remoteStr
			}
			user = cfg.User
			keyPath = cfg.IdentityFile
		} else if strings.Contains(remoteStr, "@") {
			parts := strings.SplitN(remoteStr, "@", 2)
			user = parts[0]
			rest := parts[1]

			if strings.Contains(rest, ":") {
				parts2 := strings.SplitN(rest, ":", 2)
				host = parts2[0]
				startPath = parts2[1]
			} else {
				host = rest
			}
		}

		fmt.Printf("Connecting to %s@%s...\n", user, host)

		// Create CLI host key callback (blocking)
		hkcb, err := createHostKeyCallback()
		if err != nil {
			return fmt.Errorf("failed to setup host key verification: %w", err)
		}

		// Try connecting with provided key or agent first
		fs, err = remotefs.NewSftpFS(host, user, "", keyPath, hkcb)
		if err != nil {
			// Check if it's a host key verification failure (user said no or mismatch)
			var keyErr *knownhosts.KeyError
			if errors.As(err, &keyErr) {
				return fmt.Errorf("host key verification failed")
			}

			// If failed, prompt for password
			fmt.Printf("Connection attempt failed: %v\n", err)
			fmt.Print("Password: ")
			bytePw, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println() // Newline after input
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}

			password = string(bytePw)
			for i := range bytePw {
				bytePw[i] = 0
			}

			fs, err = remotefs.NewSftpFS(host, user, password, keyPath, hkcb)
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
		fs = &local.LocalFS{}

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

	a := tui.NewApp(fs, startPath)
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
