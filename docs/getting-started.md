# Getting Started

This guide will help you get `fm` up and running on your system.

## Prerequisites

- **Go:** You need Go 1.25.5 or later installed on your system to build from source.
- **Terminal:** A terminal emulator that supports ANSI escape sequences (most modern terminals).

## Installation

### Automatic Install (Linux & macOS)

The easiest way to install `fm` is using the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/zulfikawr/fm/main/install.sh | bash
```

### Building from Source

To build `fm` from source, clone the repository and use the `make build` command:

```bash
# Clone the repository
git clone https://github.com/zulfikawr/fm.git
cd fm

# Build the binary
make build

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

### Command Line Tools

`fm` provides powerful CLI commands for quick operations:

#### Search
Search for files and content without entering the TUI:

```bash
# Search in current directory
fm search "your search term"

# Search in specific path
fm search "query" /path/to/search

# Search with regex patterns
fm search "func.*main" ./src
```

#### Info
Display detailed information about files and directories:

```bash
# Show info for current directory
fm info .

# Show info for specific file
fm info README.md

# JSON output for scripting
fm info --json .

# Tree view
fm info --tree --depth 2 ./src

# Remote file info
fm info -r user@server:/path
```

See the [**CLI Reference**](./cli-reference.md) for complete documentation.

### Basic Navigation

Navigation in `fm` is designed to be intuitive:

- Use **Arrow Keys** or **h, j, k, l** (Vim-style) to move.
- **Enter** or **Right Arrow** to enter a directory.
- **Backspace** or **Left Arrow** to go to the parent directory.
- **Ctrl+C** to exit the application.

## First Steps

1.  **Explore your files:** Use the navigation keys to move around.
2.  **Try the filter:** Press `/` and start typing to filter the current directory view.
3.  **Check settings:** Press `,` to view and toggle basic settings.
4.  **Open a file:** Press `Enter` on a file to open it in your default editor.
