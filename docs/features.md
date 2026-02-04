# Features

`fm` is packed with features designed to make terminal file management more powerful and efficient.

## 🌿 Git Integration

`fm` provides real-time Git status updates for your files and directories.
- **Status Markers:** Files are marked with `M` (Modified), `A` (Added), `D` (Deleted), `?` (Untracked), `!` (Ignored), or `U` (Unmerged).
- **Live Repository Summary:** The breadcrumb header displays the current Git branch and a summary of pending changes (e.g., `[main] • 3 Staged • 2 Modified • 1 Untracked`).
- **Performance:** Git status is fetched concurrently to avoid slowing down navigation.

## 🔎 Fuzzy Content Search (Find in Files)

One of the most powerful features of `fm` is the deep content search (`Alt+/`).
- **Concurrent Engine:** Uses a worker-pool architecture to search thousands of files quickly.
- **Regex Support:** Use full regular expressions by enabling "Regex Search" in settings or using the `--regex` flag in the CLI.
- **Smart Filtering:** Automatically ignores files specified in `.gitignore` and binary files.
- **Interactive Results:** Results are grouped by file. You can expand/collapse results with `Tab` and jump directly to a line in your editor by pressing `Enter`.

## 📦 Archive Management

Manage compressed files without leaving the TUI.
- **Zipping:** Select multiple files or directories and press `z` to create a new `.zip` archive.
- **Unzipping:** Highlight a `.zip` or `.tar.gz` file and press `u` to extract it to the current directory.
- **ZipSlip Protection:** Built-in security to prevent directory traversal attacks during extraction.

## 📊 Disk Usage Analysis

Quickly identify what is taking up space on your disk with the built-in analyzer (`Alt+U`).
- **Concurrent Scanner:** Uses a worker-pool to walk directory trees at high speed.
- **One Filesystem Rule:** Automatically stays within the same filesystem to avoid "fake" large files in virtual directories like `/proc` or `/sys`.
- **Interactive Heat-map:** Visual bars show what percentage of the parent directory each item consumes.
- **Actionable:** Drill down into subdirectories with `Enter` or delete space-hogs directly with `d`.
- **Remote Support:** Analyze disk usage on remote servers over SFTP just like your local machine.

## 🛡️ Conflict Resolution

When moving or copying files, `fm` handles name collisions gracefully:
- **Interactive Prompt:** If a file already exists, `fm` will ask you how to proceed.
- **Policies:** Choose between **Overwrite**, **Skip**, or **Auto-rename** (which appends a suffix to the new file).

## 🖱️ Mouse Support

`fm` features full mouse support for a more intuitive experience. For a detailed guide on all mouse interactions (including drag-to-select and drag-to-move), see the [**Mouse Support Guide**](./mouse-support.md).

- **Scrolling:** Use the mouse wheel to scroll through file lists, logs, settings, and fuzzy search results.
- **Selection:** Click on any file, directory, search result, or setting to focus it.
- **Navigation:**
  - Double-click on any file or directory to open it.
  - Double-click a fuzzy search result to open that file at the specific line.
- **Settings:** Double-click a setting to toggle it or cycle through available options.
- **Breadcrumbs:** Click on any part of the path breadcrumb in the header to navigate directly to that directory.
- **Fuzzy Search:** Click the `▼`/`▶` arrows in search results to expand or collapse files.
- **Tab Switching:** Click on tab indicators in the header to switch between active tabs.
- **Configurable:** Mouse support can be toggled on or off in the settings menu.

## 💎 Nerd Font Icons

`fm` supports modern file and folder icons using Nerd Fonts to enhance visual clarity.
- **Lazy Loading:** Icons are not bundled by default to keep the binary small. They are downloaded automatically when the feature is enabled.
- **Smart Mapping:** Icons are assigned based on file extensions, specific filenames (like `Dockerfile` or `LICENSE`), and folder names.
- **Theme Integrated:** Icon colors automatically match your chosen application theme.
- **Setup Guide:** For detailed installation instructions, see the [**Icons Guide**](./icons.md).

## 🎨 Theme System

`fm` supports multiple themes to match your terminal's aesthetic. You can cycle through themes in the settings menu (`.`). Popular themes include:
- Gruvbox (Default)
- Nord
- Dracula
- Monokai
- Solarized Dark
- Red
- Tokyo Night
- Rose Pine
- Catppuccin Mocha

## 📜 Operation Logs

Long-running operations like copying large directories or searching through millions of lines happen in the background. Press `Alt+L` to view the Operation Logs, where you can track the progress of every task in real-time.
