package listing

import (
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
)

// LoadOptions encapsulates parameters for loading directory contents.
type LoadOptions struct {
	FS          core.FileSystem
	Path        string
	SortMode    sorting.SortMode
	ShowHidden  bool
	GitStatuses map[string]string
}
