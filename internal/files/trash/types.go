package trash

import (
	"fmt"
	"os"
	"time"
)

// TrashItem represents an item in the trash.
type TrashItem struct {
	TrashedName  string
	OriginalPath string
	DeletionTime time.Time
	SizeBytes    int64
	IsDirectory  bool
	Permissions  os.FileMode
}

// TrashInfo is the metadata stored for each trashed item.
type TrashInfo struct {
	Version      int       `json:"version"`
	OriginalPath string    `json:"original_path"`
	TrashedName  string    `json:"trashed_name"`
	DeletionTime time.Time `json:"deletion_time"`
	SizeBytes    int64     `json:"size_bytes"`
	IsDirectory  bool      `json:"is_directory"`
	Permissions  string    `json:"permissions"`
	OwnerUID     int       `json:"owner_uid"`
	OwnerGID     int       `json:"owner_gid"`
}

const MetadataVersion = 1

// RestoreConflictError indicates a file already exists at the restore destination.
type RestoreConflictError struct {
	TrashedName  string
	OriginalPath string
}

func (e *RestoreConflictError) Error() string {
	return fmt.Sprintf("file already exists at %s", e.OriginalPath)
}
