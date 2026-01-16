package cli

import (
	"fmt"

	"fm/internal/tui/theme"
)

// PrintHelp displays the help information to the console
func PrintHelp(styles theme.Stylesheet, themeName string) {
	fmt.Println("FM - Terminal File Manager")
	fmt.Printf("Active Theme: %s\n\n", themeName)
	fmt.Println("Usage:")
	fmt.Println("  fm [path]                Open fm in the specified directory")
	fmt.Println("  fm -r user@host[:path]   Open fm on a remote server via SFTP")
	fmt.Println("\nKeybindings:")
	fmt.Println("  j/down, k/up   Move cursor")
	fmt.Println("  l/enter        Enter directory or open file")
	fmt.Println("  h/backspace    Go to parent directory")
	fmt.Println("  [ / ]          History Back / Forward")
	fmt.Println("  Space          Toggle selection")
	fmt.Println("  alt+a          Select all")
	fmt.Println("  alt+t          Create new tab")
	fmt.Println("  alt+1-9        Switch to tab 1-9")
	fmt.Println("  alt+w          Close current tab")
	fmt.Println("  alt+l          Toggle operation logs")
	fmt.Println("  alt+c          Toggle clipboard view")
	fmt.Println("  alt+/          Fuzzy content search")
	fmt.Println("  /              Filter current directory")
	fmt.Println("  g              Go to path (local/remote)")
	fmt.Println("  c/y            Copy selected items")
	fmt.Println("  x              Cut selected items")
	fmt.Println("  v              Paste items from clipboard")
	fmt.Println("  d              Delete selected items")
	fmt.Println("  r              Rename selected item")
	fmt.Println("  z              Zip selected items")
	fmt.Println("  u              Unzip selected item")
	fmt.Println("  .              Toggle settings")
	fmt.Println("  Esc            Back / Clear selection")
	fmt.Println("  ctrl+c         Quit")
}
