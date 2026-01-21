package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.1.11", "v0.1.10", true},
		{"v0.1.10", "v0.1.10", false},
		{"v0.1.9", "v0.1.10", false},
		{"0.1.11", "0.1.10", true},
		{"v1.0.0", "v0.1.0", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.latest, tt.current), func(t *testing.T) {
			if got := isNewer(tt.latest, tt.current); got != tt.want {
				t.Errorf("isNewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckForUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v99.99.99",
		}
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	version, err := CheckForUpdate()
	testutil.AssertNoError(t, err, "CheckForUpdate should succeed")

	if version != "v99.99.99" {
		t.Errorf("CheckForUpdate() = %v, want v99.99.99", version)
	}
}

func TestCheckForUpdate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	_, err := CheckForUpdate()
	if err == nil {
		t.Error("CheckForUpdate() should have failed with 500 status")
	}
}

func TestProgressReader(t *testing.T) {
	data := []byte("hello world")
	reader := &progressReader{
		Reader:   &mockReader{data: data},
		Total:    int64(len(data)),
		OnUpdate: make(chan float64, 1),
	}

	p := make([]byte, 5)
	n, err := reader.Read(p)
	testutil.AssertNoError(t, err, "Read should succeed")
	testutil.AssertEqual(t, 5, n, "Read count should match")

	select {
	case got := <-reader.OnUpdate:
		want := 5.0 / 11.0
		if got != want {
			t.Errorf("Progress = %v, want %v", got, want)
		}
	default:
		t.Error("Progress update not received")
	}
}

type mockReader struct {
	data []byte
	off  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.off >= len(m.data) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.off:])
	m.off += n
	return n, nil
}

func TestCopyFile(t *testing.T) {
	tmpDir := testutil.TempDir(t)

	srcPath := filepath.Join(tmpDir, "src")
	dstPath := filepath.Join(tmpDir, "dst")
	content := []byte("test content")

	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	got, err := os.ReadFile(dstPath)
	testutil.AssertNoError(t, err, "ReadFile should succeed")
	if string(got) != string(content) {
		t.Errorf("Content = %s, want %s", got, content)
	}
}
