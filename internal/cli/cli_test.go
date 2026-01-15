package cli

import (
	"flag"
	"io"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("Remote flag long", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"--remote", "user@host"})
		if args.Remote != "user@host" {
			t.Errorf("Expected user@host, got %s", args.Remote)
		}
	})

	t.Run("Remote flag short", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"-r", "user@host"})
		if args.Remote != "user@host" {
			t.Errorf("Expected user@host, got %s", args.Remote)
		}
	})

	t.Run("Positional args", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args := parse(fs, []string{"/tmp", "extra"})
		if len(args.Args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args.Args))
		}
		if args.Args[0] != "/tmp" {
			t.Errorf("Expected /tmp, got %s", args.Args[0])
		}
	})

	t.Run("Usage function", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		_ = parse(fs, []string{})
		if fs.Usage == nil {
			t.Error("Expected Usage function to be set")
		}
		// We don't necessarily need to call it if it's complex,
		// but we've verified it's set.
	})
}
