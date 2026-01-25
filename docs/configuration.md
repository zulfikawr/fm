# Configuration

`fm` is highly configurable. Settings are automatically saved to and loaded from a flat JSON file.

## Configuration File Location

The configuration file is typically located at:
- **Linux/macOS:** `~/.config/fm/config.json`
- **Windows:** `%AppData%\fm\config.json`

If the directory or file doesn't exist, `fm` will create it with default values upon the first run.

## Available Settings

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
  "show_hidden": true,
  "case_sensitive": false,
  "wrap_navigation": true,
  "show_size": true,
  "show_date_modified": true,
  "show_header": true,
  "confirm_operations": true,
  "enable_git": true,
  "enable_mouse": true,
  "enable_icons": false,
  "use_trash": true,
  "theme_index": 0,
  "editor_index": 0,
  "date_format_index": 0,
  "size_format_index": 0
}
```