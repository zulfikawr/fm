package main

import (
	"fmt"
	"os"

	"github.com/zulfikawr/fm/internal/bootstrap"
	"github.com/zulfikawr/fm/internal/cli"
	"github.com/zulfikawr/fm/internal/constants"
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

	if args.ShowVersion {
		fmt.Printf("fm version %s\n", constants.AppVersion)
		return nil
	}

	if args.IsSearch {
		return cli.RunSearch(args)
	}

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

	pOpts := []tea.ProgramOption{tea.WithAltScreen()}
	if a.Model.Config.EnableMouse {
		pOpts = append(pOpts, tea.WithMouseCellMotion())
	}

	p := tea.NewProgram(a, pOpts...)
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("running file manager: %w", err)
	}

	return nil
}
