package ops

import "testing"

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"file.txt", false},
		{"file with spaces.txt", false},
		{"", true},
		{`.`, true},
		{"..", true},
		{"/path/to/file", true},
		{"C:\\path\\to\\file", true},
		{"file:name", true},
		{"file*name", true},
		{"file?name", true},
		{"file\"name", true},
		{"file<name", true},
		{"file>name", true},
		{"file|name", true},
		{string(make([]byte, 256)), true}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateFileName(tt.name); (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		base    string
		target  string
		wantErr bool
	}{
		{"/home/user", "file.txt", false},
		{"/home/user", "subdir/file.txt", true}, // Should fail because of /
		{"/home/user", "../other/file.txt", true},
		{"", "file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			if err := ValidatePath(tt.base, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q, %q) error = %v, wantErr %v", tt.base, tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSearchQuery(t *testing.T) {
	tests := []struct {
		query   string
		wantErr bool
	}{
		{"normal search", false},
		{"search with `backticks`", true},
		{"search with $dollar", true},
		{"search with ; semicolon", true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			if err := ValidateSearchQuery(tt.query); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchQuery(%q) error = %v, wantErr %v", tt.query, err, tt.wantErr)
			}
		})
	}
}
