# CLI Reference

FM provides powerful command-line tools for quick file operations, inspections, and automation alongside its TUI interface.

## Commands

### `fm` (default)
Launch the TUI file manager.

```bash
# Open in current directory
fm

# Open in specific directory
fm /path/to/directory

# Open remote directory via SFTP
fm -r user@host:/path
fm -r ssh-alias
```

**Options:**
- `-r, --remote <addr>` - Connect to remote server via SFTP
- `-v, --version` - Show version information
- `--help` - Show help message

---

### `fm -c, --config`
Manage application configuration directly from the CLI.

```bash
# View current configuration
fm -c
fm --config

# Reset configuration to defaults
fm -c --reset

# Run interactive configuration wizard
fm -c --init
```

**Features:**
- **Themed View**: Displays current settings in a beautiful, theme-aware colorful list.
- **Easy Reset**: Quickly revert all settings to factory defaults if something goes wrong.
- **Interactive Init**: Step-by-step wizard to configure your environment without manually editing JSON.

**Options:**
- `-c, --config` - Manage configuration
- `--reset` - Reset configuration to default values
- `--init` - Launch the interactive configuration wizard

---

### `fm -s, --search`
Perform fuzzy or regex search for files and content within directories.

```bash
# Search in current directory (Fuzzy)
fm -s <query>
fm --search <query>

# Search using Regular Expressions
fm -s <query> --regex [path]
fm -s <query> -e [path]

# Search in specific directory
fm -s <query> [path]

# Search remote server (requires SSH)
fm -r user@host -s <query> [path]
```

**Features:**
- Fuzzy file name matching (default)
- Full Regular Expression support with `--regex` or `-e`
- Content search across files (substring matching)
- Case-insensitive matching
- Respects `.gitignore` rules
- Ignores `.git` directory
- Colorized output with match highlighting
- Shows line numbers for content matches

**Note:** Search uses substring/fuzzy matching, not regex patterns. For example:
- `fm -s "Handle"` will match "HandleUpdate", "handleError", etc.
- `fm -s "config"` will match "config.json", "configuration.md", lines containing "config", etc.

**Options:**
- `-s, --search <query>` - Search query
- `-e, --regex` - Use regular expressions for search

---

### `fm -a, --analyze`
Analyze disk usage of a directory recursively.

```bash
# Analyze current directory
fm -a
fm --analyze

# Analyze specific path
fm -a /home/user

# Analyze remote server disk usage
fm -r user@host -a /var/log
```

**Features:**
- Recursive disk usage calculation
- Fast concurrent scanning
- Themed tree view with percentage bars
- Respects "One Filesystem" rule by default

**Options:**
- `-a, --analyze` - Analyze disk usage
- `-r, --remote <addr>` - Analyze remote disk usage via SFTP

---

### `fm -i, --info`
Display detailed information about files and directories.

```bash
# Show info for current directory
fm -i
fm -i .
fm --info

# Show info for specific path
fm -i /path/to/file
fm -i ./directory

# JSON output for scripting
fm -i --json .

# Tree view with depth control
fm -i --tree --depth 2 .
fm -i --tree --depth 0 .  # Unlimited depth

# Remote file info (requires SSH access)
fm -r user@host -i /path/to/file
fm -r ssh-alias -i /etc/hostname
```

**Options:**
- `-i, --info` - Display file/directory information
- `--json` - Output in JSON format (machine-readable)
- `--tree` - Display directory tree structure
- `--depth N` - Tree depth limit (default: 2, 0 for unlimited)

**Output Includes:**

For **files**:
- Path, size, permissions, mode
- Modification time
- Read/write permissions
- Git status (if in repository)

For **directories**:
- Path, size, permissions
- File count, directory count
- Total size (recursive)
- Git repository info (root, branch)
- Git statistics (modified, added, deleted, untracked, staged)

**JSON Format:**
```json
{
  "path": "/path/to/file",
  "type": "file",
  "size": 1024,
  "size_formatted": "1.0 K",
  "permissions": "-rw-r--r--",
  "mode": "0644",
  "modified": "2026-02-01T12:00:00Z",
  "can_read": true,
  "can_write": true,
  "in_git_repo": true,
  "git_root": "/path/to/repo",
  "git_branch": "main",
  "git_status": "M "
}
```

**Tree View Example:**
```
Directory Tree

internal/
├── bootstrap/
│   ├── bootstrap.go (1.3 K)
│   └── bootstrap_test.go (1.5 K)
├── cli/
│   ├── cli.go (2.4 K)
│   └── help.go (3.3 K)
└── config/
    └── config.go (3.5 K)
```

---

## Global Options

These options work with any command:

- `-r, --remote <address>` - Connect to remote SFTP server
  - Format: `user@host[:port]` or `user@host[:port][:path]`
  - Or use SSH alias from `~/.ssh/config`
  - Requires SSH key or interactive password authentication
- `-v, --version` - Show FM version
- `--help` - Show help for command

**Example Usage:**
```bash
fm -r user@server -i /path        # Remote info
fm -r user@server -s "query"      # Remote search
fm -a /home/user                  # Local analyze
fm -c --reset                     # Reset config
```

---

## Usage Patterns

### Quick File Inspection
```bash
# Check file details
fm -i README.md

# Get size and permissions in JSON
fm -i --json /var/log/app.log | jq '.size, .permissions'
```

### Directory Analysis
```bash
# See directory structure
fm -i --tree --depth 3 ./src

# Count files and total size
fm -i --json . | jq '.file_count, .total_size'
```

### Git Status Overview
```bash
# Check git stats for directory
fm -i --json . | jq '.git_stats'

# Check specific branch
fm -i /path/to/repo | grep Branch
```

### Search Operations
```bash
# Find TODOs in codebase
fm -s "TODO" ./src

# Find function names
fm -s "HandleUpdate" ./internal

# Search for config files
fm -s "config" .

# Find imports (searches file content)
fm -s "bubbletea" ./internal
```

### Remote Operations

**Note:** Remote operations require SSH key-based authentication or interactive password entry.

```bash
# Inspect remote file info
fm -r user@server -i /var/log/app.log

# Search remote directory
fm -r user@server -s "config" /path

# Analyze remote disk usage
fm -r user@server -a /var/log

# Browse remote with TUI (interactive)
fm -r production-server

# Using SSH config alias
fm -r dev-server -i /home/user
```

**Requirements:**
- SSH access to the remote server
- SSH key authentication (recommended) or password
- SFTP enabled on remote server

### Scripting Examples

**Bash - Get directory size:**
```bash
#!/bin/bash
size=$(fm -i --json "$1" | jq -r '.total_size')
echo "Directory size: $size bytes"
```

**Python - Parse git stats:**
```python
import subprocess
import json

result = subprocess.run(['fm', '-i', '--json', '.'], 
                       capture_output=True, text=True)
data = json.loads(result.stdout)

if data['in_git_repo']:
    stats = data['git_stats']
    print(f"Modified: {stats['modified']}")
    print(f"Staged: {stats['staged']}")
```

**Fish - Count files recursively:**
```fish
function count-files
    fm -i --json $argv[1] | jq '.file_count'
end
```

---

## Exit Codes

- `0` - Success
- `1` - Error (file not found, permission denied, etc.)

---

## Configuration

CLI commands respect the same configuration file as the TUI:
- **Linux/macOS:** `~/.config/fm/config.json`
- **Windows:** `%AppData%\fm\config.json`

Settings used by CLI:
- `enable_git` - Enable/disable git integration
- `theme_index` - Color theme for output
- `size_format_index` - Size display format

---

## Tips

1. **Combine with other tools:**
   ```bash
   fm -i --json . | jq
   fm -s "pattern" | grep -i specific
   ```

2. **Alias common operations:**
   ```bash
   alias fmtree='fm -i --tree --depth 2'
   alias fmj='fm -i --json'
   alias fms='fm -s'
   ```

3. **Use in scripts:**
   - `--json` output is stable and parseable
   - All commands support piping
   - Error messages go to stderr

4. **Remote shortcuts:**
   - Define SSH aliases in `~/.ssh/config`
   - Use `-r alias` instead of full connection strings

5. **Navigate to directories with command-like names:**
   ```bash
   # Now you can navigate to directories named after commands
   fm config      # Opens ./config directory
   fm search      # Opens ./search directory
   fm info        # Opens ./info directory
   
   # Use flags for actual commands
   fm -c          # Opens config manager
   fm -s "query"  # Performs search
   fm -i          # Shows info
   ```

---

## See Also

- [Getting Started](./getting-started.md) - Installation and basic usage
- [Features](./features.md) - Detailed TUI features
- [Configuration](./configuration.md) - Customization options
- [Remote Access](./remote-access.md) - SSH/SFTP setup
