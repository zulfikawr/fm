# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.2] - 2026-01-09

### Added
- **Test Reorganization:** Standardized the test suite by implementing dedicated `_test.go` files for every source module, ensuring a 1:1 mapping between code and tests.
- **Coverage Milestone:** Significantly increased test coverage across all internal packages to exceed the **80%** threshold.

### Changed
- **Architectural Refactor:** Performed a comprehensive modularization of the codebase to improve maintainability and reduce technical debt.
- **Decomposed `internal/files`:** Split the primary file management logic into specialized modules: `item.go`, `ops.go`, `list.go`, `sort.go`, and `format.go`.
- **Restructured `internal/tui`:** Decomposed large TUI files into focused components:
    - Dedicated view modules for the file list, settings, and shared UI components.
    - Specialized update handlers for navigation, file operations, search, and settings.
    - Centralized `commands.go` for Bubble Tea messages and command factories.
- **Lean CLI Entry Point:** Refactored `cmd/fm/main.go` to remove TUI rendering logic, moving the help screen implementation to the `tui` package.
- **Test Suite Cleanup:** Removed large "god" test files (`tui_test.go`, `render_test.go`, `coverage_test.go`, and `files_test.go`) in favor of the new modular testing structure.

## [v0.1.1] - 2026-01-09

### Added
- **Configurable Formats:** Added settings to choose between multiple date/time formats (Default, ISO, US, Short) and file size units (Short, Full, Bytes).
- **Expanded Themes:** Added Monokai, Solarized Dark, Red, Tokyo Night, Rose Pine, and Catppuccin Mocha dark themes. **Gruvbox** is now the default theme.
- **Enhanced UI Coloring:** Colorized keybindings and path headers to be more prominent; dimmed secondary info like dates, sizes, and the "up dir" (..) entry for better focus.
- **Item Counter Fix:** Footer counter now correctly excludes the "up dir" from the total count and index.
- **Date Modified Column:** Added a new column showing file modification time in `DD/MM/YYYY 24H` format.
- **Display Settings:** Added settings to toggle the visibility of "Size", "Date Modified", and "Column Headers".
- **Responsive Truncation:** Filenames are now intelligently truncated with `…` based on available terminal width and active columns.
- **Rendering Tests:** Added comprehensive tests for row rendering and responsive truncation logic.
- **Improved Column Headers:** Added transparent headers with a separator line for better visual distinction from the breadcrumb.
- **Directory Sizes:** Replaced the `<DIR>` placeholder with actual recursive directory size calculations.
- **Unicode Navigation:** Replaced the standard `..` with a more descriptive `↑ ..` unicode icon.
- **Increase test coverage**

### Changed
- **Redesigned Settings TUI:** Reorganized settings into categorized groups (File Operations, Display Options, Appearance) with Title Case headers, improved spacing, and responsive layout. Added theme-aware styling for `[ON]` (primary color) and `[OFF]` (dimmed) states.
- **Adaptive Selection Mode:** Selection indicators (`[ ]` / `[x]`) and left padding are now hidden by default to provide a cleaner view.
- **Dynamic UI:** Selection mode is automatically activated when using `Space` and deactivated when clearing selections with `Esc`.
- **Improved Sorting:** All sorting modes (Name, Size, Date) now correctly mix files and folders together based on the selected criteria. The **Default Sort** remains "folders first" for logical navigation.
- **Git Integration:** Now respects `.gitignore` files. Ignored files and directories are automatically detected and displayed with a dimmed style in the TUI.
- **Codebase Cleanup:** Removed unused style declarations to maintain a lean TUI engine.
- **Settings UI:** Improved settings list with `< Value >` indicators for cyclable options like Themes.

## [v0.1.0] - 2026-01-09

### Added
- **Core Engine:** Modular TUI implementation using Bubble Tea.
- **File System:** Recursive copy, delete, and rename operations.
- **Bulk Selection:** Multi-file selection using `Space` for batch actions.
- **Git Integration:** Porcelain status markers and ghost entries for deleted files.
- **Settings System:** Full-screen settings menu with dynamic theme switching.
- **Persistence:** JSON-based configuration management.
- **Navigation Memory:** Directory-specific cursor and scroll position tracking.
- **CLI Interface:** Help flags and themed command-line output.
- **Scaling:** Asynchronous directory loading and background watchers.