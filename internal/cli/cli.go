package cli

import (
	"flag"
	"os"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Args contains the parsed command line arguments
type Args struct {
	Remote      string
	ShowVersion bool
	IsSearch    bool
	SearchQuery string
	Args        []string
}

// Parse handles command line flag parsing
func Parse() *Args {
	return parse(flag.CommandLine, os.Args[1:])
}

func parse(f *flag.FlagSet, args []string) *Args {
	var remoteStr string
	var showVersion bool
	var searchStr string
	f.StringVar(&remoteStr, "remote", "", "Remote address (user@host[:path] or ssh-alias)")
	f.StringVar(&remoteStr, "r", "", "Remote address (shorthand)")
	f.BoolVar(&showVersion, "version", false, "Show version information")
	f.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	f.StringVar(&searchStr, "search", "", "Perform fuzzy search for files and content")
	f.StringVar(&searchStr, "s", "", "Perform fuzzy search (shorthand)")

	// Custom Usage
	f.Usage = func() {
		cfg := config.Load()
		t := theme.Themes[cfg.ThemeIndex]
		styles := theme.NewStylesheet(t)
		PrintHelp(styles, t.Name)
	}

	_ = f.Parse(args)

	isSearch := searchStr != ""
	remainingArgs := f.Args()

	// Handle "search" as a subcommand
	if !isSearch && len(remainingArgs) > 0 && remainingArgs[0] == "search" {
		isSearch = true
		if len(remainingArgs) > 1 {
			searchStr = remainingArgs[1]
			remainingArgs = remainingArgs[2:]
		} else {
			remainingArgs = remainingArgs[1:]
		}
	}

	return &Args{
		Remote:      remoteStr,
		ShowVersion: showVersion,
		IsSearch:    isSearch,
		SearchQuery: searchStr,
		Args:        remainingArgs,
	}
}
