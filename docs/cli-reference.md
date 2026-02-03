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

### `fm search`
Perform fuzzy search for files and content within directories.

```bash
# Search in current directory
fm search <query>

# Search in specific directory
fm search <query> [path]

# Search remote server (requires SSH)
fm -r user@host search <query> [path]

# Examples
fm search "TODO" ./src
fm search "HandleUpdate" ./internal
fm search "config.json"
fm -r server search "error" /var/log
```

**Features:**
- Fuzzy file name matching
- Content search across files (substring matching)
- Case-insensitive matching
- Respects `.gitignore` rules
- Ignores `.git` directory
- Colorized output with match highlighting
- Shows line numbers for content matches

**Note:** Search uses substring/fuzzy matching, not regex patterns. For example:
- `fm search "Handle"` will match "HandleUpdate", "handleError", etc.
- `fm search "config"` will match "config.json", "configuration.md", lines containing "config", etc.

**Options:**
- `-s <query>` - Alternative flag syntax

---

### `fm analyze`
Analyze disk usage of a directory recursively.

```bash
# Analyze current directory
fm analyze

# Analyze specific path
fm analyze /home/user

# Analyze remote server disk usage
fm -r user@host analyze /var/log
```

**Features:**
- Recursive disk usage calculation
- Fast concurrent scanning
- Themed tree view with percentage bars
- Respects "One Filesystem" rule by default

**Options:**
- `-r, --remote <addr>` - Analyze remote disk usage via SFTP

---

### `fm info`
Display detailed information about files and directories.

```bash
# Show info for current directory
fm info
fm info .

# Show info for specific path
fm info /path/to/file
fm info ./directory

# JSON output for scripting
fm info --json .

# Tree view with depth control
fm info --tree --depth 2 .
fm info --tree --depth 0 .  # Unlimited depth

# Remote file info (requires SSH access)
fm -r user@host info /path/to/file
fm -r ssh-alias info /etc/hostname
```

**Options:**
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

These options work with any command and must be specified **before** the subcommand:

- `-r, --remote <address>` - Connect to remote SFTP server
  - Format: `user@host[:port]` or `user@host[:port][:path]`
  - Or use SSH alias from `~/.ssh/config`
  - Requires SSH key or interactive password authentication
  - **Must be specified before subcommand**: `fm -r host info /path` ✅ NOT `fm info -r host /path` ❌
- `-v, --version` - Show FM version
- `--help` - Show help for command

**Example Flag Order:**
```bash
fm -r user@server info /path        # Correct ✅
fm -r user@server search "query"    # Correct ✅
fm info -r user@server /path        # Wrong ❌ (flag after subcommand)
```

---

## Usage Patterns

### Quick File Inspection
```bash
# Check file details
fm info README.md

# Get size and permissions in JSON
fm info --json /var/log/app.log | jq '.size, .permissions'
```

### Directory Analysis
```bash
# See directory structure
fm info --tree --depth 3 ./src

# Count files and total size
fm info --json . | jq '.file_count, .total_size'
```

### Git Status Overview
```bash
# Check git stats for directory
fm info --json . | jq '.git_stats'

# Check specific branch
fm info /path/to/repo | grep Branch
```

### Search Operations
```bash
# Find TODOs in codebase
fm search "TODO" ./src

# Find function names
fm search "HandleUpdate" ./internal

# Search for config files
fm search "config" .

# Find imports (searches file content)
fm search "bubbletea" ./internal
```

### Remote Operations

**Note:** Remote operations require SSH key-based authentication or interactive password entry. The `-r` flag must come **before** the subcommand.

```bash
# Correct syntax: -r flag BEFORE subcommand
fm -r user@server info /var/log/app.log
fm -r user@server search "config"

# Inspect remote file info
fm -r user@server info /path/to/file

# Search remote directory
fm -r user@server search "API_KEY" /path

# Browse remote with TUI (interactive)
fm -r production-server

# Using SSH config alias
fm -r dev-server info /home/user
```

**Requirements:**
- SSH access to the remote server
- SSH key authentication (recommended) or password
- SFTP enabled on remote server

### Scripting Examples

**Bash - Get directory size:**
```bash
#!/bin/bash
size=$(fm info --json "$1" | jq -r '.total_size')
echo "Directory size: $size bytes"
```

**Python - Parse git stats:**
```python
import subprocess
import json

result = subprocess.run(['fm', 'info', '--json', '.'], 
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
    fm info --json $argv[1] | jq '.file_count'
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
   fm info --json . | jq
   fm search "pattern" | grep -i specific
   ```

2. **Alias common operations:**
   ```bash
   alias fmtree='fm info --tree --depth 2'
   alias fmj='fm info --json'
   ```

3. **Use in scripts:**
   - `--json` output is stable and parseable
   - All commands support piping
   - Error messages go to stderr

4. **Remote shortcuts:**
   - Define SSH aliases in `~/.ssh/config`
   - Use `-r alias` instead of full connection strings

---

## See Also

- [Getting Started](./getting-started.md) - Installation and basic usage
- [Features](./features.md) - Detailed TUI features
- [Configuration](./configuration.md) - Customization options
- [Remote Access](./remote-access.md) - SSH/SFTP setup
