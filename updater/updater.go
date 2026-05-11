package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

const (
	apiURL  = "https://api.github.com/repos/FutureSolutionDev/internet-monitor/releases/latest"
	checkInterval = 6 * time.Hour
)

// Info holds the result of a version check.
type Info struct {
	HasUpdate    bool   `json:"has_update"`
	LatestVersion string `json:"latest_version"`
	CurrentVersion string `json:"current_version"`
	DownloadURL  string `json:"download_url"`
	ReleaseNotes string `json:"release_notes"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Body    string    `json:"body"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Check queries GitHub and returns update info.
func Check(currentVersion string) (*Info, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "internet-monitor/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	info := &Info{
		HasUpdate:     compareVersions(rel.TagName, currentVersion) > 0,
		LatestVersion: rel.TagName,
		CurrentVersion: currentVersion,
		ReleaseNotes:  rel.Body,
	}

	if info.HasUpdate {
		info.DownloadURL = pickAsset(rel.Assets)
	}

	return info, nil
}

// Apply downloads the binary at downloadURL and atomically replaces the current exe.
// Returns nil on success; the caller should restart the process.
func Apply(downloadURL string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	return selfupdate.Apply(resp.Body, selfupdate.Options{})
}

// Restart launches the updated binary and exits the current process.
func Restart() {
	exe, err := os.Executable()
	if err != nil {
		os.Exit(0)
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Dir = filepath.Dir(exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	os.Exit(0)
}

// pickAsset selects the correct release asset for the current OS/arch/binary type.
func pickAsset(assets []ghAsset) string {
	// Detect if current binary is the GUI version
	exe, _ := os.Executable()
	isGUI := strings.Contains(strings.ToLower(filepath.Base(exe)), "gui")

	var goos   = runtime.GOOS
	var goarch = runtime.GOARCH

	// Build expected name fragments
	var patterns []string
	switch goos {
	case "windows":
		if isGUI {
			patterns = []string{"gui-windows", "gui_windows"}
		} else {
			patterns = []string{"internet-monitor-windows"}
		}
	case "darwin":
		switch goarch {
		case "arm64":
			if isGUI {
				patterns = []string{"gui-macos-arm64"}
			} else {
				patterns = []string{"macos-arm64"}
			}
		default:
			if isGUI {
				patterns = []string{"gui-macos-intel"}
			} else {
				patterns = []string{"macos-intel"}
			}
		}
	default: // linux
		if isGUI {
			patterns = []string{"gui-linux"}
		} else {
			patterns = []string{"internet-monitor-linux"}
		}
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		for _, pat := range patterns {
			if strings.Contains(name, pat) {
				return asset.BrowserDownloadURL
			}
		}
	}
	return ""
}

// compareVersions returns 1 if a > b, -1 if a < b, 0 if equal.
// Handles "v" prefix (e.g. "v1.2.0" vs "1.2.0").
func compareVersions(a, b string) int {
	a = strings.TrimPrefix(strings.TrimSpace(a), "v")
	b = strings.TrimPrefix(strings.TrimSpace(b), "v")

	if a == b || b == "dev" || b == "" {
		return 0
	}

	partsA := strings.SplitN(a, ".", 3)
	partsB := strings.SplitN(b, ".", 3)

	for i := 0; i < 3; i++ {
		var pa, pb int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &pa)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &pb)
		}
		if pa > pb {
			return 1
		}
		if pa < pb {
			return -1
		}
	}
	return 0
}
