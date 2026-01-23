# Keybindings

`fm` relies heavily on keybindings for fast navigation and operation. Below is a comprehensive list of available shortcuts.

## Navigation

| Key | Action |
| --- | --- |
| `Enter` / `→` / `l` | Open directory / Open file in editor |
| `Backspace` / `←` / `h` | Navigate to parent directory |
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Shift` + `j` / `↓` | Range select down |
| `Shift` + `k` / `↑` | Range select up |
| `g` | Go to path (opens a prompt for Local or Remote path) |
| `[` | History Back |
| `]` | History Forward |

## Selection & Bulk Actions

| Key | Action |
| --- | --- |
| `Space` | Toggle selection for the current item |
| `Shift` + `Left Click` | Toggle selection / Range select |
| `Alt+A` | Select all items in the current directory |
| `Esc` | Clear all selections |

## Tabs

| Key | Action |
| --- | --- |
| `Alt+T` | Open a new tab (up to 9 tabs) |
| `Alt+W` | Close the current tab |
| `Alt+1` - `Alt+9` | Switch to the corresponding tab |

## File Operations

| Key | Action |
| --- | --- |
| `c` | Copy selected items to clipboard |
| `x` | Cut selected items to clipboard |
| `v` | Paste items from clipboard |
| `r` | Rename the highlighted item |
| `d` | Move selected items to trash (Note: currently performs permanent deletion) |
| `Alt+N` | Create a new file or folder (use `Tab` to toggle between them) |
| `z` | Create a Zip archive of selected items |
| `u` | Unzip the highlighted archive |

## Search & Filtering

| Key | Action |
| --- | --- |
| `/` | Enter filter mode (filters current directory listing) |
| `Alt+/` | Open Fuzzy Content Search (Find in Files) |
| `Alt+M` / `Alt+N` | Jump between files in fuzzy search results |
| `Esc` | Exit search/filter mode |

## Miscellaneous

| Key | Action |
| --- | --- |
| `Alt+C` | View current clipboard contents |
| `Alt+L` | View background operation logs |
| `s` | Cycle through 7 sort modes (Name, Size, Date, Extension, etc.) |
| `.` | Toggle settings menu |
| `r` | (In Settings Menu) Reset all settings to default |
| `Ctrl+C` | Quit `fm` |

## Mouse Support

| Action | Result |
| --- | --- |
| `Left Click` | Focus item |
| `Double Click` | Open file / Navigate to directory |
| `Scroll Wheel` | Scroll list / logs / settings |
| `Click Breadcrumb` | Navigate directly to that path |
| `Click Tab [n]` | Switch to that tab |