package main

import (
	"fmt"
	"os"

	"github.com/zulfikawr/fm/internal/bootstrap"
	"github.com/zulfikawr/fm/internal/cli"
	"github.com/zulfikawr/fm/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := cli.Parse()

	a, err := bootstrap.InitializeApp(args.Remote, args.Args)
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
