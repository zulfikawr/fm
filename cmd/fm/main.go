// Package main is the entry point for the fm terminal file manager.
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

// run initializes and executes the file manager application.
// It handles CLI argument parsing, mode selection (TUI, search, info, analyze, config),
// and ensures proper cleanup on exit or panic.
func run() error {
	args, err := cli.Parse()
	if err != nil {
		return fmt.Errorf("parsing arguments: %w", err)
	}

	if args.ShowVersion {
		fmt.Printf("fm version %s\n", constants.AppVersion)
		return nil
	}

	if args.IsSearch {
		return cli.RunSearch(args)
	}

	if args.IsInfo {
		opts := cli.InfoOptions{
			Path:      ".",
			JSON:      args.InfoJSON,
			Tree:      args.InfoTree,
			TreeDepth: args.InfoDepth,
			Remote:    args.Remote,
		}
		if len(args.Args) > 0 {
			opts.Path = args.Args[0]
		}
		return cli.RunInfo(opts)
	}

	if args.IsAnalyze {
		return cli.RunAnalyze(args)
	}

	if args.IsConfig {
		return cli.RunConfig(args)
	}

	a, err := bootstrap.InitializeApp(args.Remote, args.Args)
	if err != nil {
		return err
	}
	a.Model.StartInAnalyzeMode = args.IsAnalyze
	defer tui.Close(a.Model)

	// Ensure cleanup happens even on panic
	defer func() {
		if r := recover(); r != nil {
			tui.Close(a.Model)
			panic(r) // Re-panic after cleanup
		}
	}()

	pOpts := []tea.ProgramOption{tea.WithAltScreen()}
	if a.Model.Config.UI.EnableMouse {
		pOpts = append(pOpts, tea.WithMouseCellMotion())
	}

	p := tea.NewProgram(a, pOpts...)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running file manager (final state: %+v): %w", finalModel, err)
	}

	return nil
}
