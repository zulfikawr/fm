# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.3] - 2026-01-10

### Added
- **Remote Filesystems:** Added native support for browsing and managing files on remote servers via SSH/SFTP. Use `fm --remote user@host` (or `-r`) to connect. Includes SSH agent, key, and password authentication.
- **Production-Ready Logging:** Introduced a thread-safe `internal/logger` package with INFO/ERROR levels and cross-platform path resolution using `os.UserConfigDir()`.
- **SSH Host Key Verification:** Implemented secure host key verification using the standard `~/.ssh/known_hosts` file for SFTP connections.
- **Async File Operations:** Refactored Delete and Paste operations to run asynchronously, preventing TUI freezes during large file transfers.
- **Resource Management:** Added a `Close()` method to the `FileSystem` interface and `Model` for clean release of SSH connections and filesystem watchers.
- **Open File Feature:** Integrated the ability to open files directly from the TUI. Code and text files open in a user-preferred editor, while others use the system default application.
- **Editor Selection Setting:** Added a new configuration option to choose the preferred text editor (supporting vim, nano, vi, emacs, code, subl, cursor, and zed).
- **Trash Bin Integration:** Modified deletion logic to move files to the system trash (using `gio` on Linux, `osascript` on macOS, and PowerShell on Windows) instead of permanent deletion.
- **Asynchronous Background Caching:** Reimplemented directory size calculations and Git status fetching to run in the background, ensuring the TUI remains responsive when browsing large repositories.
- **Directory Size Caching**: Added a cache for directory sizes. Now, when navigate back to a directory or re-open one, the size is displayed instantly if it was previously calculated.
- **SSH/Large Directory Performance**: Using size cache to significantly reduced the number of recursive filesystem walks performed.
- **UI Overflow Fix**: Reduced the file list height by an additional 3 lines when the header is enabled.

### Fixed
- **Config Persistence:** Improved configuration loading with error reporting to handle and debug potential JSON parsing failures.
- **Windows Git Refresh Loop:** Improved `.git` directory filtering to handle Windows backslashes, preventing infinite refresh loops on Windows systems.
- **SFTP Git Branching:** Enabled Git branch detection and status markers for remote SFTP connections.
- **Cursor Stability:** Fixed an issue where the cursor would jump to the top during background reloads of the same directory.
- **Git Refresh Loop Fix**: Modified the file watcher to ignore events originating from within the .git directory.

## [v0.1.2] - 2026-01-09

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
...