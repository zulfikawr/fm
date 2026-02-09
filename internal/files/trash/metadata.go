package trash

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// writeMetadata writes trash metadata to a JSON file.
func (m *Manager) writeMetadata(info *TrashInfo) error {
	metaPath := filepath.Join(m.infoDir, info.TrashedName+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	return os.WriteFile(metaPath, data, 0644)
}

// readMetadata reads trash metadata from a JSON file.
func (m *Manager) readMetadata(trashedName string) (*TrashInfo, error) {
	metaPath := filepath.Join(m.infoDir, trashedName+".json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var info TrashInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return &info, nil
}

// deleteMetadata removes the metadata file for a trashed item.
func (m *Manager) deleteMetadata(trashedName string) error {
	metaPath := filepath.Join(m.infoDir, trashedName+".json")
	return os.Remove(metaPath)
}

// createTrashInfo creates metadata for a file being moved to trash.
func createTrashInfo(path, trashedName string, info os.FileInfo) *TrashInfo {
	uid, gid := getOwnership(info)
	return &TrashInfo{
		Version:      MetadataVersion,
		OriginalPath: path,
		TrashedName:  trashedName,
		DeletionTime: info.ModTime(),
		SizeBytes:    info.Size(),
		IsDirectory:  info.IsDir(),
		Permissions:  info.Mode().String(),
		OwnerUID:     uid,
		OwnerGID:     gid,
	}
}

// getOwnership extracts UID and GID from FileInfo (Unix-specific).
func getOwnership(info os.FileInfo) (uid, gid int) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Uid), int(stat.Gid)
	}
	return 0, 0
}

// generateTrashedName creates a unique name for a trashed item.
func generateTrashedName(originalPath string) string {
	base := filepath.Base(originalPath)
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	return base + "." + timestamp
}
