package header

import (
	"fmt"
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderTabList(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	tests := []struct {
		name        string
		config      TabConfig
		wantEmpty   bool
		wantCount   int
		activeIndex int
	}{
		{
			name: "No tabs",
			config: TabConfig{
				TabCount:    0,
				ActiveIndex: 0,
			},
			wantEmpty: true,
		},
		{
			name: "Single tab",
			config: TabConfig{
				TabCount:    1,
				ActiveIndex: 0,
			},
			wantCount:   1,
			activeIndex: 0,
		},
		{
			name: "Multiple tabs",
			config: TabConfig{
				TabCount:    3,
				ActiveIndex: 1,
			},
			wantCount:   3,
			activeIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTabList(tt.config, styles)

			if tt.wantEmpty && result != "" {
				t.Error("Expected empty result for no tabs")
			}

			if !tt.wantEmpty {
				for i := 1; i <= tt.wantCount; i++ {
					if !strings.Contains(result, fmt.Sprintf("[%d]", i)) {
						t.Errorf("Tab list should contain [%d]", i)
					}
				}
			}
		})
	}
}

func TestCalculateTabWidth(t *testing.T) {
	tests := []struct {
		name         string
		tabCount     int
		showShortcut bool
		want         int
	}{
		{"No tabs", 0, false, 0},
		{"One tab", 1, true, 3},
		{"Three tabs", 3, true, 11}, // "[1] [2] [3]" = 3*3 + 2 spaces = 11
		{"Five tabs", 5, true, 19},  // "[1] [2] [3] [4] [5]" = 5*3 + 4 spaces = 19
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateTabWidth(tt.tabCount)
			if got != tt.want {
				t.Errorf("calculateTabWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldShowTabs(t *testing.T) {
	tests := []struct {
		name     string
		tabCount int
		want     bool
	}{
		{"Zero tabs", 0, false},
		{"One tab", 1, false},
		{"Two tabs", 2, true},
		{"Many tabs", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldShowTabs(tt.tabCount)
			if got != tt.want {
				t.Errorf("shouldShowTabs() = %v, want %v", got, tt.want)
			}
		})
	}
}
