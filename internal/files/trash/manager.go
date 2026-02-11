package trash

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/logger"
)

// Manager handles trash operations.
type Manager struct {
	trashDir string
	filesDir string
	infoDir  string
	fs       core.FileSystem
}

// NewManager creates a new trash manager.
func NewManager(fs core.FileSystem) (*Manager, error) {
	if !fs.IsLocal() {
		return nil, fmt.Errorf("trash only supported for local filesystem")
	}

	homeDir, err := fs.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}

	trashDir := filepath.Join(homeDir, ".cache", "fm", "trash")
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")

	// Create directories if they don't exist
	for _, dir := range []string{trashDir, filesDir, infoDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create trash dir: %w", err)
		}
	}

	return &Manager{
		trashDir: trashDir,
		filesDir: filesDir,
		infoDir:  infoDir,
		fs:       fs,
	}, nil
}

// MoveToTrash moves a file or directory to trash.
func (m *Manager) MoveToTrash(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get file info
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	// Generate unique trashed name
	trashedName := generateTrashedName(path)
	trashedPath := filepath.Join(m.filesDir, trashedName)

	// Create metadata first (fail early if metadata write fails)
	metadata := createTrashInfo(path, trashedName, info)
	metadata.DeletionTime = time.Now()
	if err := m.writeMetadata(metadata); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	// Move file to trash (atomic if same filesystem)
	if err := os.Rename(path, trashedPath); err != nil {
		// Cleanup metadata on failure
		logger.LogIfError(m.deleteMetadata(trashedName), "trash: failed to cleanup metadata after failed move")
		return fmt.Errorf("move to trash: %w", err)
	}

	return nil
}

// List returns all items in the trash.
func (m *Manager) List() ([]TrashItem, error) {
	entries, err := os.ReadDir(m.infoDir)
	if err != nil {
		return nil, fmt.Errorf("read trash info: %w", err)
	}

	var items []TrashItem
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		trashedName := entry.Name()[:len(entry.Name())-5] // Remove .json
		info, err := m.readMetadata(trashedName)
		if err != nil {
			continue // Skip corrupted metadata
		}

		items = append(items, TrashItem{
			TrashedName:  info.TrashedName,
			OriginalPath: info.OriginalPath,
			DeletionTime: info.DeletionTime,
			SizeBytes:    info.SizeBytes,
			IsDirectory:  info.IsDirectory,
		})
	}

	return items, nil
}

// GetInfo returns metadata for a specific trashed item.
func (m *Manager) GetInfo(trashedName string) (*TrashInfo, error) {
	return m.readMetadata(trashedName)
}

// Delete permanently deletes an item from trash.
func (m *Manager) Delete(ctx context.Context, trashedName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	trashedPath := filepath.Join(m.filesDir, trashedName)
	markerPath := trashedPath + ".deleting"

	// Write deletion marker
	if err := os.WriteFile(markerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("write deletion marker: %w", err)
	}

	defer func() {
		logger.LogIfError(os.Remove(markerPath), "trash: failed to remove deletion marker")
	}()

	// Delete file
	if err := os.RemoveAll(trashedPath); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	// Delete metadata
	if err := m.deleteMetadata(trashedName); err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	return nil
}

// Empty removes all items from trash.
func (m *Manager) Empty(ctx context.Context) error {
	items, err := m.List()
	if err != nil {
		return fmt.Errorf("list trash: %w", err)
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := m.Delete(ctx, item.TrashedName); err != nil {
			return fmt.Errorf("delete %s: %w", item.TrashedName, err)
		}
	}

	return nil
}

// Restore moves an item from trash back to its original location.
func (m *Manager) Restore(ctx context.Context, trashedName string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := m.readMetadata(trashedName)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	trashedPath := filepath.Join(m.filesDir, trashedName)
	originalPath := info.OriginalPath

	// Check if original path exists
	if _, err := os.Stat(originalPath); err == nil {
		// File exists - conflict!
		return &RestoreConflictError{
			TrashedName:  trashedName,
			OriginalPath: originalPath,
		}
	}

	// Try to create parent directories if they don't exist
	parentDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Move file back
	if err := os.Rename(trashedPath, originalPath); err != nil {
		return fmt.Errorf("restore file: %w", err)
	}

	// Delete metadata
	if err := m.deleteMetadata(trashedName); err != nil {
		// File was restored but metadata cleanup failed
		return fmt.Errorf("delete metadata: %w", err)
	}

	return nil
}

// RestoreWithOverwrite restores an item, overwriting if it exists.
func (m *Manager) RestoreWithOverwrite(ctx context.Context, trashedName string) error {
	info, err := m.readMetadata(trashedName)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	// Remove existing file if present
	logger.LogIfError(os.RemoveAll(info.OriginalPath), "trash: failed to remove existing item for overwrite")

	return m.Restore(ctx, trashedName)
}

// RestoreWithRename restores an item with a new name to avoid conflicts.
func (m *Manager) RestoreWithRename(ctx context.Context, trashedName, newName string) error {
	info, err := m.readMetadata(trashedName)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}

	trashedPath := filepath.Join(m.filesDir, trashedName)
	parentDir := filepath.Dir(info.OriginalPath)
	newPath := filepath.Join(parentDir, newName)

	// Create parent directories if needed
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	// Move to new location
	if err := os.Rename(trashedPath, newPath); err != nil {
		return fmt.Errorf("restore file: %w", err)
	}

	// Delete metadata
	if err := m.deleteMetadata(trashedName); err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	return nil
}

// RecoverInterruptedDeletions completes any deletions that were interrupted.
func (m *Manager) RecoverInterruptedDeletions(ctx context.Context) error {
	entries, err := os.ReadDir(m.filesDir)
	if err != nil {
		return fmt.Errorf("read files dir: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".deleting" {
			trashedName := entry.Name()[:len(entry.Name())-9] // Remove .deleting
			logger.LogIfError(m.Delete(ctx, trashedName), "trash: failed to complete interrupted deletion")
		}
	}

	return nil
}

// AutoCleanup removes old items based on age and size limits.
func (m *Manager) AutoCleanup(ctx context.Context, maxAgeDays int, maxSizeMB int64) error {
	// Check if cleanup should run (24-hour throttle)
	lastCleanupPath := filepath.Join(m.trashDir, ".last_cleanup")
	if data, err := os.ReadFile(lastCleanupPath); err == nil {
		if lastTime, err := time.Parse(time.RFC3339, string(data)); err == nil {
			if time.Since(lastTime) < 24*time.Hour {
				return nil // Skip cleanup
			}
		}
	}

	items, err := m.List()
	if err != nil {
		return fmt.Errorf("list trash: %w", err)
	}

	// Delete items older than maxAgeDays
	if maxAgeDays > 0 {
		cutoff := time.Now().Add(-time.Duration(maxAgeDays) * 24 * time.Hour)
		for _, item := range items {
			if item.DeletionTime.Before(cutoff) {
				logger.LogIfError(m.Delete(ctx, item.TrashedName), "trash: failed to delete old item")
			}
		}
	}

	// Delete oldest items if size limit exceeded
	if maxSizeMB > 0 {
		items, err = m.List() // Refresh list
		if err != nil {
			logger.Errorf("trash: failed to list items during auto-cleanup: %v", err)
		}
		totalSize := int64(0)
		for _, item := range items {
			totalSize += item.SizeBytes
		}

		maxBytes := maxSizeMB * 1024 * 1024
		if totalSize > maxBytes {
			// Sort by deletion time (oldest first)
			for i := 0; i < len(items)-1; i++ {
				for j := i + 1; j < len(items); j++ {
					if items[i].DeletionTime.After(items[j].DeletionTime) {
						items[i], items[j] = items[j], items[i]
					}
				}
			}

			// Delete oldest until under limit
			for _, item := range items {
				if totalSize <= maxBytes {
					break
				}
				logger.LogIfError(m.Delete(ctx, item.TrashedName), "trash: failed to delete item exceeding size limit")
				totalSize -= item.SizeBytes
			}
		}
	}

	// Update last cleanup time
	logger.LogIfError(os.WriteFile(lastCleanupPath, []byte(time.Now().Format(time.RFC3339)), 0644), "trash: failed to update last cleanup time")

	return nil
}
