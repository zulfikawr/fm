package ops

import (
	"context"
	"strings"
	"testing"
)

func TestCancellableReader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	content := "hello world"
	r := NewCancellableReader(ctx, strings.NewReader(content))

	// Test normal read
	buf := make([]byte, 5)
	n, err := r.Read(buf)
	if err != nil || n != 5 || string(buf) != "hello" {
		t.Errorf("Normal read failed: n=%d, err=%v, buf=%s", n, err, string(buf))
	}

	// Test cancellation
	cancel()
	n, err = r.Read(buf)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v (n=%d)", err, n)
	}
}

func TestCancellableWriter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// Use a mock writer
	w := NewCancellableWriter(ctx, &strings.Builder{})

	// Test cancellation
	cancel()
	n, err := w.Write([]byte("test"))
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v (n=%d)", err, n)
	}
}
