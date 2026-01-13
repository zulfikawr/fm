package ops

import (
	"testing"

	"fm/internal/constants"
)

func TestBufferPool(t *testing.T) {
	// Test acquisition
	buf := GetBuffer()
	if len(buf) != constants.CopyBufferSize {
		t.Errorf("Expected buffer size %d, got %d", constants.CopyBufferSize, len(buf))
	}

	// Test re-use (informal check)
	PutBuffer(buf)
	buf2 := GetBuffer()
	if len(buf2) != constants.CopyBufferSize {
		t.Errorf("Expected buffer size %d, got %d", constants.CopyBufferSize, len(buf2))
	}
}
