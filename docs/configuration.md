# Configuration

`fm` is highly configurable. Settings are automatically saved to and loaded from JSON files.

## Configuration File Locations

- **Main Config:** `~/.config/fm/config.json` (Linux/macOS) or `%AppData%\fm\config.json` (Windows)
- **Keybindings:** `~/.config/fm/keybindings.json` (Linux/macOS) or `%AppData%\fm\keybindings.json` (Windows)

If these files don't exist, `fm` will create them with default values upon the first run.

## Main Configuration Settings

The configuration uses a flat JSON structure. Below are the available fields:

- `show_hidden` (boolean): Whether to show hidden files (starting with `.`). Default: `true`.
- `case_sensitive` (boolean): Use case-sensitive sorting and filtering. Default: `false`.
- `wrap_navigation` (boolean): If true, moving down at the bottom of a list wraps to the top. Default: `true`.
- `show_size` (boolean): Display file sizes in the listing. Default: `true`.
- `show_date_modified` (boolean): Display the last modified date. Default: `true`.
- `show_header` (boolean): Show the top header with path and tab information. Default: `true`.
- `confirm_operations` (boolean): Ask for confirmation before potentially destructive actions. Default: `true`.
- `enable_git` (boolean): Enable or disable real-time Git status markers. Default: `true`.
- `enable_mouse` (boolean): Enable or disable mouse interaction (scrolling, clicking). Default: `true`.
- `enable_icons` (boolean): Enable or disable Nerd Font icons for files and folders. Default: `false`.
- `enable_regex_search` (boolean): Enable Regular Expression support for content search. Default: `false`.
- `show_ram_usage` (boolean): Display RAM usage (in MB) in the footer next to sort mode. Default: `false`.
- `use_trash` (boolean): If true, `d` is intended to move files to the system trash. *Note: This feature is currently under maintenance and may perform permanent deletions.*
- `theme_index` (integer): The index of the active theme.
- `editor_index` (integer): The index of your preferred editor (0: vim, 1: nano, 2: code, etc.).
- `date_format_index` (integer): Index for preferred date display format.
- `size_format_index` (integer): Index for preferred size display format.

## Editing Configuration

### Via Settings Menu
The easiest way to change settings is within `fm` itself:
1. Press `.` to open the settings menu.
2. Use the arrow keys to navigate and `Enter` to toggle or change values.
3. Press `r` within the settings menu to reset all settings to defaults.
4. Changes are saved automatically when you exit the menu or the application.

### Manual Editing
You can also edit the `config.json` file manually. Ensure you use a flat JSON structure.

Example `config.json`:
```json
{
  "config_version": 1,
  "show_hidden": true,
  "case_sensitive": false,
  "wrap_navigation": true,
  "show_size": true,
  "show_date_modified": true,
  "show_header": true,
  "show_ram_usage": false,
  "confirm_operations": true,
  "enable_git": true,
  "enable_mouse": true,
  "enable_icons": false,
  "enable_regex_search": false,
  "use_trash": true,
  "theme_index": 0,
  "editor_index": 0,
  "date_format_index": 0,
  "size_format_index": 0,
  "keybindings": []
}
```

## Custom Keybindings

`fm` supports fully customizable keybindings. You can modify them through the Settings menu or by editing `~/.config/fm/keybindings.json` directly.

### Keybinding File Structure

The keybindings file uses the following JSON structure:

```json
{
  "version": 1,
  "keybindings": [
    {
      "action": "open",
      "keys": ["enter", "l", "right"],
      "category": "navigation"
    },
    {
      "action": "quit",
      "keys": ["ctrl+c"],
      "category": "general"
    }
  ]
}
```

### Keybinding Fields

- `action` (string): The action identifier (e.g., "open", "copy", "quit")
- `keys` (array): List of key combinations that trigger this action
- `category` (string): Grouping category (navigation, file_ops, tabs, selection, search, general)

### Available Actions

**Navigation:**
- `open` - Open file/directory
- `go_parent` - Go to parent directory
- `move_down` / `move_up` - Move cursor
- `page_down` / `page_up` - Page navigation
- `go_to_path` - Go to specific path
- `history_back` / `history_forward` - Navigate history
- `cycle_sort` - Change sort mode

**File Operations:**
- `copy` - Copy selected items
- `cut` - Cut selected items
- `paste` - Paste clipboard
- `rename` - Rename item
- `delete` - Delete item
- `create` - Create new file/directory
- `zip` - Create archive
- `unzip` - Extract archive

**Selection:**
- `toggle_selection` - Toggle item selection
- `select_all` - Select all items
- `clear_selection` - Clear selection

**Tabs:**
- `new_tab` - Create new tab
- `close_tab` - Close current tab
- `switch_tab_1` through `switch_tab_9` - Switch to specific tab

**Search & Filter:**
- `filter` - Enter filter mode
- `fuzzy_search` - Deep content search
- `toggle_regex_search` - Toggle regex mode

**General:**
- `quit` - Exit application
- `settings` - Open settings
- `help` - Toggle help
- `analyze` - Disk usage analyzer
- `clipboard_view` - View clipboard
- `logs_view` - View logs

### Key Format

Keys can be specified as:
- Single keys: `"a"`, `"enter"`, `"space"`, `"esc"`
- With modifiers: `"ctrl+c"`, `"alt+t"`, `"shift+j"`
- Special keys: `"up"`, `"down"`, `"left"`, `"right"`, `"pgup"`, `"pgdown"`, `"backspace"`, `"tab"`

### Customizing via Settings Menu

1. Press `.` to open Settings
2. Scroll to the Keybindings section
3. Select the action you want to rebind
4. Press `Enter` to start recording
5. Press the key(s) you want to bind (modifiers are automatically detected)
6. Press `Enter` to confirm

The system validates keybindings to prevent conflicts and warns about browser-hijacking keys (like `Ctrl+T`).

### Resetting Keybindings

To reset all keybindings to defaults:
- In Settings menu, press `r` to reset all settings including keybindings
- Or delete `~/.config/fm/keybindings.json` and restart `fm`