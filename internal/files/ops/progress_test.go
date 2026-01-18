package ops

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestProgressWriter(t *testing.T) {
	mockFile := testutil.NewMockFile("test", nil)
	progChan := make(chan core.Progress, 10)
	pw := &progressWriter{
		Writer:   mockFile,
		Total:    10,
		Label:    "Testing",
		progChan: progChan,
	}

	data := []byte("12345")
	n, err := pw.Write(data)
	testutil.AssertNoError(t, err, "Write should succeed")
	testutil.AssertEqual(t, 5, n, "Write count")

	// Check if progress was sent
	select {
	case p := <-progChan:
		if p.Percent < 0 || p.Percent > 1.0 {
			t.Errorf("Invalid progress percent: %f", p.Percent)
		}
	default:
		// It's okay if it wasn't sent due to timing, but we want to cover the code
	}

	// Write more to reach total
	data2 := []byte("67890")
	_, _ = pw.Write(data2)

	// Final progress should definitely be sent because Current == Total
	var lastPercent float64
	for {
		select {
		case p := <-progChan:
			lastPercent = p.Percent
		default:
			goto done
		}
	}
done:
	if lastPercent != 1.0 {
		t.Errorf("Expected 100%% progress, got %f", lastPercent)
	}
}
