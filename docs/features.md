# Features

`fm` is packed with features designed to make terminal file management more powerful and efficient.

## 🌿 Git Integration

`fm` provides real-time Git status updates for your files and directories.
- **Status Markers:** Files are marked with `M` (Modified), `A` (Added), `D` (Deleted), `?` (Untracked), `!` (Ignored), or `U` (Unmerged).
- **Branch Display:** The current Git branch is visible in the UI header.
- **Performance:** Git status is fetched concurrently to avoid slowing down navigation.

## 🔎 Fuzzy Content Search (Find in Files)

One of the most powerful features of `fm` is the deep content search (`Alt+/`).
- **Concurrent Engine:** Uses a worker-pool architecture to search thousands of files quickly.
- **Smart Filtering:** Automatically ignores files specified in `.gitignore` and binary files.
- **Interactive Results:** Results are grouped by file. You can expand/collapse results with `Tab` and jump directly to a line in your editor by pressing `Enter`.

## 📦 Archive Management

Manage compressed files without leaving the TUI.
- **Zipping:** Select multiple files or directories and press `z` to create a new `.zip` archive.
- **Unzipping:** Highlight a `.zip` or `.tar.gz` file and press `u` to extract it to the current directory.
- **ZipSlip Protection:** Built-in security to prevent directory traversal attacks during extraction.

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
