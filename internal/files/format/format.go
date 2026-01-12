package format

import (
	"fmt"
)

var DateFormats = []struct {
	Name   string
	Layout string
}{
	{"Default", "02/01/2006 15:04"},
	{"ISO", "2006-01-02 15:04"},
	{"US", "01/02/2006 03:04 PM"},
	{"Short", "02/01/06 15:04"},
}

var SizeFormats = []string{
	"Short (K, M, G)",
	"Full (KB, MB, GB)",
	"Bytes",
}

// FormatSize converts a byte count into a human-readable string based on the selected format.
func FormatSize(b int64, formatIdx int) string {
	if b < 0 {
		return ""
	}
	if formatIdx == 2 { // Bytes
		return fmt.Sprintf("%d B", b)
	}

	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := "KMGTPE"
	suffix := string(units[exp])
	if formatIdx == 1 { // Full
		suffix += "B"
	}

	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffix)
}
