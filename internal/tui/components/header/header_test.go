package header

import (
	"fmt"
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRender(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	t.Run("Settings Header", func(t *testing.T) {
		props := Props{
			Width:        80,
			SettingsOpen: true,
			Styles:       styles,
		}

		result := Render(props)
		if !strings.Contains(result, "Settings") {
			t.Error("Settings header should contain 'Settings'")
		}
	})

	t.Run("File Header Without Tabs", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/home/user/docs",
			Separator: "/",
			GitBranch: "",
			ReadOnly:  false,
			TabCount:  1,
			ActiveTab: 0,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "home") || !strings.Contains(result, "docs") {
			t.Error("Header should contain path components")
		}
		if strings.Contains(result, "[1]") {
			t.Error("Single tab should not show tab indicators")
		}
	})

	t.Run("File Header With Multiple Tabs", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/home/user",
			Separator: "/",
			TabCount:  3,
			ActiveTab: 1,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "[1]") || !strings.Contains(result, "[2]") || !strings.Contains(result, "[3]") {
			t.Error("Multiple tabs should show tab indicators")
		}
	})

	t.Run("File Header With Git Branch", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/project",
			Separator: "/",
			GitBranch: "main",
			TabCount:  1,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "main") {
			t.Error("Header should contain git branch")
		}
	})

	t.Run("File Header Read Only", func(t *testing.T) {
		props := Props{
			Width:     80,
			Path:      "/readonly",
			Separator: "/",
			ReadOnly:  true,
			TabCount:  1,
			Styles:    styles,
		}

		result := Render(props)
		if !strings.Contains(result, "RO") {
			t.Error("Header should contain RO indicator")
		}
	})
}

func TestRenderBreadcrumbPath(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	tests := []struct {
		name      string
		path      string
		separator string
		want      []string // strings that should appear in output
	}{
		{
			name:      "Root path",
			path:      "/",
			separator: "/",
			want:      []string{"/"},
		},
		{
			name:      "Simple path",
			path:      "/home/user",
			separator: "/",
			want:      []string{"home", "user"},
		},
		{
			name:      "Deep path",
			path:      "/home/user/docs/projects",
			separator: "/",
			want:      []string{"home", "user", "docs", "projects"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBreadcrumbPath(tt.path, tt.separator, styles)
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("breadcrumb should contain %q, got %q", want, result)
				}
			}
		})
	}
}

func TestAddGitBranch(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	tests := []struct {
		name       string
		breadcrumb string
		gitBranch  string
		wantGit    bool
	}{
		{
			name:       "With git branch",
			breadcrumb: "/project",
			gitBranch:  "main",
			wantGit:    true,
		},
		{
			name:       "Without git branch",
			breadcrumb: "/project",
			gitBranch:  "",
			wantGit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addGitBranch(tt.breadcrumb, tt.gitBranch, styles)

			if tt.gitBranch == "" {
				// If no git branch provided, result should equal breadcrumb
				hasGit := result != tt.breadcrumb
				if hasGit != tt.wantGit {
					t.Errorf("git branch presence = %v, want %v", hasGit, tt.wantGit)
				}
			} else {
				// If git branch provided, check if it's in the result
				hasGit := strings.Contains(result, tt.gitBranch)
				if hasGit != tt.wantGit {
					t.Errorf("git branch presence = %v, want %v", hasGit, tt.wantGit)
				}
			}
		})
	}
}

func TestAddReadOnlyIndicator(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	tests := []struct {
		name       string
		breadcrumb string
		readOnly   bool
		wantRO     bool
	}{
		{
			name:       "Read only",
			breadcrumb: "/path",
			readOnly:   true,
			wantRO:     true,
		},
		{
			name:       "Writable",
			breadcrumb: "/path",
			readOnly:   false,
			wantRO:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addReadOnlyIndicator(tt.breadcrumb, tt.readOnly, styles)
			hasRO := strings.Contains(result, "RO")
			if hasRO != tt.wantRO {
				t.Errorf("RO indicator presence = %v, want %v", hasRO, tt.wantRO)
			}
		})
	}
}

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
