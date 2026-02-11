package cli

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/config"
)

func TestMain(m *testing.M) {
	// Isolate config for all tests in this package
	tempDir, err := os.MkdirTemp("", "fm-cli-test-*")
	if err != nil {
		panic(err)
	}

	config.SetConfigPath(filepath.Join(tempDir, "config.json"))

	code := m.Run()

	// Clean up
	if err := os.RemoveAll(tempDir); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestParse(t *testing.T) {
	t.Run("Remote flag long", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args, err := parse(fs, []string{"--remote", "user@host"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if args.Remote != "user@host" {
			t.Errorf("Expected user@host, got %s", args.Remote)
		}
	})

	t.Run("Remote flag short", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args, err := parse(fs, []string{"-r", "user@host"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if args.Remote != "user@host" {
			t.Errorf("Expected user@host, got %s", args.Remote)
		}
	})

	t.Run("Positional args", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		args, err := parse(fs, []string{"/tmp", "extra"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
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
		args, err := parse(fs, []string{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if args == nil {
			t.Error("Expected non-nil args")
		}
		if fs.Usage == nil {
			t.Error("Expected Usage function to be set")
		}
	})
}
