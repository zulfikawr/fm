package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/logger"
	"golang.org/x/mod/semver"
)

const (
	repoOwner = "zulfikawr"
	repoName  = "fm"
	timeout   = 5 * time.Second
)

var (
	githubAPI  = "https://api.github.com/repos/%s/%s/releases/latest"
	executable = os.Executable
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a GitHub release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// CheckForUpdate checks if a new version is available
func CheckForUpdate() (string, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(fmt.Sprintf(githubAPI, repoOwner, repoName))
	if err != nil {
		return "", err
	}
	defer logger.CloseAndLog(resp.Body, "update check response body")

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if isNewer(release.TagName, constants.AppVersion) {
		return release.TagName, nil
	}

	return "", nil
}

func isNewer(latest, current string) bool {
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}
	if !strings.HasPrefix(current, "v") {
		current = "v" + current
	}
	return semver.Compare(latest, current) > 0
}

// DownloadAndInstall downloads the latest release and installs it
func DownloadAndInstall(version string, progress chan float64) error {
	resp, err := http.Get(fmt.Sprintf(githubAPI, repoOwner, repoName))
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(resp.Body, "update download metadata response body")

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	var targetAsset *Asset
	suffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	for i := range release.Assets {
		asset := release.Assets[i]
		if strings.HasSuffix(asset.Name, suffix) {
			targetAsset = &asset
			break
		}
	}

	if targetAsset == nil {
		return fmt.Errorf("no binary found for %s", suffix)
	}

	// Download the asset
	resp, err = http.Get(targetAsset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(resp.Body, "update binary download response body")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: %d", resp.StatusCode)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "fm-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		logger.LogIfError(os.Remove(tmpPath), "update: failed to remove temporary file")
	}()

	// Wrap response body for progress
	reader := &progressReader{
		Reader:   resp.Body,
		Total:    targetAsset.Size,
		OnUpdate: progress,
	}

	if n, err := io.Copy(tmpFile, reader); err != nil {
		logger.LogIfError(err, fmt.Sprintf("Failed to copy update (copied %d bytes)", n))
		logger.CloseAndLog(tmpFile, "temporary update file on copy error")
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Set executable permissions
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	// Replace the current executable
	selfPath, err := executable()
	if err != nil {
		return err
	}

	// On Unix, we can just rename. On Windows, it's more complex.
	if runtime.GOOS == "windows" {
		oldPath := selfPath + ".old"
		logger.LogIfError(os.Remove(oldPath), "update: failed to remove old executable")
		if err := os.Rename(selfPath, oldPath); err != nil {
			return err
		}
		if err := copyFile(tmpPath, selfPath); err != nil {
			return err
		}
	} else {
		// Try to move it directly
		if err := os.Rename(tmpPath, selfPath); err != nil {
			// If rename fails (e.g. across filesystems), try copying
			if err := copyFile(tmpPath, selfPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(s, "source file during update copy")

	d, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer logger.CloseAndLog(d, "destination file during update copy")

	n, err := io.Copy(d, s)
	if err != nil {
		logger.LogIfError(err, fmt.Sprintf("Failed to copy file (copied %d bytes)", n))
	}
	return err
}

type progressReader struct {
	io.Reader
	Total    int64
	Current  int64
	OnUpdate chan float64
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.Total > 0 && pr.OnUpdate != nil {
		pr.OnUpdate <- float64(pr.Current) / float64(pr.Total)
	}
	return
}

// Restart restarts the application
func Restart() error {
	selfPath, err := executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(selfPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CanUpdate checks if the current process has permission to replace the binary
func CanUpdate() bool {
	selfPath, err := executable()
	if err != nil {
		return false
	}

	// 1. Check if the file itself is writable
	f, err := os.OpenFile(selfPath, os.O_WRONLY, 0)
	if err == nil {
		logger.CloseAndLog(f, "self executable during update check")
		return true
	}

	// 2. If file open fails, check the parent directory
	// In many cases (like /usr/local/bin), we need write permission on the DIR
	// because we are deleting/renaming the file, not just writing to it.
	dir := filepath.Dir(selfPath)
	tmpFile, err := os.CreateTemp(dir, ".fm-up-test-*")
	if err != nil {
		return false
	}
	logger.CloseAndLog(tmpFile, "temporary check file during update check")
	logger.LogIfError(os.Remove(tmpFile.Name()), "update: failed to remove temporary check file")

	return true
}
