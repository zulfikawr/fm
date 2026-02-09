# Trash Management

FM includes a built-in trash/recycle bin system that provides safe file deletion with the ability to restore items. Unlike traditional permanent deletion, the trash system allows you to recover accidentally deleted files.

## Overview

The trash system is:
- **Zero Dependencies**: Pure Go implementation, no external tools required
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Automatic Cleanup**: Configurable auto-deletion of old items
- **Conflict-Safe**: Handles restore conflicts intelligently
- **Crash-Resistant**: Recovers from interrupted operations

## Quick Start

### Enable Trash

1. Press `.` to open Settings
2. Navigate to "Use Trash (Move to Trash)"
3. Press `Enter` to toggle it on
4. Files will now move to trash instead of being permanently deleted

### Using Trash

```bash
# Delete a file (moves to trash if enabled)
d

# Open trash view
t

# In trash view:
r    # Restore selected item
d    # Delete permanently
e    # Empty entire trash
Esc  # Close trash view
```

## Trash Location

Trash files are stored in:
```
~/.cache/fm/trash/
├── .last_cleanup       # Cleanup timestamp
├── files/              # Actual trashed files
│   ├── report.pdf.1234567890
│   └── folder.1234567891/
└── info/               # Metadata (JSON)
    ├── report.pdf.1234567890.json
    └── folder.1234567891.json
```

## Features

### Auto-Cleanup

Trash automatically cleans up old items based on your configuration:

**Age-Based Cleanup:**
- Default: Items older than 30 days are automatically deleted
- Configurable: Set `trash_auto_cleanup_days` in config
- Disable: Set to `0` for unlimited retention

**Size-Based Cleanup:**
- Optional: Set maximum trash size in MB
- When exceeded, oldest items are deleted first
- Configurable: Set `trash_max_size_mb` in config
- Disable: Set to `0` for unlimited size

**Cleanup Schedule:**
- Runs on application startup
- 24-hour throttle (won't run more than once per day)
- Non-blocking background operation

### Restore Conflict Handling

When restoring a file, if the destination already exists:

**Automatic Resolution:**
- File is renamed with " (restored)" suffix
- Example: `report.pdf` → `report (restored).pdf`
- Original file at destination is preserved
- Restored file gets the new name

**Future Enhancement:**
- Interactive conflict dialog (planned)
- Options: Overwrite, Keep Both, Skip, Cancel

### Missing Path Recovery

If the original directory no longer exists:

**Automatic Recovery:**
- Parent directories are automatically recreated
- File is restored to original location
- Permissions: 0755 for created directories

**Fallback:**
- If parent creation fails, shows error
- File remains in trash for manual handling

### Crash Recovery

If FM crashes during deletion:

**Automatic Recovery:**
- `.deleting` marker files are detected on next startup
- Interrupted deletions are completed automatically
- No orphaned files or metadata
- Silent operation (logs only)

## Configuration

Edit `~/.config/fm/config.json`:

```json
{
  "use_trash": true,
  "trash_auto_cleanup_days": 30,
  "trash_max_size_mb": 500
}
```

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_trash` | boolean | `false` | Enable trash instead of permanent deletion |
| `trash_auto_cleanup_days` | integer | `30` | Auto-delete items older than N days (0 = unlimited) |
| `trash_max_size_mb` | integer | `0` | Maximum trash size in MB (0 = unlimited) |

### Examples

**Keep items for 7 days:**
```json
{
  "trash_auto_cleanup_days": 7
}
```

**Keep items forever:**
```json
{
  "trash_auto_cleanup_days": 0
}
```

**Limit trash to 1GB:**
```json
{
  "trash_max_size_mb": 1024
}
```

**Aggressive cleanup (3 days, 100MB):**
```json
{
  "trash_auto_cleanup_days": 3,
  "trash_max_size_mb": 100
}
```

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `r` | Restore selected item to original location |
| `d` | Delete selected item permanently (no recovery) |
| `e` | Empty entire trash (deletes all items) |
| `t` / `Esc` | Close trash view |

### Information Displayed

- **Name**: Original filename or directory name
- **Original Path**: Where the item was deleted from (truncated if long)
- **Deleted**: Human-readable time ago (e.g., "2 hours ago", "3 days ago")
- **Header**: Total item count and combined size

## Metadata

Each trashed item has associated metadata stored as JSON:

```json
{
  "version": 1,
  "original_path": "/home/user/documents/report.pdf",
  "trashed_name": "report.pdf.1234567890",
  "deletion_time": "2026-02-09T12:00:00Z",
  "size_bytes": 1048576,
  "is_directory": false,
  "permissions": "0644",
  "owner_uid": 1000,
  "owner_gid": 1000
}
```

This metadata enables:
- Accurate restoration to original location
- Conflict detection
- Size calculations
- Age-based cleanup
- Permission preservation

## Best Practices

### When to Use Trash

✅ **Use trash for:**
- Regular file management
- Cleaning up projects
- Removing temporary files
- Organizing directories

❌ **Don't use trash for:**
- Sensitive data (use permanent deletion)
- Very large files (consider size limits)
- Automated scripts (use permanent deletion)

### Maintenance

**Regular Cleanup:**
- Check trash periodically with `t`
- Empty trash when not needed: `e`
- Adjust `trash_auto_cleanup_days` based on usage

**Size Management:**
- Set `trash_max_size_mb` if disk space is limited
- Monitor trash size in trash view header
- Empty trash before large operations

**Performance:**
- Trash operations are fast (< 100ms for single files)
- Large directories may take longer
- Auto-cleanup runs in background (non-blocking)

## Troubleshooting

### Trash Not Working

**Check if enabled:**
1. Press `.` to open Settings
2. Look for "Use Trash (Move to Trash)"
3. Should show `[✓]` if enabled

**Check filesystem:**
- Trash only works on local filesystems
- Remote (SFTP) filesystems use permanent deletion
- Archive filesystems use permanent deletion

### Items Not Auto-Cleaning

**Check configuration:**
```bash
cat ~/.config/fm/config.json | grep trash
```

**Check last cleanup:**
```bash
cat ~/.cache/fm/trash/.last_cleanup
```

**Force cleanup:**
- Restart FM (cleanup runs on startup)
- Or wait 24 hours since last cleanup

### Restore Fails

**Common causes:**
- Insufficient permissions
- Disk full
- Original path on different filesystem

**Solutions:**
- Check permissions on destination directory
- Free up disk space
- Manually move file from `~/.cache/fm/trash/files/`

### Trash Taking Too Much Space

**Quick fix:**
1. Press `t` to open trash
2. Press `e` to empty trash
3. Confirm operation

**Long-term solution:**
```json
{
  "trash_auto_cleanup_days": 7,
  "trash_max_size_mb": 500
}
```

## Technical Details

### Naming Convention

Trashed files use unique names to prevent collisions:
```
{original_name}.{unix_timestamp_nanoseconds}
```

Example:
- Original: `report.pdf`
- Trashed: `report.pdf.1707480000123456789`

### Atomic Operations

All trash operations are atomic:
1. Write metadata first (fail early)
2. Move file to trash (atomic rename when possible)
3. On failure, cleanup metadata
4. Use `.deleting` markers for crash safety

### Cross-Filesystem Moves

When trash is on a different filesystem:
- Falls back to copy + delete
- Slower than atomic rename
- Still safe and reliable

### Performance

**Operation Times:**
- Move to trash: < 100ms (single file)
- Restore: < 100ms (single file)
- List trash: < 50ms (1000 items)
- Auto-cleanup: < 1s (1000 items)

**Memory Usage:**
- Minimal overhead (< 10MB for 1000 items)
- Metadata loaded on-demand
- No persistent cache

## See Also

- [Configuration](./configuration.md) - Customize trash settings
- [Keybindings](./keybindings.md) - All keyboard shortcuts
- [Features](./features.md) - Other FM features
