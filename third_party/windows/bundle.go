package windowsbundle

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed ffmpeg/current/ffmpeg.exe
//go:embed ffmpeg/current/ffprobe.exe
//go:embed ffmpeg-over-ip/current/ffmpeg-over-ip-client.exe
//go:embed ffmpeg-over-ip/current/ffmpeg-over-ip-server.exe
//go:embed ffmpeg-over-ip/current/ffmpeg.exe
//go:embed ffmpeg-over-ip/current/ffprobe.exe
var bundled embed.FS

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

func EnsureStudioTools() (StudioTools, error) {
	if runtime.GOOS != "windows" {
		return StudioTools{}, nil
	}
	root := os.TempDir()
	ffmpegPath, err := extract("ffmpeg/current/ffmpeg.exe", filepath.Join(root, "autosync_ffmpeg.exe"))
	if err != nil {
		return StudioTools{}, err
	}
	ffprobePath, err := extract("ffmpeg/current/ffprobe.exe", filepath.Join(root, "ffprobe.exe"))
	if err != nil {
		return StudioTools{}, err
	}
	clientPath, err := extract("ffmpeg-over-ip/current/ffmpeg-over-ip-client.exe", filepath.Join(root, "ffmpeg-over-ip.exe"))
	if err != nil {
		return StudioTools{}, err
	}
	return StudioTools{
		FFmpeg:     ffmpegPath,
		FFprobe:    ffprobePath,
		ClientPath: clientPath,
	}, nil
}

func EnsureServerTools() (ServerTools, error) {
	if runtime.GOOS != "windows" {
		return ServerTools{}, nil
	}
	root := os.TempDir()
	serverBinary, err := extract("ffmpeg-over-ip/current/ffmpeg-over-ip-server.exe", filepath.Join(root, "ffmpeg-over-ip-server.exe"))
	if err != nil {
		return ServerTools{}, err
	}
	ffmpegPath, err := extract("ffmpeg-over-ip/current/ffmpeg.exe", filepath.Join(root, "ffmpeg.exe"))
	if err != nil {
		return ServerTools{}, err
	}
	ffprobePath, err := extract("ffmpeg-over-ip/current/ffprobe.exe", filepath.Join(root, "ffprobe.exe"))
	if err != nil {
		return ServerTools{}, err
	}
	return ServerTools{
		ServerBinary: serverBinary,
		FFmpegPath:   ffmpegPath,
		FFprobePath:  ffprobePath,
	}, nil
}

func extract(src, dst string) (string, error) {
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
