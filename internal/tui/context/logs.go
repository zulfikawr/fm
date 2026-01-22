package context

import "time"

// --- Log State ---

// LogEntry represents a single entry in the operation log
type LogEntry struct {
	ID        string
	Timestamp time.Time
	Type      string
	Level     LogLevel
	Status    LogStatus
	Message   string
	Details   string
}

// LogState holds the history of operations
type LogState struct {
	Entries []LogEntry
	Cursor  int // Current scroll position in the log view
	Offset  int // Viewport offset
}

// AddEntry adds a new entry to the log state
func (ls *LogState) AddEntry(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	// Limit history to last 200 entries
	if len(ls.Entries) >= 200 {
		ls.Entries = ls.Entries[1:]
	}
	ls.Entries = append(ls.Entries, entry)
}

// UpdateStatus updates the status and level of an existing entry by ID
func (ls *LogState) UpdateStatus(id string, entry LogEntry) {
	for i := range ls.Entries {
		if ls.Entries[i].ID == id {
			ls.Entries[i].Status = entry.Status
			ls.Entries[i].Level = entry.Level
			if entry.Message != "" {
				ls.Entries[i].Message = entry.Message
			}
			if entry.Details != "" {
				ls.Entries[i].Details = entry.Details
			}
			break
		}
	}
}
