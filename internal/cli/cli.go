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
	var isInfo bool
	var infoJSON bool
	var infoTree bool
	var infoDepth int
	var isAnalyze bool
	var isConfig bool
	var configReset bool
	var configInit bool

	f.StringVar(&remoteStr, "remote", "", "Remote address (user@host[:path] or ssh-alias)")
	f.StringVar(&remoteStr, "r", "", "Remote address (shorthand)")
	f.BoolVar(&showVersion, "version", false, "Show version information")
	f.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	f.StringVar(&searchStr, "search", "", "Perform fuzzy search for files and content")
	f.StringVar(&searchStr, "s", "", "Perform fuzzy search (shorthand)")
	f.BoolVar(&isRegex, "regex", false, "Use regular expressions for search")
	f.BoolVar(&isRegex, "e", false, "Use regular expressions (shorthand)")
	f.BoolVar(&isInfo, "info", false, "Display file/directory information")
	f.BoolVar(&isInfo, "i", false, "Display file/directory information (shorthand)")
	f.BoolVar(&infoJSON, "json", false, "Output in JSON format (for info)")
	f.BoolVar(&infoTree, "tree", false, "Display directory tree (for info)")
	f.IntVar(&infoDepth, "depth", 2, "Tree depth (for info, 0 for unlimited)")
	f.BoolVar(&isAnalyze, "analyze", false, "Analyze disk usage")
	f.BoolVar(&isAnalyze, "a", false, "Analyze disk usage (shorthand)")
	f.BoolVar(&isConfig, "config", false, "Manage configuration")
	f.BoolVar(&isConfig, "c", false, "Manage configuration (shorthand)")
	f.BoolVar(&configReset, "reset", false, "Reset configuration to default (for config)")
	f.BoolVar(&configInit, "init", false, "Run interactive configuration wizard (for config)")

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
	remainingArgs := f.Args()

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
