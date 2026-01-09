package files

// SortMode defines the sorting strategy for directory contents.
type SortMode int

const (
	SortDefault SortMode = iota
	SortName
	SortNameDesc
	SortNewest
	SortOldest
	SortSizeDesc
	SortSizeAsc
)

// String returns a human-readable representation of the sort mode.
func (s SortMode) String() string {
	switch s {
	case SortDefault:
		return "[ ⇅ ] Default"
	case SortName:
		return "[ A-Z ] Name (Asc)"
	case SortNameDesc:
		return "[ Z-A ] Name (Desc)"
	case SortNewest:
		return "[ ↓ ] Newest"
	case SortOldest:
		return "[ ↑ ] Oldest"
	case SortSizeDesc:
		return "[ ▼ ] Size (Lrg)"
	case SortSizeAsc:
		return "[ ▲ ] Size (Sml)"
	default:
		return "[ ? ] Unknown"
	}
}
