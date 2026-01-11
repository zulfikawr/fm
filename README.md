# FM - Terminal File Manager

![Screenshot](./image.png)

A fast, modular, and feature-rich TUI file manager written in Go.

[![GitHub Release](https://img.shields.io/github/v/release/zulfikawr/fm)](https://github.com/zulfikawr/fm/releases)
[![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **🚀 Performance:** Parallel directory counting, disk-backed caching, and real-time watching.
- **📁 File Operations:** Full CRUD with cut/paste, interactive conflict resolution, and trash support.
- **🔒 Security:** SSH host key verification and read-only filesystem protection.
- **☁️ Remote Access:** Connect to remote servers via SSH/SFTP with password or key auth.
- **🌿 Git Integration:** Visual status markers (`M`, `A`, `D`, `?`) and current branch display.
- **🎨 Theme System:** Multiple color schemes including Nord, Dracula, and Gruvbox.
- **🔍 Search & Filter:** Real-time fuzzy-style filtering within directories.
- **⚙️ Persistent Config:** Settings saved automatically to `~/.config/fm/config.json`.
- **💾 Memory:** Remembers your cursor and scroll position for every directory visited.
- **📑 Tabs:** Multitasking support with up to 9 active directory tabs.

## Keybindings

| Key | Action |
| --- | --- |
| `Enter` / `→` / `l` | Open directory / Open file in editor |
| `Backspace` / `←` / `h` | Navigate to parent directory |
| `Space` | Toggle selection for bulk actions |
| `Alt+T` | New tab (max 9) |
| `Alt+1`-`9` | Switch between tabs |
| `Alt+W` | Close current tab |
| `/` | Enter search/filter mode |
| `s` | Cycle sort modes (Name, Size, Date) |
| `c` | Copy selection to clipboard |
| `x` | Cut selection to clipboard |
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

Settings are stored in `~/.config/fm/config.json`. Key options include:

- **General:** `show_hidden`, `case_sensitive`, `wrap_navigation`
- **Display:** `show_size`, `show_date_modified`, `show_header`
- **Behavior:** `confirm_operations`, `enable_git`, `use_trash`
- **Preferences:** `theme_index`, `editor_index` (vim, nano, code, etc.)
- **Formats:** `date_format_index`, `size_format_index`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
