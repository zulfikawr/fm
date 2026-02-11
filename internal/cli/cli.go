package cli

import (
	"flag"
	"os"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Args contains the parsed command line arguments
type Args struct {
	// ... (omitting fields for brevity in instruction, but keeping them in new_string)
	Remote      string
	ShowVersion bool
	IsSearch    bool
	SearchQuery string
	IsRegex     bool
	IsInfo      bool
	InfoJSON    bool
	InfoTree    bool
	InfoDepth   int
	IsAnalyze   bool
	IsConfig    bool
	ConfigReset bool
	ConfigInit  bool
	Args        []string
}

// Parse handles command line flag parsing
func Parse() (*Args, error) {
	return parse(flag.CommandLine, os.Args[1:])
}

func parse(f *flag.FlagSet, args []string) (*Args, error) {
	var remoteStr string
	var showVersion bool
	var searchStr string
	var isRegex bool
	var infoJSON bool
	var infoTree bool
	var infoDepth int
	f.StringVar(&remoteStr, "remote", "", "Remote address (user@host[:path] or ssh-alias)")
	f.StringVar(&remoteStr, "r", "", "Remote address (shorthand)")
	f.BoolVar(&showVersion, "version", false, "Show version information")
	f.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	f.StringVar(&searchStr, "search", "", "Perform fuzzy search for files and content")
	f.StringVar(&searchStr, "s", "", "Perform fuzzy search (shorthand)")
	f.BoolVar(&isRegex, "regex", false, "Use regular expressions for search")
	f.BoolVar(&isRegex, "e", false, "Use regular expressions (shorthand)")

	// Custom Usage
	f.Usage = func() {
		cfg := config.Load()
		t := theme.Themes[cfg.UI.ThemeIndex]
		styles := theme.NewStylesheet(t)
		PrintHelp(styles, t.Name)
	}

	if err := f.Parse(args); err != nil {
		return nil, err
	}

	isSearch := searchStr != ""
	isInfo := false
	isAnalyze := false
	isConfig := false
	configReset := false
	configInit := false
	remainingArgs := f.Args()

	// Handle "search" as a subcommand
	if !isSearch && len(remainingArgs) > 0 && remainingArgs[0] == "search" {
		isSearch = true
		remainingArgs = remainingArgs[1:]

		// Parse search-specific flags
		searchFlags := flag.NewFlagSet("search", flag.ContinueOnError)
		searchFlags.BoolVar(&isRegex, "regex", false, "Use regular expressions")
		searchFlags.BoolVar(&isRegex, "e", false, "Use regular expressions")

		if err := searchFlags.Parse(remainingArgs); err != nil {
			return nil, err
		}
		remainingArgs = searchFlags.Args()

		if len(remainingArgs) > 0 {
			searchStr = remainingArgs[0]
			remainingArgs = remainingArgs[1:]
		}
	}

	// Handle "info" as a subcommand with its own flags
	if !isSearch && !isInfo && len(remainingArgs) > 0 && remainingArgs[0] == "info" {
		isInfo = true
		remainingArgs = remainingArgs[1:]

		// Parse info-specific flags
		infoFlags := flag.NewFlagSet("info", flag.ContinueOnError)
		infoFlags.BoolVar(&infoJSON, "json", false, "Output in JSON format")
		infoFlags.BoolVar(&infoTree, "tree", false, "Display directory tree")
		infoFlags.IntVar(&infoDepth, "depth", 2, "Tree depth (0 for unlimited)")

		if err := infoFlags.Parse(remainingArgs); err != nil {
			return nil, err
		}
		remainingArgs = infoFlags.Args()
	}

	if !isSearch && !isInfo && len(remainingArgs) > 0 && remainingArgs[0] == "analyze" {
		isAnalyze = true
		remainingArgs = remainingArgs[1:]
	}

	// Handle "config" as a subcommand
	if !isSearch && !isInfo && !isAnalyze && len(remainingArgs) > 0 && remainingArgs[0] == "config" {
		isConfig = true
		remainingArgs = remainingArgs[1:]

		if len(remainingArgs) > 0 && remainingArgs[0] == "init" {
			configInit = true
			remainingArgs = remainingArgs[1:]
		} else {
			// Parse config-specific flags
			configFlags := flag.NewFlagSet("config", flag.ContinueOnError)
			configFlags.BoolVar(&configReset, "reset", false, "Reset configuration to default")

			if err := configFlags.Parse(remainingArgs); err != nil {
				return nil, err
			}
			remainingArgs = configFlags.Args()
		}
	}

	return &Args{
		Remote:      remoteStr,
		ShowVersion: showVersion,
		IsSearch:    isSearch,
		SearchQuery: searchStr,
		IsRegex:     isRegex,
		IsInfo:      isInfo,
		InfoJSON:    infoJSON,
		InfoTree:    infoTree,
		InfoDepth:   infoDepth,
		IsAnalyze:   isAnalyze,
		IsConfig:    isConfig,
		ConfigReset: configReset,
		ConfigInit:  configInit,
		Args:        remainingArgs,
	}, nil
}
