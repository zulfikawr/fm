# FM - Terminal File Manager

![Screenshot](./image.png)

A fast, modular, and feature-rich TUI file manager written in Go.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Version](https://img.shields.io/badge/version-0.1.3-green.svg)

## Features

- **🚀 Performance:** Asynchronous directory loading and real-time file system watching.
- **📁 File Operations:** Full CRUD support (Copy, Paste, Rename, Delete) with bulk selection.
- **☁️ Remote Access:** Connect to remote servers via SSH/SFTP with password or key auth.
- **🌿 Git Integration:** Visual status markers (`M`, `A`, `D`, `?`) and current branch display.
- **🎨 Theme System:** Multiple color schemes including Nord, Dracula, and Gruvbox.
- **🔍 Search & Filter:** Real-time fuzzy-style filtering within directories.
- **⚙️ Persistent Config:** Settings saved automatically to `~/.config/fm/config.json`.
- **💾 Memory:** Remembers your cursor and scroll position for every directory visited.

## Keybindings

| Key | Action |
| --- | --- |
| `Enter` / `→` / `l` | Open directory / Open file in editor |
| `Backspace` / `←` / `h` | Navigate to parent directory |
| `Space` | Toggle selection for bulk actions |
| `/` | Enter search/filter mode |
| `s` | Cycle sort modes (Name, Size, Date) |
| `c` | Copy selection to clipboard |
| `v` | Paste clipboard contents |
| `r` | Rename highlighted item |
| `d` | Trash selection (with confirmation) |
| `.` | Open Settings & Themes |
| `Esc` | Unselect all / Clear message / Close Settings |
| `q` | Quit |

## Installation

```bash
# Build from source
go build -o fm ./cmd/fm

# Run
./fm [path]
```

## Remote Access (SSH/SFTP)

Connect directly to remote servers using the `--remote` (or `-r`) flag.

```bash
# Connect to a remote host (prompts for password if needed)
fm --remote user@192.168.1.50

# Connect to a specific path
fm -r user@example.com:/var/www

# Uses your local SSH agent and keys automatically.
```

## Configuration

Settings are stored in `~/.config/fm/config.json`. You can toggle features like:
- Hidden file visibility
- Case-sensitive search
- Operation confirmations
- Git status integration
- Navigation wrapping
- Default text editor
- Trash usage
- Themes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
