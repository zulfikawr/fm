# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.5] - 2026-01-12

### Added
- **Go to Path ('g'):** New navigation feature with "Smart Path Detection" (poly-mode):
  - **Local Jump:** Instantly navigate to local paths starting with `/`, `./`, `~/`, or `C:\`.
  - **Remote Connect:** Connect to remote servers using `user@host` syntax directly from the footer.
  - **SSH Config Support:** Support for SSH aliases defined in `~/.ssh/config`. Typing a host alias (e.g., `myserver`) automatically uses the configured `HostName`, `User`, and `IdentityFile`.
- **Interactive Remote Auth:** Improved remote connection flow:
  - Automatically attempts connection with SSH agent or default keys.
  - If initial connection fails, prompts for "Password or PEM path" directly in the footer.
  - Smart input detection: If the provided string is an existing file path, it's used as a PEM key; otherwise, it's treated as a password.
- **Background Host Verification:** Host key confirmation (`known_hosts`) is now handled asynchronously, preventing the TUI from blocking while waiting for user input.

### Performance
- **Pre-calculated Metadata Strings:** Formatted file size and modification dates are now computed once during directory load instead of every frame, significantly reducing CPU usage in large directories.
- **Efficient Selection Tracking:** Replaced $O(N)$ selection counting with a cached $O(1)$ counter in the state model.
- **Fast-Path Selection Restore:** Optimized directory reloading to skip selection state reconciliation when no items are selected.
- **ViewState Pointer Optimization:** Refactored UI state passing to use pointers, reducing memory allocations and stack copying overhead during the render loop.
- **Memoized Footer Prompts:** Confirmation prompts are now pre-calculated and cached, eliminating redundant string parsing and ANSI colorization cycles.
- **Stylesheet Memoization:** The theme stylesheet is now fetched once per render cycle and shared across all sub-components.
- **Optimized List Rendering:** Introduced `ListLayout` to pre-calculate column widths once per frame, removing redundant calculations from individual row rendering.
- **Git Status Lookups:** Replaced complex branching logic in row rendering with $O(1)$ map-based style lookups for Git markers.

### Refactor
- **Modularization:** Performed a major restructuring of the codebase, moving from flat files to specialized individual packages to improve maintainability and separation of concerns:
  - Decomposed `internal/files` into sub-packages: `errors`, `format`, `listing`, `local`, `ops`, `remote`, and `sorting`.
  - Extracted Git logic into a dedicated `internal/git` package.
  - Refactored `internal/tui` into a more granular structure with sub-packages for `actions`, `cache`, `commands`, `components`, `errors`, `filter`, `help`, `state`, `theme`, `update`, and `view`.
- **SSH/SFTP Enhancements:**
  - **Interactive Host Key Confirmation:** Added a security prompt to verify and trust unknown remote hosts during connection, with automatic persistence to `~/.ssh/known_hosts`.
  - **Private Key Support:** Added support for identity files (`.pem`, etc.) via a new positional argument: `fm -r user@host /path/to/key`.
  - **Connection Resilience:** Improved SFTP setup to automatically create missing `.ssh` directories and `known_hosts` files with secure permissions.
- **Test Infrastructure:** Comprehensive migration to a centralized `testutil` package.
  - Replaced manual `os.MkdirTemp` and `os.WriteFile` calls with `testutil.TempDir` and `testutil.CreateTestFile` for consistent cleanup.
  - Implemented a delegating `MockFileSystem` in `testutil` to allow partial mocking.
  - Standardized error type assertions using `testutil.AssertErrorType`.
  - Improved test helper compatibility with both `*testing.T` and `*rapid.T`.

### Removed
- **Directory Size Calculation:** Removed automatic directory size calculation feature to simplify codebase
  - Removed `GetDirSize` method from FileSystem interface and all implementations (LocalFS, SftpFS)
  - Removed directory size cache and related persistence logic (`LRUCache`, `SizeCacheEntry`, `GetSizeCachePath`)
  - Removed background size calculation workers and batch update mechanism
  - Directories now display with blank size field (only file sizes are shown)

### Fixed
- **Footer Backgrounds:** Ensured footer text inputs correctly inherit the theme's background color.
- **CLI Remote Aliases:** Fixed `fm -r <alias>` to correctly parse and use `~/.ssh/config` settings.
- **Breadcrumb issues on Linux** The leading / is now rendered in the primary color, and the duplication at the root directory (/ /) has been resolved.

## [v0.1.4] - 2026-01-10

### Security
- **SSH Host Key Verification:** Removed insecure fallback, now requires valid `~/.ssh/known_hosts`
- **Path Traversal Protection:** Added validation to prevent directory traversal attacks via malicious filenames
- **Filename Validation:** Added checks for invalid characters and path separators in rename operations
- **Password Memory Handling:** Zero password bytes immediately after use to minimize exposure in memory

### Performance
- **Persistent Directory Size Cache:** Implemented a disk-backed cache (`~/.cache/fm/sizes.gob`) to preserve directory sizes across application sessions.
- **MTime-Based Size Validation:** Added automatic invalidation of cached directory sizes by comparing filesystem modification times (`MTime`). The app now trusts the cache instantly if the directory hasn't changed, but triggers a background re-count if it has.
- **Improved Size Counting Loop:** Enhanced the batch update mechanism to ensure directory sizes continue counting in the background even when navigating between tabs or parent directories.
- **Concurrent Size Counter:** Reimplemented directory size calculation using a parallel walker with worker pools, significantly increasing speed on multi-core systems.
- **Batched UI Updates:** Implemented a debouncing mechanism for size updates (100ms batches), preventing TUI lag and flickering during large directory scans.
- **Actual Disk Usage (Blocks):** Switched from reporting "apparent size" to "actual disk usage" using filesystem blocks. This accurately handles sparse files (fixing the 128TB issue) and better reflects real disk consumption.
- **Filesystem Isolation:** Added "same device" detection (similar to `du -x`) to prevent size calculations from bleeding into virtual filesystems like `/proc`, `/sys`, or external mount points.
- **Filtered File Counting:** Optimized the walker to skip special files (sockets, devices) and correctly handle symlink sizes without following them, preventing double-counting and infinite loops.

### Fixed
- **Operation Locking:** Implemented a background process tracker to prevent race conditions (e.g., trying to delete a file that is currently being copied).
- **Real-time Git Synchronization:** The file watcher now monitors the `.git` directory to immediately update status markers (M, A, D) when commits or stage changes happen externally.
- **Cross-Tab Git Consistency:** Ensured Git repository state is correctly maintained and refreshed when switching between multiple tabs.
- **Locked File Detection:** Improved error handling for "File in use by another process" (Windows) and "Text file busy" (Unix) with user-friendly messages.
- **Enhanced Error Messaging:** Standardized error reporting across all file operations to provide better context (e.g., "Delete failed: permission denied").
- **Partial File Cleanup:** Operations now automatically remove incomplete destination files if a copy fails (e.g., due to disk full or cancellation).
- **Symlink Loop Protection:** Implemented path tracking in recursive directory operations to detect and prevent infinite loops from circular symbolic links.
- **Improved Move Atomicity:** Added clear warnings and error messages if a move operation succeeds in copying but fails to delete the source (notifying user of duplicates).
- **Graceful Error Handling:** Refactored main() to use run() function ensuring cleanup runs on all exit paths
- **Watcher Reliability:** Added automatic restart mechanism when file watcher encounters errors or closes unexpectedly
- **Resource Cleanup:** Deferred cleanup (Close()) now guaranteed to run before program exit
- **Memory Leaks:** Implemented LRU caches (max 1000 entries) for dirSizeCache, cursorMemory, and offsetMemory to prevent unbounded growth
- **File Descriptor Leak:** Fixed potential leak in copyFile when second fs.Open() fails
- **Race Condition:** Added pathGeneration counter to prevent stale directory contents from displaying after rapid navigation
- **Error Messages:** Standardized error messages with consistent format: "operation failed: context: details"
- **Git Performance:** Significantly improved Git integration performance
  - Combined multiple git commands into single call (repo root + branch together)
  - Limited git status scope to current directory only instead of entire repository
  - Added in-memory caching for git status results to avoid redundant git calls
  - Cache automatically invalidates on file system changes

### Added
- **Permissions Visibility:** Files and directories now display their access status.
  - Items without write permission are marked with a `!` indicator
  - Items without read permission are dimmed in the list
  - Friendly "Access Denied" messages when attempting to enter restricted directories
- **Read-Only Filesystem Support:** Automatically detects and respects read-only mount points.
  - Displays a red `[RO]` indicator in the header when browsing read-only partitions
  - Destructive operations (Delete, Rename, Cut, Paste) are strictly disabled to prevent errors
- **Interactive Conflict Resolution:** Added a new system to handle filename collisions during copy/move operations.
  - Interactive prompt: Choose between Overwrite, Skip, or Auto-rename ([y/n/r])
  - Auto-rename logic: Automatically generates non-conflicting names (e.g., `file (1).txt`)
  - Batch processing: Remembers choices and continues with remaining items in a selection
- **Move (Cut/Paste) Support:** Implemented file and directory move operations via Cut ('x') and Paste ('v').
  - Dedicated "Cut" state to distinguish between copy and move operations
  - Background execution for non-blocking UI during large transfers
  - Cross-device move support with automatic fallback to Copy + Delete strategy
  - Comprehensive 80-100% test coverage for all core file operations
- **Settings Keybindings Section:** Added a new informational section to the settings screen listing all available keyboard shortcuts.
- **Add Tabs Support:** Navigate multiple directories simultaneously with independent tab sessions
  - Press `Alt+T` to create a new tab
  - Press `Alt+1` through `Alt+9` to switch between tabs (only works when multiple tabs exist)
  - Press `Alt+W` to close the current tab (when more than one tab is open)
  - Each tab maintains its own navigation state (path, cursor position, scroll offset)
  - Active tab highlighted in primary color in header breadcrumb
  - Tab indicators shown as [1] [2] [3] before the path breadcrumb
- **Theme-Aware Progress Bar:** Added visual feedback for long-running file operations in the footer.
  - Real-time progress updates for Copy, Move, and Delete operations.
  - Custom, responsive single-line implementation ([Label] [###...] 100%) within the footer.
  - Fully integrated with the current theme (Gruvbox, Monokai, etc.) using theme-specific colors.
  - Non-blocking implementation ensures the TUI remains responsive even during heavy I/O.
- **Input Validation:** Prevent command injection or path traversal in the "Search" and "Rename" inputs.

### Changed
- **Go Upgrade:** Bumped minimum Go version and toolchain to **1.25.5**, and upgraded all upgradable dependencies. 
- **Module Renaming:** Renamed the project module from `filemanager` to `fm` for a more concise internal import structure.
- **Streamlined Footer UI:** Removed persistent shortcut information from the main footer to reduce visual clutter.
  - Shortcut hints are now only shown when one or more files are selected (e.g., [c] Copy, [x] Cut, [r] Rename, [d] Delete)
  - Selection count correctly displayed in the footer
  - Full keybinding list moved to the Settings screen and Help screen
  - Settings screen footer retains navigation hints for better usability
- **FileSystem Interface:** Updated all interface methods to accept `context.Context` as the first parameter
- **TUI Commands:** All Bubble Tea commands now create contexts with appropriate timeouts before invoking file operations
- **Test Suite:** Updated all tests to pass `context.Background()` to function calls

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
- **Persistence:** JSON-based configuration management.
- **Navigation Memory:** Directory-specific cursor and scroll position tracking.
- **CLI Interface:** Help flags and themed command-line output.
- **Scaling:** Asynchronous directory loading and background watchers.
