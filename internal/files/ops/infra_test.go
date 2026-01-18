package ops

import (
	"bytes"
	"context"
	"testing"

	"fm/internal/testutil"
)

func TestBufferPool(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}
	PutBuffer(buf)
}

func TestCancellableIO(t *testing.T) {
	t.Run("CancellableReader", func(t *testing.T) {
		data := []byte("test data")
		inner := bytes.NewReader(data)
		ctx, cancel := context.WithCancel(context.Background())

		r := NewCancellableReader(ctx, inner)

		buf := make([]byte, 4)
		n, err := r.Read(buf)
		testutil.AssertNoError(t, err, "Read should succeed")
		testutil.AssertEqual(t, 4, n, "Read count should match")

		cancel()
		n, err = r.Read(buf)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
		testutil.AssertEqual(t, 0, n, "Read count should be 0 after cancel")
	})

	t.Run("CancellableWriter", func(t *testing.T) {
		var inner bytes.Buffer
		ctx, cancel := context.WithCancel(context.Background())

		w := NewCancellableWriter(ctx, &inner)

		data := []byte("test")
		n, err := w.Write(data)
		testutil.AssertNoError(t, err, "Write should succeed")
		testutil.AssertEqual(t, 4, n, "Write count should match")

		cancel()
		n, err = w.Write(data)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
		testutil.AssertEqual(t, 0, n, "Write count should be 0 after cancel")
	})
}
