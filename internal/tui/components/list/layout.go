package list

import (
	"fm/internal/files/format"
)

// Layout contains calculated dimensions for the list view
type Layout struct {
	ViewportHeight     int
	NameWidth          int
	DateWidth          int
	SizeWidth          int
	MarkerWidth        int
	GitMarkerWidth     int
	PermIndicatorWidth int
	ColumnGap          int
}

// CalculateLayout computes the dimensions for each column based on props
func CalculateLayout(props Props) Layout {
	sizeWidth := 11
	switch props.SizeFormatIndex {
	case 1:
		sizeWidth = 12
	case 2:
		sizeWidth = 15
	}

	dateWidth := len(props.DateLayout)
	if dateWidth < 10 {
		// Use a fallback if layout is not yet loaded or is very short
		// Attempt to check if we have format package available for default width
		dateWidth = 10
		if props.DateFormatIndex < len(format.DateFormats) {
			dateWidth = len(format.DateFormats[props.DateFormatIndex].Layout)
		}
	}

	const columnGap = 2
	markerWidth := 0
	if props.SelectMode {
		markerWidth = 4
	}
	gitMarkerWidth := 3 // git status + perm indicator space

	nameWidth := props.Width - markerWidth - gitMarkerWidth
	if props.ShowSize {
		nameWidth -= (sizeWidth + columnGap)
	}
	if props.ShowDateModified {
		nameWidth -= (dateWidth + columnGap)
	}

	if nameWidth < 1 {
		nameWidth = 1
	}

	return Layout{
		ViewportHeight:     props.Height,
		NameWidth:          nameWidth,
		DateWidth:          dateWidth,
		SizeWidth:          sizeWidth,
		MarkerWidth:        markerWidth,
		GitMarkerWidth:     gitMarkerWidth,
		PermIndicatorWidth: 1,
		ColumnGap:          columnGap,
	}
}
