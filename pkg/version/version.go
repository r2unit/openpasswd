package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Version information - updated by build process or releases
var (
	// Version is the current version of OpenPasswd (semver format: X.Y.Z)
	Version = "0.1.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = "dev"

	// BuildDate is when the binary was built (set during build)
	BuildDate = "unknown"

	// GitHubRepo is the GitHub repository for checking updates
	GitHubRepo = "r2unit/openpasswd"

	// GitHubAPIURL is the GitHub API endpoint for releases
	GitHubAPIURL = "https://api.github.com/repos/" + GitHubRepo + "/releases/latest"
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo returns the current version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("OpenPasswd v%s (commit: %s, built: %s, go: %s, platform: %s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdate checks GitHub for newer versions
func CheckForUpdate() (*GitHubRelease, bool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", GitHubAPIURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid rate limiting
	req.Header.Set("User-Agent", "OpenPasswd/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, false, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Skip draft and pre-release versions
	if release.Draft || release.Prerelease {
		return nil, false, nil
	}

	// Compare versions
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	updateAvailable, err := isNewer(latestVersion, Version)
	if err != nil {
		return nil, false, fmt.Errorf("failed to compare versions: %w", err)
	}

	return &release, updateAvailable, nil
}

// isNewer compares two semantic versions and returns true if v1 > v2
func isNewer(v1, v2 string) (bool, error) {
	// Parse v1
	major1, minor1, patch1, err := parseVersion(v1)
	if err != nil {
		return false, fmt.Errorf("invalid version v1 (%s): %w", v1, err)
	}

	// Parse v2
	major2, minor2, patch2, err := parseVersion(v2)
	if err != nil {
		return false, fmt.Errorf("invalid version v2 (%s): %w", v2, err)
	}

	// Compare major version
	if major1 > major2 {
		return true, nil
	} else if major1 < major2 {
		return false, nil
	}

	// Compare minor version
	if minor1 > minor2 {
		return true, nil
	} else if minor1 < minor2 {
		return false, nil
	}

	// Compare patch version
	return patch1 > patch2, nil
}

// parseVersion parses a semantic version string (X.Y.Z)
func parseVersion(v string) (major, minor, patch int, err error) {
	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")

	// Split into parts
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: expected X.Y.Z, got %s", v)
	}

	// Parse major
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	// Parse minor
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	// Parse patch (may contain additional info like -beta, -rc1)
	patchStr := parts[2]
	if idx := strings.IndexAny(patchStr, "-+"); idx != -1 {
		patchStr = patchStr[:idx]
	}
	patch, err = strconv.Atoi(patchStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// Upgrade downloads and installs the latest version
func Upgrade() error {
	// Check for updates first
	release, updateAvailable, err := CheckForUpdate()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !updateAvailable {
		return fmt.Errorf("already running the latest version (v%s)", Version)
	}

	fmt.Printf("Upgrading from v%s to v%s...\n", Version, release.TagName)

	// Detect platform
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var ext string
	if goos == "windows" {
		ext = ".exe"
	}

	// Construct binary name
	binaryName := fmt.Sprintf("openpasswd-%s-%s%s", goos, goarch, ext)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		GitHubRepo, release.TagName, binaryName)
	checksumURL := fmt.Sprintf("%s.sha256", downloadURL)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "openpasswd-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempBinary := filepath.Join(tempDir, "openpasswd"+ext)

	// Download binary
	fmt.Printf("Downloading %s...\n", binaryName)
	if err := downloadFile(downloadURL, tempBinary); err != nil {
		// If binary download fails, try install script method
		fmt.Println("Pre-built binary not available, running install script...")
		return upgradeViaInstallScript()
	}

	// Download and verify checksum if available
	checksumFile := tempBinary + ".sha256"
	if err := downloadFile(checksumURL, checksumFile); err == nil {
		if err := verifyChecksum(tempBinary, checksumFile); err != nil {
			fmt.Printf("Warning: checksum verification failed: %v\n", err)
		} else {
			fmt.Println("✓ Checksum verified")
		}
	}

	// Make binary executable
	if err := os.Chmod(tempBinary, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Replace binary (might need sudo)
	fmt.Printf("Installing to %s...\n", execPath)
	if err := replaceExecutable(tempBinary, execPath); err != nil {
		return fmt.Errorf("failed to replace executable: %w", err)
	}

	fmt.Printf("\n✓ Successfully upgraded to v%s!\n", release.TagName)
	fmt.Println("Restart your application to use the new version.")

	return nil
}

// upgradeViaInstallScript falls back to using the install script
func upgradeViaInstallScript() error {
	installURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/install.sh", GitHubRepo)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("curl -sSL %s | bash", installURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	return nil
}

// downloadFile downloads a file from URL to filepath
func downloadFile(url, filepath string) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// verifyChecksum verifies the SHA256 checksum of a file
func verifyChecksum(binaryPath, checksumPath string) error {
	// Read checksum file
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	// Parse checksum (format: "hash filename")
	parts := strings.Fields(string(checksumData))
	if len(parts) < 1 {
		return fmt.Errorf("invalid checksum format")
	}
	expectedHash := parts[0]

	// Calculate actual hash
	file, err := os.Open(binaryPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	// Compare
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

// replaceExecutable replaces the current executable with a new one
func replaceExecutable(newPath, targetPath string) error {
	// Try direct copy first
	if err := copyFile(newPath, targetPath); err != nil {
		// If permission denied, try with sudo
		if os.IsPermission(err) {
			fmt.Println("Permission denied, trying with sudo...")
			cmd := exec.Command("sudo", "cp", newPath, targetPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			return cmd.Run()
		}
		return err
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// GetLatestVersion returns the latest version from GitHub without full release info
func GetLatestVersion() (string, error) {
	release, _, err := CheckForUpdate()
	if err != nil {
		return "", err
	}
	if release == nil {
		return Version, nil
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}
