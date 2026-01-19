package file

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

func TestList_Render(t *testing.T) {
	styles := theme.GetStylesheet(0)

	items := []core.Item{
		{Name: "file1.txt", Path: "/test/file1.txt"},
		{Name: "file2.txt", Path: "/test/file2.txt"},
	}

	props := Props{
		Width:  80,
		Height: 10,
		Items:  items,
		Styles: styles,
	}

	v := Render(props)
	stripped := testutil.StripANSI(v)

	if !strings.Contains(stripped, "file1.txt") {
		t.Errorf("expected file1.txt in view, got:\n%s", stripped)
	}
	if !strings.Contains(stripped, "file2.txt") {
		t.Errorf("expected file2.txt in view, got:\n%s", stripped)
	}
}

func TestList_CalculateLayout_Responsiveness(t *testing.T) {
	// Default date layout "02/01/2006 15:04" is 16 chars
	// sizeWidth is 11 chars
	// columnGap is 2 chars
	// markerWidth=0, gitMarkerWidth=2, permIndicatorWidth=1, safety=2 -> total overhead 5

	tests := []struct {
		name     string
		width    int
		showSize bool
		showDate bool
		wantSize bool
		wantDate bool
	}{
		{
			name:     "Enough space for both",
			width:    80,
			showSize: true,
			showDate: true,
			wantSize: true,
			wantDate: true,
		},
		{
			name:     "Small space - hide date",
			width:    40,
			showSize: true,
			showDate: true,
			wantSize: true,
			wantDate: false,
		},
		{
			name:     "Very small space - hide both",
			width:    25,
			showSize: true,
			showDate: true,
			wantSize: false,
			wantDate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := Props{
				Width:            tt.width,
				ShowSize:         tt.showSize,
				ShowDateModified: tt.showDate,
			}
			layout := calculateLayout(props)
			if layout.ShowSize != tt.wantSize {
				t.Errorf("ShowSize = %v, want %v", layout.ShowSize, tt.wantSize)
			}
			if layout.ShowDate != tt.wantDate {
				t.Errorf("ShowDate = %v, want %v", layout.ShowDate, tt.wantDate)
			}
		})
	}
}
