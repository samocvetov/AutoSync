package windowsbundle

import (
	"archive/zip"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"autosyncstudio/internal/bundles"
)

//go:embed ffmpeg/current/ffmpeg.exe
//go:embed ffmpeg/current/ffprobe.exe
//go:embed ffmpeg-over-ip/current/ffmpeg-over-ip-client.exe
//go:embed ffmpeg-over-ip/current/ffmpeg-over-ip-server.exe
//go:embed ffmpeg-over-ip/current/ffmpeg.exe
//go:embed ffmpeg-over-ip/current/ffprobe.exe
var bundled embed.FS

const (
	platformName              = "windows-amd64"
	clientAssetName           = "windows-amd64-ffmpeg-over-ip-client.zip"
	serverAssetName           = "windows-amd64-ffmpeg-over-ip-server.zip"
	clientBinaryName          = "ffmpeg-over-ip-client.exe"
	serverBinaryName          = "ffmpeg-over-ip-server.exe"
	ffmpegBinaryName          = "ffmpeg.exe"
	ffprobeBinaryName         = "ffprobe.exe"
	latestReleaseAPIURL       = "https://api.github.com/repos/steelbrain/ffmpeg-over-ip/releases/latest"
	httpUserAgent             = "AutoSyncStudio/managed-ffmpeg-over-ip"
	managedStateFileName      = "current.json"
	managedStateSourceBundled = "bundled"
	managedStateSourceRemote  = "remote"
)

type StudioTools struct {
	FFmpeg     string
	FFprobe    string
	ClientPath string
}

type ServerTools struct {
	ServerBinary string
	FFmpegPath   string
	FFprobePath  string
}

type FFmpegOverIPStatus struct {
	InstalledVersion string `json:"installedVersion"`
	AvailableVersion string `json:"availableVersion,omitempty"`
	UpdateAvailable  bool   `json:"updateAvailable"`
	ManagedRoot      string `json:"managedRoot"`
	Source           string `json:"source,omitempty"`
	LastCheckedAt    string `json:"lastCheckedAt,omitempty"`
	LastError        string `json:"lastError,omitempty"`
}

type managedState struct {
	CurrentVersion string `json:"currentVersion"`
	Source         string `json:"source"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

type githubRelease struct {
	TagName     string               `json:"tag_name"`
	PublishedAt string               `json:"published_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func EnsureStudioTools() (StudioTools, error) {
	if runtime.GOOS != "windows" {
		return StudioTools{}, nil
	}
	ffmpegPath, err := extractBundledFFmpeg("ffmpeg/current/ffmpeg.exe", filepath.Join(managedToolsRoot(), "ffmpeg", ffmpegBinaryName))
	if err != nil {
		return StudioTools{}, err
	}
	ffprobePath, err := extractBundledFFmpeg("ffmpeg/current/ffprobe.exe", filepath.Join(managedToolsRoot(), "ffmpeg", ffprobeBinaryName))
	if err != nil {
		return StudioTools{}, err
	}
	release, err := ensureActiveFFmpegOverIPRelease()
	if err != nil {
		return StudioTools{}, err
	}
	return StudioTools{
		FFmpeg:     ffmpegPath,
		FFprobe:    ffprobePath,
		ClientPath: filepath.Join(release, clientBinaryName),
	}, nil
}

func EnsureServerTools() (ServerTools, error) {
	if runtime.GOOS != "windows" {
		return ServerTools{}, nil
	}
	release, err := ensureActiveFFmpegOverIPRelease()
	if err != nil {
		return ServerTools{}, err
	}
	return ServerTools{
		ServerBinary: filepath.Join(release, serverBinaryName),
		FFmpegPath:   filepath.Join(release, ffmpegBinaryName),
		FFprobePath:  filepath.Join(release, ffprobeBinaryName),
	}, nil
}

func GetFFmpegOverIPStatus(ctx context.Context, includeLatest bool) FFmpegOverIPStatus {
	status := FFmpegOverIPStatus{
		InstalledVersion: bundledReleaseVersion(),
		ManagedRoot:      ffmpegOverIPManagedRoot(),
		Source:           managedStateSourceBundled,
	}
	if runtime.GOOS != "windows" {
		return status
	}

	if release, err := ensureActiveFFmpegOverIPRelease(); err != nil {
		status.LastError = err.Error()
	} else {
		status.ManagedRoot = release
	}

	if state, err := readManagedState(); err == nil {
		if state.CurrentVersion != "" {
			status.InstalledVersion = state.CurrentVersion
		}
		if state.Source != "" {
			status.Source = state.Source
		}
		if state.UpdatedAt != "" {
			status.LastCheckedAt = state.UpdatedAt
		}
	}

	if !includeLatest {
		return status
	}

	release, err := fetchLatestRelease(ctx)
	if err != nil {
		status.LastError = err.Error()
		return status
	}
	status.AvailableVersion = release.TagName
	status.UpdateAvailable = normalizeVersion(release.TagName) != normalizeVersion(status.InstalledVersion)
	status.LastCheckedAt = time.Now().Format(time.RFC3339)
	return status
}

func UpdateFFmpegOverIP(ctx context.Context) (FFmpegOverIPStatus, error) {
	status := GetFFmpegOverIPStatus(ctx, false)
	if runtime.GOOS != "windows" {
		return status, nil
	}

	release, err := fetchLatestRelease(ctx)
	if err != nil {
		status.LastError = err.Error()
		return status, err
	}

	releaseDir := ffmpegOverIPReleaseDir(release.TagName)
	if !releaseFilesExist(releaseDir) {
		if err := os.MkdirAll(releaseDir, 0755); err != nil {
			return status, err
		}
		clientAsset, err := findReleaseAsset(release, clientAssetName)
		if err != nil {
			return status, err
		}
		serverAsset, err := findReleaseAsset(release, serverAssetName)
		if err != nil {
			return status, err
		}
		if err := downloadAndExtractReleaseZip(ctx, clientAsset.BrowserDownloadURL, releaseDir, []string{clientBinaryName}); err != nil {
			return status, err
		}
		if err := downloadAndExtractReleaseZip(ctx, serverAsset.BrowserDownloadURL, releaseDir, []string{serverBinaryName, ffmpegBinaryName, ffprobeBinaryName}); err != nil {
			return status, err
		}
	}

	if err := writeManagedState(managedState{
		CurrentVersion: release.TagName,
		Source:         managedStateSourceRemote,
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}); err != nil {
		return status, err
	}

	return GetFFmpegOverIPStatus(ctx, true), nil
}

func extractBundledFFmpeg(src, dst string) (string, error) {
	data, err := fs.ReadFile(bundled, src)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", err
	}
	if existing, err := os.ReadFile(dst); err == nil && len(existing) == len(data) {
		return dst, nil
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return "", err
	}
	return dst, nil
}

func ensureActiveFFmpegOverIPRelease() (string, error) {
	version := bundledReleaseVersion()
	source := managedStateSourceBundled

	if state, err := readManagedState(); err == nil {
		if strings.TrimSpace(state.CurrentVersion) != "" {
			version = strings.TrimSpace(state.CurrentVersion)
		}
		if strings.TrimSpace(state.Source) != "" {
			source = strings.TrimSpace(state.Source)
		}
	}

	releaseDir := ffmpegOverIPReleaseDir(version)
	if releaseFilesExist(releaseDir) {
		return releaseDir, nil
	}

	bundledDir, err := ensureBundledFFmpegOverIPRelease()
	if err != nil {
		return "", err
	}
	if normalizeVersion(version) != normalizeVersion(bundledReleaseVersion()) || source != managedStateSourceBundled {
		_ = writeManagedState(managedState{
			CurrentVersion: bundledReleaseVersion(),
			Source:         managedStateSourceBundled,
			UpdatedAt:      time.Now().Format(time.RFC3339),
		})
	}
	return bundledDir, nil
}

func ensureBundledFFmpegOverIPRelease() (string, error) {
	version := bundledReleaseVersion()
	releaseDir := ffmpegOverIPReleaseDir(version)
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return "", err
	}
	targets := map[string]string{
		"ffmpeg-over-ip/current/ffmpeg-over-ip-client.exe": filepath.Join(releaseDir, clientBinaryName),
		"ffmpeg-over-ip/current/ffmpeg-over-ip-server.exe": filepath.Join(releaseDir, serverBinaryName),
		"ffmpeg-over-ip/current/ffmpeg.exe":                filepath.Join(releaseDir, ffmpegBinaryName),
		"ffmpeg-over-ip/current/ffprobe.exe":               filepath.Join(releaseDir, ffprobeBinaryName),
	}
	for src, dst := range targets {
		data, err := fs.ReadFile(bundled, src)
		if err != nil {
			return "", err
		}
		if existing, err := os.ReadFile(dst); err == nil && len(existing) == len(data) {
			continue
		}
		if err := os.WriteFile(dst, data, 0755); err != nil {
			return "", err
		}
	}
	if _, err := readManagedState(); err != nil {
		_ = writeManagedState(managedState{
			CurrentVersion: version,
			Source:         managedStateSourceBundled,
			UpdatedAt:      time.Now().Format(time.RFC3339),
		})
	}
	return releaseDir, nil
}

func fetchLatestRelease(ctx context.Context) (githubRelease, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPIURL, nil)
	if err != nil {
		return githubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", httpUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("latest release check failed: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, err
	}
	if strings.TrimSpace(release.TagName) == "" {
		return githubRelease{}, errors.New("latest release tag is missing")
	}
	return release, nil
}

func downloadAndExtractReleaseZip(ctx context.Context, assetURL, destinationDir string, requiredEntries []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", httpUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(destinationDir), "*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	found := map[string]bool{}
	for _, file := range zipReader.File {
		name := filepath.Base(file.Name)
		if !containsString(requiredEntries, name) {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destinationDir, name)
		targetFile, err := os.Create(targetPath)
		if err != nil {
			_ = rc.Close()
			return err
		}
		if _, err := io.Copy(targetFile, rc); err != nil {
			_ = targetFile.Close()
			_ = rc.Close()
			return err
		}
		if err := targetFile.Close(); err != nil {
			_ = rc.Close()
			return err
		}
		_ = rc.Close()
		if err := os.Chmod(targetPath, 0755); err != nil {
			return err
		}
		found[name] = true
	}

	for _, required := range requiredEntries {
		if !found[required] {
			return fmt.Errorf("release archive is missing %s", required)
		}
	}
	return nil
}

func findReleaseAsset(release githubRelease, assetName string) (githubReleaseAsset, error) {
	for _, asset := range release.Assets {
		if strings.EqualFold(strings.TrimSpace(asset.Name), assetName) {
			return asset, nil
		}
	}
	return githubReleaseAsset{}, fmt.Errorf("release asset %s not found", assetName)
}

func releaseFilesExist(releaseDir string) bool {
	required := []string{
		filepath.Join(releaseDir, clientBinaryName),
		filepath.Join(releaseDir, serverBinaryName),
		filepath.Join(releaseDir, ffmpegBinaryName),
		filepath.Join(releaseDir, ffprobeBinaryName),
	}
	for _, path := range required {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return false
		}
	}
	return true
}

func readManagedState() (managedState, error) {
	data, err := os.ReadFile(ffmpegOverIPStatePath())
	if err != nil {
		return managedState{}, err
	}
	var state managedState
	if err := json.Unmarshal(data, &state); err != nil {
		return managedState{}, err
	}
	return state, nil
}

func writeManagedState(state managedState) error {
	if err := os.MkdirAll(ffmpegOverIPManagedRoot(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ffmpegOverIPStatePath(), data, 0600)
}

func ffmpegOverIPManagedRoot() string {
	return filepath.Join(managedToolsRoot(), "ffmpeg-over-ip")
}

func ffmpegOverIPReleaseDir(version string) string {
	cleanVersion := strings.TrimSpace(version)
	if cleanVersion == "" {
		cleanVersion = bundledReleaseVersion()
	}
	return filepath.Join(ffmpegOverIPManagedRoot(), "releases", cleanVersion)
}

func ffmpegOverIPStatePath() string {
	return filepath.Join(ffmpegOverIPManagedRoot(), managedStateFileName)
}

func managedToolsRoot() string {
	return filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "tools", platformName)
}

func runtimeWorkspaceRoot() string {
	exePath, err := os.Executable()
	if err == nil && exePath != "" {
		return filepath.Dir(exePath)
	}
	wd, err := os.Getwd()
	if err == nil && wd != "" {
		return wd
	}
	return "."
}

func bundledReleaseVersion() string {
	for _, component := range bundles.ComponentsForPlatform(platformName) {
		if component.Name == "ffmpegOverIPClient" && strings.TrimSpace(component.Version) != "" {
			return strings.TrimSpace(component.Version)
		}
	}
	return "v5.0.0"
}

func normalizeVersion(version string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.ToLower(version), "v"))
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
