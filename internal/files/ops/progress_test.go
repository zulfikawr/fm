package ops

import (
	"io"
	"testing"

	"fm/internal/files/core"
)

func TestProgressWriter(t *testing.T) {
	progChan := make(chan core.Progress, 10)
	pw := &progressWriter{
		Writer:   io.Discard,
		Total:    100,
		Label:    "Testing",
		progChan: progChan,
	}

	data := make([]byte, 50)
	n, err := pw.Write(data)
	if err != nil || n != 50 {
		t.Fatalf("Write failed: %v, %d", err, n)
	}

	// We might not get a message immediately because of the 100ms throttle,
	// but the final "Done" message should be sent if we reach total.

	n, err = pw.Write(data) // Total 100
	if err != nil || n != 50 {
		t.Fatalf("Write failed: %v, %d", err, n)
	}

	// Check if we got the 100% message
	var last core.Progress
loop:
	for {
		select {
		case p := <-progChan:
			last = p
		default:
			break loop
		}
	}

	if last.Percent != 1.0 {
		t.Errorf("Expected 1.0 progress, got %f", last.Percent)
	}
}
