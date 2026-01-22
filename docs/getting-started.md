# Getting Started

This guide will help you get `fm` up and running on your system.

## Prerequisites

- **Go:** You need Go 1.25.5 or later installed on your system to build from source.
- **Terminal:** A terminal emulator that supports ANSI escape sequences (most modern terminals).

## Installation

### Building from Source

To build `fm` from source, clone the repository and use the `go build` command:

```bash
# Clone the repository
git clone https://github.com/zulfikawr/fm.git
cd fm

# Build the binary
go build -o fm ./cmd/fm

# Move it to your PATH (optional)
sudo mv fm /usr/local/bin/
```

## Basic Usage

### Launching `fm`

Simply run the command to open `fm` in your current directory:

```bash
fm
```

To open a specific directory:

```bash
fm ~/Documents
```

### Basic Navigation

Navigation in `fm` is designed to be intuitive:

- Use **Arrow Keys** or **h, j, k, l** (Vim-style) to move.
- **Enter** or **Right Arrow** to enter a directory.
- **Backspace** or **Left Arrow** to go to the parent directory.
- **Ctrl+C** to exit the application.

### Command Line Search

`fm` also includes a standalone search command that can be used directly from your shell without entering the TUI:

```bash
# Search for a query in the current directory
fm search "your search term"

# Search in a specific path
fm search "query" /path/to/search
```

## First Steps

1.  **Explore your files:** Use the navigation keys to move around.
2.  **Try the filter:** Press `/` and start typing to filter the current directory view.
3.  **Check settings:** Press `.` to view and toggle basic settings.
4.  **Open a file:** Press `Enter` on a file to open it in your default editor.
