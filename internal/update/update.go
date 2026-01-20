package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"golang.org/x/mod/semver"
)

const (
	repoOwner = "zulfikawr"
	repoName  = "fm"
	timeout   = 5 * time.Second
)

var (
	githubAPI = "https://api.github.com/repos/%s/%s/releases/latest"
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	var targetAsset *Asset
	suffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		suffix += ".exe"
	}

	for _, asset := range release.Assets {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: %d", resp.StatusCode)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "fm-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Wrap response body for progress
	reader := &progressReader{
		Reader:   resp.Body,
		Total:    targetAsset.Size,
		OnUpdate: progress,
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// Set executable permissions
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	// Replace the current executable
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	// On Unix, we can just rename. On Windows, it's more complex.
	if runtime.GOOS == "windows" {
		oldPath := selfPath + ".old"
		_ = os.Remove(oldPath)
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
	defer s.Close()

	d, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
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
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(selfPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
