# FM - Terminal File Manager

![Screenshot](./image.png)

A fast, modular, and feature-rich TUI file manager written in Go. 

[![Website](https://img.shields.io/badge/Website-fm.zulfikar.site-blue?style=flat-square)](https://fm.zulfikar.site/)
[![GitHub Release](https://img.shields.io/github/v/release/zulfikawr/fm)](https://github.com/zulfikawr/fm/releases)
[![Go Version](https://img.shields.io/badge/Go-1.25.5-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Documentation

For detailed guides, configuration options, and advanced usage, please see the [**docs/**](./docs/index.md) directory:

- [**Getting Started**](./docs/getting-started.md): Installation and basic usage.
- [**Mouse Support**](./docs/mouse-support.md): Full guide to clicking, dragging, and scrolling.
- [**Keybindings**](./docs/keybindings.md): Comprehensive list of shortcuts.
- [**Features**](./docs/features.md): Deep dive into Git, Search, and Archive management.
- [**Remote Access**](./docs/remote-access.md): Managing files over SSH/SFTP.
- [**Configuration**](./docs/configuration.md): Customizing `fm` to your needs.

## 🚀 Quick Start

```bash
# Build from source
go build -o fm ./cmd/fm

# Run
./fm [path]
```

## ⌨️ Quick Keybindings

| Key | Action |
| --- | --- |
| `Enter` / `→` / `l` | Open directory / Open file in editor |
| `Backspace` / `←` / `h` | Navigate to parent directory |
| `Space` | Toggle selection for bulk actions |
| `c` | Copy selection to clipboard |
| `x` | Cut selection to clipboard |
| `v` | Paste clipboard contents |
| `r` | Rename highlighted item |
| `/` | Enter filter mode |
| `Alt+/` | Fuzzy content search (Find in Files) |
| `g` | Go to path (Local or Remote) |
| `Alt+T` / `Alt+W` | New Tab / Close Tab |
| `Alt+1`-`9` | Switch between tabs |
| `.` | Toggle settings |
| `Ctrl+C` | Quit |

See [**keybindings.md**](./docs/keybindings.md) for the full list.

## ✨ Core Features

- **Performance:** Fast navigation with a modular concurrent architecture.
- **Mouse Support:** Modern interaction system with scrolling, drag-to-select, and clickable UI elements.
- **Git Integration:** Real-time status markers and branch information.
- **Remote Access:** Full SFTP support for managing remote servers.
- **Fuzzy Search:** Deep content search powered by a concurrent engine.
- **Tabs:** Multitasking with up to 9 active directory tabs.
- **Archive Support:** Create and extract Zip/Tar archives directly in the UI.

## 🛠️ Technology Stack

- **Go:** Core logic and high-performance concurrency.
- **Bubble Tea:** The TUI framework for building terminal applications.
- **Lip Gloss:** For styling and layouts.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
