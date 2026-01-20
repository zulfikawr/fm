package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	version, err := CheckForUpdate()
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v", err)
	}

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

func TestCheckForUpdate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	_, err := CheckForUpdate()
	if err == nil {
		t.Error("CheckForUpdate() should have failed with invalid JSON")
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
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	if n != 5 {
		t.Errorf("Read count = %d, want 5", n)
	}

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
		return 0, nil
	}
	n = copy(p, m.data[m.off:])
	m.off += n
	return n, nil
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

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
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("Content = %s, want %s", got, content)
	}
}

func TestDownloadAndInstall_NoBinary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{
			TagName: "v99.99.99",
			Assets: []Asset{
				{Name: "fm-wrong-os-arch", BrowserDownloadURL: "http://example.com"},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	progress := make(chan float64, 10)
	err := DownloadAndInstall("v99.99.99", progress)
	if err == nil {
		t.Error("DownloadAndInstall() should have failed because no binary found")
	}
}

func TestDownloadAndInstall_HttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	progress := make(chan float64, 10)
	err := DownloadAndInstall("v99.99.99", progress)
	if err == nil {
		t.Error("DownloadAndInstall() should have failed with 404 status")
	}
}

func TestDownloadAndInstall_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	progress := make(chan float64, 10)
	err := DownloadAndInstall("v99.99.99", progress)
	if err == nil {
		t.Error("DownloadAndInstall() should have failed with invalid JSON")
	}
}

func TestDownloadAndInstall_Success(t *testing.T) {
	binaryName := fmt.Sprintf("fm-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/binary" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("fake binary content"))
			return
		}
		release := Release{
			TagName: "v99.99.99",
			Assets: []Asset{
				{
					Name:               binaryName,
					BrowserDownloadURL: "http://" + r.Host + "/binary",
					Size:               19,
				},
			},
		}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	oldAPI := githubAPI
	githubAPI = server.URL + "/%s/%s"
	defer func() { githubAPI = oldAPI }()

	progress := make(chan float64, 100)
	_ = DownloadAndInstall("v99.99.99", progress)
}
