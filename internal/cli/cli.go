package cli

import (
	"flag"
	"os"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// Args contains the parsed command line arguments
type Args struct {
	Remote string
	Args   []string
}

// Parse handles command line flag parsing
func Parse() *Args {
	return parse(flag.CommandLine, os.Args[1:])
}

func parse(f *flag.FlagSet, args []string) *Args {
	var remoteStr string
	f.StringVar(&remoteStr, "remote", "", "Remote address (user@host[:path] or ssh-alias)")
	f.StringVar(&remoteStr, "r", "", "Remote address (shorthand)")

	// Custom Usage
	f.Usage = func() {
		cfg := config.Load()
		t := theme.Themes[cfg.ThemeIndex]
		styles := theme.NewStylesheet(t)
		PrintHelp(styles, t.Name)
	}

	_ = f.Parse(args)

	return &Args{
		Remote: remoteStr,
		Args:   f.Args(),
	}
}
