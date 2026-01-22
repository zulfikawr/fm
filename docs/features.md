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

## 🎨 Theme System

`fm` supports multiple themes to match your terminal's aesthetic. You can cycle through themes in the settings menu (`.`). Popular themes include:
- Nord
- Dracula
- Gruvbox
- Catppuccin
- And more...

## 📜 Operation Logs

Long-running operations like copying large directories or searching through millions of lines happen in the background. Press `Alt+L` to view the Operation Logs, where you can track the progress of every task in real-time.
