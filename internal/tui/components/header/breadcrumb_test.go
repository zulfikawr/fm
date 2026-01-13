package header

import (
	"strings"
	"testing"

	"fm/internal/tui/theme"
)

func TestRenderBreadcrumbPath(t *testing.T) {
	styles := theme.NewStylesheet(theme.Themes[0])

	tests := []struct {
		name      string
		path      string
		separator string
		remote    string
		want      []string // strings that should appear in output
	}{
		{
			name:      "Root path",
			path:      "/",
			separator: "/",
			remote:    "",
			want:      []string{"/"},
		},
		{
			name:      "Simple path",
			path:      "/home/user",
			separator: "/",
			remote:    "",
			want:      []string{"home", "user"},
		},
		{
			name:      "Deep path",
			path:      "/home/user/docs/projects",
			separator: "/",
			remote:    "",
			want:      []string{"home", "user", "docs", "projects"},
		},
		{
			name:      "Remote root",
			path:      "/home",
			separator: "/",
			remote:    "user@host",
			want:      []string{"user@host", "home"},
		},
		{
			name:      "Remote root exactly",
			path:      "/",
			separator: "/",
			remote:    "user@host",
			want:      []string{"user@host"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBreadcrumbPath(tt.path, tt.separator, tt.remote, styles)
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("breadcrumb should contain %q, got %q", want, result)
				}
			}
			if tt.name == "Remote root exactly" {
				// Result should ONLY contain remote address, not "/" root
				if strings.Contains(result, " > /") || strings.HasSuffix(result, " /") {
					t.Errorf("breadcrumb should not contain root / for remote, got %q", result)
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
