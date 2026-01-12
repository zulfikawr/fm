package state

import "testing"

func TestConflictState_Clear(t *testing.T) {
	c := ConflictState{
		Source:       "/src/file.txt",
		Destination:  "/dst/file.txt",
		PendingItems: []string{"/src/other.txt"},
		IsMove:       true,
	}

	c.Clear()

	if c.Source != "" {
		t.Errorf("Source should be empty, got %q", c.Source)
	}
	if c.Destination != "" {
		t.Errorf("Destination should be empty, got %q", c.Destination)
	}
	if c.PendingItems != nil {
		t.Errorf("PendingItems should be nil, got %v", c.PendingItems)
	}
	if c.IsMove {
		t.Error("IsMove should be false")
	}
}

func TestConflictState_Set(t *testing.T) {
	c := ConflictState{}

	pending := []string{"/src/a.txt", "/src/b.txt"}
	c.Set("/src/file.txt", "/dst/file.txt", pending, true)

	if c.Source != "/src/file.txt" {
		t.Errorf("Source = %q, want %q", c.Source, "/src/file.txt")
	}
	if c.Destination != "/dst/file.txt" {
		t.Errorf("Destination = %q, want %q", c.Destination, "/dst/file.txt")
	}
	if len(c.PendingItems) != 2 {
		t.Errorf("PendingItems length = %d, want 2", len(c.PendingItems))
	}
	if !c.IsMove {
		t.Error("IsMove should be true")
	}
}

func TestConflictState_HasConflict(t *testing.T) {
	tests := []struct {
		name   string
		state  ConflictState
		expect bool
	}{
		{
			name:   "empty state",
			state:  ConflictState{},
			expect: false,
		},
		{
			name: "only source",
			state: ConflictState{
				Source: "/src/file.txt",
			},
			expect: false,
		},
		{
			name: "only destination",
			state: ConflictState{
				Destination: "/dst/file.txt",
			},
			expect: false,
		},
		{
			name: "both set",
			state: ConflictState{
				Source:      "/src/file.txt",
				Destination: "/dst/file.txt",
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.HasConflict(); got != tt.expect {
				t.Errorf("HasConflict() = %v, want %v", got, tt.expect)
			}
		})
	}
}
