// Package ops implements file operations including copy, move, delete, search, and archive handling.
// All operations support context cancellation, progress reporting, and conflict resolution.
package ops

import (
	"context"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/git"
)

// OpContext encapsulates common execution context for file operations.
type OpContext struct {
	Context  context.Context
	FS       core.FileSystem
	Progress chan<- core.Progress
}

// ConflictOptions encapsulates conflict resolution strategy.
type ConflictOptions struct {
	Policy     conflict.Policy
	ApplyToAll bool
}

// CopyOptions encapsulates parameters for copy operations.
type CopyOptions struct {
	OpCtx    OpContext
	SrcFS    core.FileSystem // Used for cross-fs ops
	Src      string
	Dst      string
	Conflict ConflictOptions
}

// BatchOptions encapsulates parameters for batch operations.
type BatchOptions struct {
	OpCtx    OpContext
	SrcFS    core.FileSystem // Used for cross-fs ops
	Sources  []string
	DestDir  string
	Conflict ConflictOptions
}

// SearchOptions encapsulates parameters for search operations.
type SearchOptions struct {
	OpCtx OpContext
	Git   git.GitService
	Root  string
	Query string
	Regex bool
}

// ZipOptions encapsulates parameters for zip/unzip operations.
type ZipOptions struct {
	OpCtx    OpContext
	Srcs     []string // For Zip
	Src      string   // For Unzip
	Dst      string
	Conflict ConflictOptions
}

// CreateOptions encapsulates parameters for creating files/directories.
type CreateOptions struct {
	OpCtx    OpContext
	Path     string
	IsDir    bool
	Conflict ConflictOptions
}

// RenameOptions encapsulates parameters for renaming items.
type RenameOptions struct {
	OpCtx    OpContext
	OldPath  string
	NewPath  string
	Conflict ConflictOptions
}

// OpenOptions encapsulates parameters for opening items.
type OpenOptions struct {
	FS        core.FileSystem
	Path      string
	EditorIdx int
	Line      int
}

// TrashOptions encapsulates trash-related configuration.
type TrashOptions struct {
	UseTrash bool
}

// DeleteOptions encapsulates parameters for deleting items.
type DeleteOptions struct {
	OpCtx OpContext
	Paths []string
	Trash TrashOptions
}
