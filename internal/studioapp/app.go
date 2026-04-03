package studioapp

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"autosyncstudio/internal/appmeta"
	"autosyncstudio/internal/bundles"
	windowsbundle "autosyncstudio/third_party/windows"
)

//go:embed index.html main.js
var staticFiles embed.FS

//go:embed fonts/*
var embeddedFontFiles embed.FS

const (
	defaultAddr           = "127.0.0.1:8421"
	defaultSampleRate     = 16000
	defaultAnalyzeSeconds = 180.0
	defaultMaxLagSeconds  = 12.0
	coarseWindowSamples   = 320
	fineWindowSamples     = 80
)

type App struct {
	addr        string
	ffmpegPath  string
	ffprobePath string
	mu          sync.Mutex
	currentCmd  *exec.Cmd
	currentTask context.CancelFunc
}

type progressEvent struct {
	Percent          float64               `json:"percent,omitempty"`
	Message          string                `json:"message,omitempty"`
	Done             bool                  `json:"done,omitempty"`
	Error            string                `json:"error,omitempty"`
	OutputPath       string                `json:"outputPath,omitempty"`
	Duration         string                `json:"duration,omitempty"`
	TranscriptSource string                `json:"transcriptSource,omitempty"`
	SRTPath          string                `json:"srtPath,omitempty"`
	TextPath         string                `json:"textPath,omitempty"`
	ASSPath          string                `json:"assPath,omitempty"`
	Command          string                `json:"command,omitempty"`
	Shots            []multicamShotSummary `json:"shots,omitempty"`
	TotalTime        float64               `json:"totalTime,omitempty"`
	PlanPath         string                `json:"planPath,omitempty"`
	Files            []shortsRenderedFile  `json:"files,omitempty"`
	Failed           []string              `json:"failed,omitempty"`
	Rendered         int                   `json:"rendered,omitempty"`
	ShortsPlan       *shortsPlanResponse   `json:"shortsPlan,omitempty"`
}

type apiError struct {
	Error string `json:"error"`
}

type pickerRequest struct {
	Kind string `json:"kind"`
	Path string `json:"path,omitempty"`
}

type pickerResponse struct {
	Path string `json:"path"`
}

type pathExistsRequest struct {
	Path string `json:"path"`
}

type pathExistsResponse struct {
	Exists bool `json:"exists"`
}

type appSettings struct {
	AssemblyAIKey string `json:"assemblyAiKey,omitempty"`
	GeminiAIKey   string `json:"geminiAiKey,omitempty"`
	OpenAIKey     string `json:"openAiKey,omitempty"`
	AIKey         string `json:"aiKey,omitempty"`
}

type systemInfoResponse struct {
	Name              string                           `json:"name"`
	Version           string                           `json:"version"`
	Address           string                           `json:"address"`
	FFmpegPath        string                           `json:"ffmpegPath"`
	FFprobePath       string                           `json:"ffprobePath"`
	BundledPlatform   string                           `json:"bundledPlatform"`
	BundledComponents []bundles.NamedComponent         `json:"bundledComponents"`
	RemoteTools       windowsbundle.FFmpegOverIPStatus `json:"remoteTools"`
}

type backendStatusRequest struct {
	ExecutionMode    string `json:"executionMode"`
	RemoteAddress    string `json:"remoteAddress"`
	RemoteSecret     string `json:"remoteSecret"`
	RemoteClientPath string `json:"remoteClientPath"`
}

type backendStatusResponse struct {
	Mode            string `json:"mode"`
	OverallStatus   string `json:"overallStatus"`
	ModeLabel       string `json:"modeLabel"`
	ClientStatus    string `json:"clientStatus"`
	ServerStatus    string `json:"serverStatus"`
	BackendReady    bool   `json:"backendReady"`
	ServerReachable bool   `json:"serverReachable"`
	ClientFound     bool   `json:"clientFound"`
	ResolvedClient  string `json:"resolvedClient"`
	ResolvedAddress string `json:"resolvedAddress"`
	Message         string `json:"message"`
}

type syncAnalyzeRequest struct {
	VideoPath      string  `json:"videoPath"`
	AudioPath      string  `json:"audioPath"`
	AnalyzeSeconds float64 `json:"analyzeSeconds"`
	MaxLagSeconds  float64 `json:"maxLagSeconds"`
}

type syncAnalyzeResponse struct {
	DelaySeconds   float64 `json:"delaySeconds"`
	DelayMs        int     `json:"delayMs"`
	Confidence     float64 `json:"confidence"`
	VideoDuration  float64 `json:"videoDuration"`
	AudioDuration  float64 `json:"audioDuration"`
	Recommendation string  `json:"recommendation"`
	RenderSummary  string  `json:"renderSummary"`
}

type syncRenderRequest struct {
	VideoPath        string  `json:"videoPath"`
	AudioPath        string  `json:"audioPath"`
	OutputPath       string  `json:"outputPath"`
	PreviewSeconds   float64 `json:"previewSeconds"`
	DelaySeconds     float64 `json:"delaySeconds"`
	CRF              int     `json:"crf"`
	Preset           string  `json:"preset"`
	ExecutionMode    string  `json:"executionMode"`
	RemoteAddress    string  `json:"remoteAddress"`
	RemoteSecret     string  `json:"remoteSecret"`
	RemoteClientPath string  `json:"remoteClientPath"`
}

type syncRenderResponse struct {
	OutputPath string `json:"outputPath"`
	Duration   string `json:"duration"`
	Command    string `json:"command"`
}

type multicamAnalyzeRequest struct {
	MasterAudioPath string   `json:"masterAudioPath"`
	CameraPaths     []string `json:"cameraPaths"`
	AnalyzeSeconds  float64  `json:"analyzeSeconds"`
	MaxLagSeconds   float64  `json:"maxLagSeconds"`
}

type multicamCameraResult struct {
	Path           string  `json:"path"`
	DelaySeconds   float64 `json:"delaySeconds"`
	DelayMs        int     `json:"delayMs"`
	Confidence     float64 `json:"confidence"`
	Duration       float64 `json:"duration"`
	Recommendation string  `json:"recommendation"`
}

type multicamAnalyzeResponse struct {
	MasterAudioPath string                 `json:"masterAudioPath"`
	Cameras         []multicamCameraResult `json:"cameras"`
}

type multicamExportRequest struct {
	MasterAudioPath  string                 `json:"masterAudioPath"`
	CameraPaths      []string               `json:"cameraPaths"`
	AnalyzeSeconds   float64                `json:"analyzeSeconds"`
	MaxLagSeconds    float64                `json:"maxLagSeconds"`
	MeasuredCameras  []multicamCameraResult `json:"measuredCameras,omitempty"`
	OutputDir        string                 `json:"outputDir"`
	CRF              int                    `json:"crf"`
	Preset           string                 `json:"preset"`
	ExecutionMode    string                 `json:"executionMode"`
	RemoteAddress    string                 `json:"remoteAddress"`
	RemoteSecret     string                 `json:"remoteSecret"`
	RemoteClientPath string                 `json:"remoteClientPath"`
}

type multicamExportPlan struct {
	Path         string  `json:"path"`
	DelaySeconds float64 `json:"delaySeconds"`
	DelayMs      int     `json:"delayMs"`
	Confidence   float64 `json:"confidence"`
	OutputPath   string  `json:"outputPath"`
	Strategy     string  `json:"strategy"`
	Command      string  `json:"command"`
}

type multicamExportResponse struct {
	MasterAudioPath string               `json:"masterAudioPath"`
	OutputDir       string               `json:"outputDir"`
	Plans           []multicamExportPlan `json:"plans"`
	Note            string               `json:"note"`
}

type multicamRenderRequest struct {
	MasterAudioPath   string                 `json:"masterAudioPath"`
	CameraPaths       []string               `json:"cameraPaths"`
	AnalyzeSeconds    float64                `json:"analyzeSeconds"`
	MaxLagSeconds     float64                `json:"maxLagSeconds"`
	MeasuredCameras   []multicamCameraResult `json:"measuredCameras,omitempty"`
	OutputPath        string                 `json:"outputPath"`
	PreviewSeconds    float64                `json:"previewSeconds"`
	CRF               int                    `json:"crf"`
	Preset            string                 `json:"preset"`
	ExecutionMode     string                 `json:"executionMode"`
	RemoteAddress     string                 `json:"remoteAddress"`
	RemoteSecret      string                 `json:"remoteSecret"`
	RemoteClientPath  string                 `json:"remoteClientPath"`
	ShotWindowSeconds float64                `json:"shotWindowSeconds"`
	MinShotSeconds    float64                `json:"minShotSeconds"`
	PrimaryCamera     int                    `json:"primaryCamera"`
	EditMode          string                 `json:"editMode"`
	AssemblyAIKey     string                 `json:"assemblyAiKey"`
	AIProvider        string                 `json:"aiProvider"`
	AIKey             string                 `json:"aiKey"`
	AIPrompt          string                 `json:"aiPrompt"`
}

type multicamRenderResponse struct {
	OutputPath   string                `json:"outputPath"`
	Duration     string                `json:"duration"`
	Command      string                `json:"command"`
	Shots        []multicamShotSummary `json:"shots"`
	TotalSeconds float64               `json:"totalSeconds"`
}

type shortsPlanRequest struct {
	VideoPath       string   `json:"videoPath"`
	AudioPath       string   `json:"audioPath,omitempty"`
	AnalyzeSeconds  float64  `json:"analyzeSeconds,omitempty"`
	MaxLagSeconds   float64  `json:"maxLagSeconds,omitempty"`
	AssemblyAIKey   string   `json:"assemblyAiKey"`
	AIProvider      string   `json:"aiProvider"`
	AIKey           string   `json:"aiKey"`
	AIPrompt        string   `json:"aiPrompt"`
	ShortsCount     int      `json:"shortsCount"`
	MasterAudioPath string   `json:"masterAudioPath,omitempty"`
	CameraPaths     []string `json:"cameraPaths,omitempty"`
	PrimaryCamera   int      `json:"primaryCamera,omitempty"`
}

type shortSegment struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
	Duration       float64 `json:"duration"`
	Reason         string  `json:"reason"`
	CameraHint     int     `json:"cameraHint,omitempty"`
	Enabled        bool    `json:"enabled"`
	PreviewCommand string  `json:"previewCommand"`
	Command        string  `json:"command,omitempty"`
}

type shortsPlanResponse struct {
	Provider         string              `json:"provider"`
	Segments         []shortSegment      `json:"segments"`
	Utterances       []AssemblyUtterance `json:"utterances,omitempty"`
	Note             string              `json:"note"`
	TimelineSource   string              `json:"timelineSource,omitempty"`
	SyncDelaySeconds float64             `json:"syncDelaySeconds,omitempty"`
	TimelineDuration float64             `json:"timelineDuration,omitempty"`
}

type shortsRenderRequest struct {
	VideoPath         string              `json:"videoPath"`
	AudioPath         string              `json:"audioPath,omitempty"`
	OutputDir         string              `json:"outputDir"`
	Segments          []shortSegment      `json:"segments"`
	Utterances        []AssemblyUtterance `json:"utterances,omitempty"`
	Formats           []string            `json:"formats"`
	CaptionsMode      string              `json:"captionsMode"`
	SubtitleFont      string              `json:"subtitleFont,omitempty"`
	SubtitleBgColor   string              `json:"subtitleBgColor,omitempty"`
	SubtitleBgOpacity int                 `json:"subtitleBgOpacity,omitempty"`
	SyncDelaySeconds  float64             `json:"syncDelaySeconds,omitempty"`
	CRF               int                 `json:"crf"`
	Preset            string              `json:"preset"`
	ExecutionMode     string              `json:"executionMode"`
	RemoteAddress     string              `json:"remoteAddress"`
	RemoteSecret      string              `json:"remoteSecret"`
	RemoteClientPath  string              `json:"remoteClientPath"`
}

type fullCaptionsRenderRequest struct {
	VideoPath         string `json:"videoPath"`
	AudioPath         string `json:"audioPath,omitempty"`
	OutputDir         string `json:"outputDir"`
	CaptionsMode      string `json:"captionsMode"`
	AssemblyAIKey     string `json:"assemblyAiKey"`
	SubtitleFont      string `json:"subtitleFont,omitempty"`
	SubtitleBgColor   string `json:"subtitleBgColor,omitempty"`
	SubtitleBgOpacity int    `json:"subtitleBgOpacity,omitempty"`
	CRF               int    `json:"crf"`
	Preset            string `json:"preset"`
	ExecutionMode     string `json:"executionMode"`
	RemoteAddress     string `json:"remoteAddress"`
	RemoteSecret      string `json:"remoteSecret"`
	RemoteClientPath  string `json:"remoteClientPath"`
}

type fullCaptionsRenderResponse struct {
	OutputPath       string `json:"outputPath"`
	Duration         string `json:"duration"`
	TranscriptSource string `json:"transcriptSource"`
	SRTPath          string `json:"srtPath,omitempty"`
	TextPath         string `json:"textPath,omitempty"`
	ASSPath          string `json:"assPath,omitempty"`
}

type shortsRenderedFile struct {
	SegmentID string  `json:"segmentId"`
	Title     string  `json:"title"`
	Format    string  `json:"format"`
	Output    string  `json:"output"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
}

type shortsRenderResponse struct {
	OutputDir     string               `json:"outputDir"`
	PlanPath      string               `json:"planPath"`
	Files         []shortsRenderedFile `json:"files"`
	Duration      string               `json:"duration"`
	RenderedCount int                  `json:"renderedCount"`
	Failed        []string             `json:"failed,omitempty"`
}

type shortRenderPreset struct {
	ID         string
	FileSuffix string
	Width      int
	Height     int
}

type multicamShotSummary struct {
	CameraIndex int     `json:"cameraIndex"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
}

type syncMetrics struct {
	DelaySeconds  float64
	Confidence    float64
	VideoDuration float64
	AudioDuration float64
}

type executionPlan struct {
	Mode       string
	Executable string
	PrefixArgs []string
	Cleanup    func()
}

type videoStreamMeta struct {
	Width    int
	Height   int
	FPS      float64
	Duration float64
	Rotation float64
}

type multicamAnalysis struct {
	Path     string
	Metrics  syncMetrics
	Envelope []float64
	Meta     videoStreamMeta
}

type shotSegment struct {
	CameraIndex int
	Start       float64
	End         float64
}

type AssemblyUploadRes struct {
	UploadURL string `json:"upload_url"`
	Error     string `json:"error"`
}

type AssemblyTranscriptRes struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	Error         string  `json:"error"`
	AudioDuration float64 `json:"audio_duration"`
}

type AssemblyUtterance struct {
	Speaker string         `json:"speaker"`
	Start   int            `json:"start"`
	End     int            `json:"end"`
	Text    string         `json:"text"`
	Words   []AssemblyWord `json:"words,omitempty"`
}

type AssemblyWord struct {
	Text    string `json:"text"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Speaker string `json:"speaker,omitempty"`
}

type AssemblyPollRes struct {
	Status        string              `json:"status"`
	Error         string              `json:"error"`
	Utterances    []AssemblyUtterance `json:"utterances"`
	AudioDuration float64             `json:"audio_duration"`
}

func NewApp() *App {
	return NewAppWithAddr(defaultAddr)
}

func NewAppWithAddr(addr string) *App {
	ffmpegPath := findBinary("ffmpeg")
	ffprobePath := findBinary("ffprobe")
	if runtime.GOOS == "windows" {
		if tools, err := windowsbundle.EnsureStudioTools(); err == nil {
			if tools.FFmpeg != "" {
				ffmpegPath = tools.FFmpeg
			}
			if tools.FFprobe != "" {
				ffprobePath = tools.FFprobe
			}
		}
	}
	return &App{
		addr:        addr,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

func findBinary(name string) string {
	exeDir := runtimeWorkspaceRoot()
	candidates := []string{filepath.Join(exeDir, name)}
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		candidates = append([]string{filepath.Join(exeDir, name+".exe")}, candidates...)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func newCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	applyWindowsCommandAttrs(cmd)
	return cmd
}

func (a *App) Run() error {
	ln, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}
	a.addr = ln.Addr().String()
	return a.RunListener(ln)
}

func (a *App) RunListener(ln net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/main.js", a.handleMainJS)
	mux.HandleFunc("/api/system", a.handleSystem)
	mux.HandleFunc("/api/ffmpeg-over-ip-tools", a.handleFFmpegOverIPTools)
	mux.HandleFunc("/api/update-ffmpeg-over-ip-tools", a.handleUpdateFFmpegOverIPTools)
	mux.HandleFunc("/api/backend-status", a.handleBackendStatus)
	mux.HandleFunc("/api/pick-file", a.handlePickFile)
	mux.HandleFunc("/api/pick-directory", a.handlePickDirectory)
	mux.HandleFunc("/api/pick-save", a.handlePickSave)
	mux.HandleFunc("/api/path-exists", a.handlePathExists)
	mux.HandleFunc("/api/settings", a.handleSettings)
	mux.HandleFunc("/api/analyze-sync", a.handleAnalyzeSync)
	mux.HandleFunc("/api/render-sync", a.handleRenderSync)
	mux.HandleFunc("/api/render-sync-stream", a.handleRenderSyncStream)
	mux.HandleFunc("/api/cancel", a.handleCancel)
	mux.HandleFunc("/api/analyze-multicam", a.handleAnalyzeMulticam)
	mux.HandleFunc("/api/export-multicam-plan", a.handleExportMulticamPlan)
	mux.HandleFunc("/api/render-multicam", a.handleRenderMulticam)
	mux.HandleFunc("/api/render-multicam-stream", a.handleRenderMulticamStream)
	mux.HandleFunc("/api/plan-shorts", a.handlePlanShorts)
	mux.HandleFunc("/api/plan-shorts-stream", a.handlePlanShortsStream)
	mux.HandleFunc("/api/render-shorts", a.handleRenderShorts)
	mux.HandleFunc("/api/render-shorts-stream", a.handleRenderShortsStream)
	mux.HandleFunc("/api/render-full-captions", a.handleRenderFullCaptions)
	mux.HandleFunc("/api/render-full-captions-stream", a.handleRenderFullCaptionsStream)

	log.Printf("AutoSync Studio is ready at http://%s\n", a.addr)
	return http.Serve(ln, mux)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	a.serveEmbeddedFile(w, "index.html", "text/html; charset=utf-8")
}

func (a *App) handleMainJS(w http.ResponseWriter, r *http.Request) {
	a.serveEmbeddedFile(w, "main.js", "application/javascript; charset=utf-8")
}

func (a *App) serveEmbeddedFile(w http.ResponseWriter, name, contentType string) {
	data, err := fs.ReadFile(staticFiles, name)
	if err != nil {
		http.Error(w, "embedded asset missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func (a *App) handleSystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.writeJSON(w, http.StatusOK, systemInfoResponse{
		Name:              appmeta.Name,
		Version:           appmeta.Version,
		Address:           a.addr,
		FFmpegPath:        a.ffmpegPath,
		FFprobePath:       a.ffprobePath,
		BundledPlatform:   "windows-amd64",
		BundledComponents: bundles.ComponentsForPlatform("windows-amd64"),
		RemoteTools:       windowsbundle.GetFFmpegOverIPStatus(r.Context(), false),
	})
}

func (a *App) handleFFmpegOverIPTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.writeJSON(w, http.StatusOK, windowsbundle.GetFFmpegOverIPStatus(r.Context(), true))
}

func (a *App) handleUpdateFFmpegOverIPTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := windowsbundle.UpdateFFmpegOverIP(r.Context())
	if err != nil {
		a.writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, status)
}

func (a *App) handleBackendStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req backendStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	status := a.inspectBackendStatus(req)
	a.writeJSON(w, http.StatusOK, status)
}

func (a *App) handlePickFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if runtime.GOOS != "windows" {
		a.writeError(w, http.StatusNotImplemented, "native picker is only implemented for Windows builds")
		return
	}

	var req pickerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	path, err := windowsPickFile(req.Kind)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, pickerResponse{Path: path})
}

func (a *App) handlePickDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if runtime.GOOS != "windows" {
		a.writeError(w, http.StatusNotImplemented, "native picker is only implemented for Windows builds")
		return
	}

	path, err := windowsPickDirectory()
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, pickerResponse{Path: path})
}

func (a *App) handlePickSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if runtime.GOOS != "windows" {
		a.writeError(w, http.StatusNotImplemented, "native picker is only implemented for Windows builds")
		return
	}

	var req pickerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	path, err := windowsPickSave(req.Kind, req.Path)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, pickerResponse{Path: path})
}

func (a *App) handlePathExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req pathExistsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	path := strings.TrimSpace(req.Path)
	if path == "" {
		a.writeJSON(w, http.StatusOK, pathExistsResponse{Exists: false})
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		a.writeJSON(w, http.StatusOK, pathExistsResponse{Exists: false})
		return
	}
	a.writeJSON(w, http.StatusOK, pathExistsResponse{Exists: true})
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := loadAppSettings()
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.writeJSON(w, http.StatusOK, settings)
	case http.MethodPost:
		var req appSettings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			a.writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		req.AssemblyAIKey = strings.TrimSpace(req.AssemblyAIKey)
		req.GeminiAIKey = strings.TrimSpace(req.GeminiAIKey)
		req.OpenAIKey = strings.TrimSpace(req.OpenAIKey)
		req.AIKey = ""
		if err := saveAppSettings(req); err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.writeJSON(w, http.StatusOK, req)
	default:
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAnalyzeSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req syncAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics, err := a.analyzeSync(req.VideoPath, req.AudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := syncAnalyzeResponse{
		DelaySeconds:   round(metrics.DelaySeconds, 3),
		DelayMs:        int(math.Round(metrics.DelaySeconds * 1000)),
		Confidence:     round(metrics.Confidence, 3),
		VideoDuration:  round(metrics.VideoDuration, 2),
		AudioDuration:  round(metrics.AudioDuration, 2),
		Recommendation: describeDelay(metrics.DelaySeconds),
		RenderSummary:  buildRenderSummary(metrics.DelaySeconds),
	}
	a.writeJSON(w, http.StatusOK, response)
}

func (a *App) handleRenderSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req syncRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	outPath, cmdString, elapsed, err := a.renderSyncedFile(req, nil)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.writeJSON(w, http.StatusOK, syncRenderResponse{
		OutputPath: outPath,
		Duration:   elapsed.String(),
		Command:    cmdString,
	})
}

func (a *App) handleRenderSyncStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req syncRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	streamJSON(w, func(send func(progressEvent)) {
		outPath, cmdString, elapsed, err := a.renderSyncedFile(req, send)
		if err != nil {
			send(progressEvent{Error: err.Error()})
			return
		}
		send(progressEvent{
			Done:       true,
			OutputPath: outPath,
			Duration:   elapsed.String(),
			Command:    cmdString,
			Message:    "render complete",
		})
	})
}

func (a *App) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.mu.Lock()
	cmd := a.currentCmd
	cancel := a.currentTask
	a.mu.Unlock()
	if cancel == nil && (cmd == nil || cmd.Process == nil) {
		a.writeJSON(w, http.StatusOK, map[string]string{"status": "idle"})
		return
	}
	if cancel != nil {
		cancel()
	}
	if cmd != nil && cmd.Process != nil {
		if err := killCommandTree(cmd); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (a *App) Shutdown() {
	a.mu.Lock()
	cmd := a.currentCmd
	cancel := a.currentTask
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = killCommandTree(cmd)
}

func (a *App) beginCancelableTask() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.currentTask = cancel
	a.mu.Unlock()
	return ctx, cancel
}

func (a *App) endCancelableTask(cancel context.CancelFunc) {
	a.mu.Lock()
	if a.currentTask != nil {
		a.currentTask = nil
	}
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) handleAnalyzeMulticam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req multicamAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.MasterAudioPath, "masterAudioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.CameraPaths) == 0 {
		a.writeError(w, http.StatusBadRequest, "cameraPaths must contain at least one path")
		return
	}

	results := make([]multicamCameraResult, 0, len(req.CameraPaths))
	for _, path := range req.CameraPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if err := validateExistingFile(path, "cameraPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		metrics, err := a.analyzeSync(path, req.MasterAudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
		if err != nil {
			a.writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: %v", path, err))
			return
		}
		results = append(results, multicamCameraResult{
			Path:           path,
			DelaySeconds:   round(metrics.DelaySeconds, 3),
			DelayMs:        int(math.Round(metrics.DelaySeconds * 1000)),
			Confidence:     round(metrics.Confidence, 3),
			Duration:       round(metrics.VideoDuration, 2),
			Recommendation: describeDelay(metrics.DelaySeconds),
		})
	}

	a.writeJSON(w, http.StatusOK, multicamAnalyzeResponse{
		MasterAudioPath: req.MasterAudioPath,
		Cameras:         results,
	})
}

func (a *App) handleExportMulticamPlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req multicamExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.MasterAudioPath, "masterAudioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.CameraPaths) == 0 {
		a.writeError(w, http.StatusBadRequest, "cameraPaths must contain at least one path")
		return
	}

	preset := strings.TrimSpace(req.Preset)
	if preset == "" {
		preset = "medium"
	}
	crf := req.CRF
	if crf <= 0 {
		crf = 18
	}
	outputDir := strings.TrimSpace(req.OutputDir)

	planBackend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath, false)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	plans := make([]multicamExportPlan, 0, len(req.CameraPaths))
	for _, path := range req.CameraPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if err := validateExistingFile(path, "cameraPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		metrics, ok := measuredMetricsForPath(path, req.MeasuredCameras)
		if !ok {
			var err error
			metrics, err = a.analyzeSync(path, req.MasterAudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
			if err != nil {
				a.writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: %v", path, err))
				return
			}
		}

		plan := buildCameraAlignPlan(path, metrics.DelaySeconds, outputDir, preset, crf, metrics.Confidence, planBackend)
		plans = append(plans, plan)
	}

	a.writeJSON(w, http.StatusOK, multicamExportResponse{
		MasterAudioPath: req.MasterAudioPath,
		OutputDir:       outputDir,
		Plans:           plans,
		Note:            "Эти команды готовят выровненные video-only mezzanine файлы по таймлайну мастер-аудио. Для финального монтажа затем подключай единый master audio отдельно.",
	})
}

func (a *App) handleRenderMulticam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req multicamRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.MasterAudioPath, "masterAudioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.CameraPaths) == 0 {
		a.writeError(w, http.StatusBadRequest, "cameraPaths must contain at least one path")
		return
	}

	outputPath, cmdString, elapsed, shots, totalSeconds, err := a.renderMulticam(req, nil)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	summaries := make([]multicamShotSummary, 0, len(shots))
	for _, shot := range shots {
		summaries = append(summaries, multicamShotSummary{
			CameraIndex: shot.CameraIndex + 1,
			Start:       round(shot.Start, 3),
			End:         round(shot.End, 3),
		})
	}

	a.writeJSON(w, http.StatusOK, multicamRenderResponse{
		OutputPath:   outputPath,
		Duration:     elapsed.String(),
		Command:      cmdString,
		Shots:        summaries,
		TotalSeconds: round(totalSeconds, 3),
	})
}

func (a *App) handleRenderMulticamStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req multicamRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.MasterAudioPath, "masterAudioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.CameraPaths) == 0 {
		a.writeError(w, http.StatusBadRequest, "cameraPaths must contain at least one path")
		return
	}

	streamJSON(w, func(send func(progressEvent)) {
		outputPath, cmdString, elapsed, shots, totalSeconds, err := a.renderMulticam(req, send)
		if err != nil {
			send(progressEvent{Error: err.Error()})
			return
		}
		summaries := make([]multicamShotSummary, 0, len(shots))
		for _, shot := range shots {
			summaries = append(summaries, multicamShotSummary{
				CameraIndex: shot.CameraIndex + 1,
				Start:       round(shot.Start, 3),
				End:         round(shot.End, 3),
			})
		}
		send(progressEvent{
			Done:       true,
			OutputPath: outputPath,
			Duration:   elapsed.String(),
			Command:    cmdString,
			Shots:      summaries,
			TotalTime:  round(totalSeconds, 3),
			Message:    "render complete",
		})
	})
}

func (a *App) handlePlanShorts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req shortsPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req = normalizeShortsPlanRequest(req)
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	response, err := a.buildShortsPlanResponse(req, nil)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, response)
}

func (a *App) handlePlanShortsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req shortsPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req = normalizeShortsPlanRequest(req)
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	streamJSON(w, func(send func(progressEvent)) {
		response, err := a.buildShortsPlanResponse(req, send)
		if err != nil {
			send(progressEvent{Error: err.Error()})
			return
		}
		send(progressEvent{
			Done:       true,
			Percent:    100,
			Message:    "shorts plan complete",
			ShortsPlan: &response,
		})
	})
}

func (a *App) handleRenderShorts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req shortsRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	response, err := a.renderShorts(req, nil)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, response)
}

func (a *App) handleRenderShortsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req shortsRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	streamJSON(w, func(send func(progressEvent)) {
		response, err := a.renderShorts(req, send)
		if err != nil {
			send(progressEvent{Error: err.Error()})
			return
		}
		send(progressEvent{
			Done:     true,
			Message:  "shorts render complete",
			PlanPath: response.PlanPath,
			Files:    response.Files,
			Failed:   response.Failed,
			Rendered: response.RenderedCount,
			Duration: response.Duration,
		})
	})
}

func (a *App) handleRenderFullCaptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req fullCaptionsRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	response, err := a.renderFullCaptions(req, nil)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, response)
}

func (a *App) handleRenderFullCaptionsStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.ensureTools(); err != nil {
		a.writeError(w, http.StatusFailedDependency, err.Error())
		return
	}

	var req fullCaptionsRenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateExistingFile(req.VideoPath, "videoPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.AudioPath) != "" {
		if err := validateExistingFile(req.AudioPath, "audioPath"); err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	streamJSON(w, func(send func(progressEvent)) {
		response, err := a.renderFullCaptions(req, send)
		if err != nil {
			send(progressEvent{Error: err.Error()})
			return
		}
		send(progressEvent{
			Done:             true,
			Percent:          100,
			Message:          "full captions complete",
			OutputPath:       response.OutputPath,
			Duration:         response.Duration,
			TranscriptSource: response.TranscriptSource,
			SRTPath:          response.SRTPath,
			TextPath:         response.TextPath,
			ASSPath:          response.ASSPath,
		})
	})
}

func (a *App) analyzeSync(videoPath, audioPath string, analyzeSeconds, maxLagSeconds float64) (syncMetrics, error) {
	stagedVideoPath, cleanupVideo, err := stageInputPathForWindows(videoPath)
	if err != nil {
		return syncMetrics{}, err
	}
	defer cleanupVideo()

	stagedAudioPath, cleanupAudio, err := stageInputPathForWindows(audioPath)
	if err != nil {
		return syncMetrics{}, err
	}
	defer cleanupAudio()

	if analyzeSeconds <= 0 {
		analyzeSeconds = defaultAnalyzeSeconds
	}
	if maxLagSeconds <= 0 {
		maxLagSeconds = defaultMaxLagSeconds
	}

	videoDuration, err := a.probeDuration(stagedVideoPath)
	if err != nil {
		return syncMetrics{}, fmt.Errorf("ffprobe video: %w", err)
	}
	audioDuration, err := a.probeDuration(stagedAudioPath)
	if err != nil {
		return syncMetrics{}, fmt.Errorf("ffprobe audio: %w", err)
	}

	windowDuration := minFloat(minFloat(videoDuration, audioDuration), analyzeSeconds)
	if windowDuration <= 1 {
		return syncMetrics{}, errors.New("files are too short for analysis")
	}

	videoEnv, err := a.extractLegacyEnvelope(stagedVideoPath, windowDuration)
	if err != nil {
		return syncMetrics{}, fmt.Errorf("extract video audio: %w", err)
	}
	audioEnv, err := a.extractLegacyEnvelope(stagedAudioPath, windowDuration)
	if err != nil {
		return syncMetrics{}, fmt.Errorf("extract master audio: %w", err)
	}
	if len(videoEnv) == 0 || len(audioEnv) == 0 {
		return syncMetrics{}, errors.New("not enough audio signal to analyze")
	}
	legacyDelay := findLegacyDelay(audioEnv, videoEnv)
	delaySec := legacyDelay
	confidence := 0.75

	videoPCM, err := a.extractPCM(stagedVideoPath, windowDuration)
	if err == nil {
		audioPCM, err := a.extractPCM(stagedAudioPath, windowDuration)
		if err == nil {
			coarseVideo := buildEnvelope(videoPCM, coarseWindowSamples)
			coarseAudio := buildEnvelope(audioPCM, coarseWindowSamples)
			fineVideo := buildEnvelope(videoPCM, fineWindowSamples)
			fineAudio := buildEnvelope(audioPCM, fineWindowSamples)

			if len(coarseVideo) > 0 && len(coarseAudio) > 0 && len(fineVideo) > 0 && len(fineAudio) > 0 {
				normalizeInPlace(coarseVideo)
				normalizeInPlace(coarseAudio)
				normalizeInPlace(fineVideo)
				normalizeInPlace(fineAudio)

				coarseStepSec := float64(coarseWindowSamples) / float64(defaultSampleRate)
				fineStepSec := float64(fineWindowSamples) / float64(defaultSampleRate)

				legacyCoarseCenter := int(math.Round(legacyDelay / coarseStepSec))
				hybridCoarseRadius := int(math.Round(0.6 / coarseStepSec))
				coarseLag, coarseScore := bestLagAround(coarseAudio, coarseVideo, legacyCoarseCenter, hybridCoarseRadius)

				legacyFineCenter := int(math.Round(legacyDelay / fineStepSec))
				coarseAsFineCenter := int(math.Round((float64(coarseLag) * coarseStepSec) / fineStepSec))
				fineCenter := legacyFineCenter
				if math.Abs(float64(coarseAsFineCenter-legacyFineCenter))*fineStepSec <= 0.25 {
					fineCenter = coarseAsFineCenter
				}
				hybridFineRadius := int(math.Round(0.18 / fineStepSec))
				fineLag, fineScore := bestLagAround(fineAudio, fineVideo, fineCenter, hybridFineRadius)
				modernDelay := float64(fineLag) * fineStepSec

				delta := math.Abs(modernDelay - legacyDelay)
				if delta <= 0.12 {
					delaySec = (legacyDelay * 0.7) + (modernDelay * 0.3)
					confidence = 0.92
				} else if delta <= 0.25 && math.Abs(fineScore) > math.Abs(coarseScore)*0.85 {
					delaySec = (legacyDelay * 0.85) + (modernDelay * 0.15)
					confidence = 0.84
				} else {
					delaySec = legacyDelay
					confidence = 0.68
				}
			}
		}
	}

	return syncMetrics{
		DelaySeconds:  round(delaySec, 3),
		Confidence:    confidence,
		VideoDuration: videoDuration,
		AudioDuration: audioDuration,
	}, nil
}

func (a *App) extractLegacyEnvelope(path string, duration float64) ([]float64, error) {
	args := []string{
		"-v", "error",
		"-t", fmt.Sprintf("%.3f", duration),
		"-i", path,
		"-vn",
		"-sn",
		"-dn",
		"-ac", "1",
		"-ar", "8000",
		"-f", "s16le",
		"pipe:1",
	}
	cmd := newCommand(a.ffmpegPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}

	data := stdout.Bytes()
	if len(data) < 2 {
		return nil, errors.New("ffmpeg returned empty audio stream")
	}

	envelope := make([]float64, 0, len(data)/160+1)
	var sum float64
	var count int
	for i := 0; i < len(data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(data[i : i+2]))
		sum += math.Abs(float64(sample))
		count++
		if count == 80 {
			envelope = append(envelope, sum/80.0)
			sum = 0
			count = 0
		}
	}
	if len(envelope) == 0 {
		return nil, errors.New("empty envelope")
	}
	var total float64
	for _, value := range envelope {
		total += value
	}
	mean := total / float64(len(envelope))
	for i := range envelope {
		envelope[i] -= mean
	}
	return envelope, nil
}

func findLegacyDelay(envA, envV []float64) float64 {
	if len(envA) == 0 || len(envV) == 0 {
		return 0
	}

	step := 10
	lenALow := len(envA) / step
	envALow := make([]float64, lenALow)
	for i := 0; i < lenALow; i++ {
		var sum float64
		for j := 0; j < step; j++ {
			sum += envA[i*step+j]
		}
		envALow[i] = sum / float64(step)
	}

	lenVLow := len(envV) / step
	envVLow := make([]float64, lenVLow)
	for i := 0; i < lenVLow; i++ {
		var sum float64
		for j := 0; j < step; j++ {
			sum += envV[i*step+j]
		}
		envVLow[i] = sum / float64(step)
	}

	maxCorrLow := -1e10
	bestDelayLow := 0
	startKLow := -(lenVLow - 1)
	endKLow := lenALow - 1
	for k := startKLow; k <= endKLow; k++ {
		startI := 0
		if k > 0 {
			startI = k
		}
		endI := lenALow
		if lenVLow+k < lenALow {
			endI = lenVLow + k
		}
		var sum float64
		for i := startI; i < endI; i++ {
			sum += envALow[i] * envVLow[i-k]
		}
		if sum > maxCorrLow {
			maxCorrLow = sum
			bestDelayLow = k
		}
	}

	approxDelay := bestDelayLow * step
	window := 200
	maxCorr := -1e10
	bestDelay := approxDelay
	startK := approxDelay - window
	if startK < -(len(envV) - 1) {
		startK = -(len(envV) - 1)
	}
	endK := approxDelay + window
	if endK > len(envA)-1 {
		endK = len(envA) - 1
	}
	for k := startK; k <= endK; k++ {
		startI := 0
		if k > 0 {
			startI = k
		}
		endI := len(envA)
		if len(envV)+k < len(envA) {
			endI = len(envV) + k
		}
		var sum float64
		for i := startI; i < endI; i++ {
			sum += envA[i] * envV[i-k]
		}
		if sum > maxCorr {
			maxCorr = sum
			bestDelay = k
		}
	}

	return float64(bestDelay) / 100.0
}

func (a *App) extractPCM(path string, duration float64) ([]int16, error) {
	args := []string{
		"-v", "error",
		"-t", fmt.Sprintf("%.3f", duration),
		"-i", path,
		"-vn",
		"-sn",
		"-dn",
		"-ac", "1",
		"-ar", strconv.Itoa(defaultSampleRate),
		"-f", "s16le",
		"pipe:1",
	}
	cmd := newCommand(a.ffmpegPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}

	data := stdout.Bytes()
	if len(data) < 2 {
		return nil, errors.New("ffmpeg returned empty audio stream")
	}
	samples := make([]int16, len(data)/2)
	for i := 0; i < len(samples); i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
	}
	return samples, nil
}

func buildEnvelope(samples []int16, window int) []float64 {
	if window <= 0 || len(samples) < window {
		return nil
	}
	count := len(samples) / window
	env := make([]float64, 0, count)
	for i := 0; i < count; i++ {
		start := i * window
		end := start + window
		var sum float64
		for _, sample := range samples[start:end] {
			sum += math.Abs(float64(sample))
		}
		env = append(env, sum/float64(window))
	}
	return env
}

func normalizeInPlace(values []float64) {
	if len(values) == 0 {
		return
	}
	var mean float64
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))

	var energy float64
	for i := range values {
		values[i] -= mean
		energy += values[i] * values[i]
	}
	if energy == 0 {
		return
	}
	scale := math.Sqrt(energy / float64(len(values)))
	if scale == 0 {
		return
	}
	for i := range values {
		values[i] /= scale
	}
}

func bestLag(reference, candidate []float64, maxLag int) (int, float64) {
	return bestLagAround(reference, candidate, 0, maxLag)
}

func bestLagAround(reference, candidate []float64, center, radius int) (int, float64) {
	bestLag := center
	bestScore := -math.MaxFloat64
	for lag := center - radius; lag <= center+radius; lag++ {
		score := scoreLag(reference, candidate, lag)
		if score > bestScore {
			bestScore = score
			bestLag = lag
		}
	}
	return bestLag, bestScore
}

func scoreLag(reference, candidate []float64, lag int) float64 {
	start := 0
	if lag > 0 {
		start = lag
	}
	end := len(reference)
	if len(candidate)+lag < end {
		end = len(candidate) + lag
	}
	if lag < 0 {
		start = 0
		end = minInt(len(reference), len(candidate)+lag)
	}
	if end-start <= 8 {
		return -math.MaxFloat64
	}

	var sum float64
	var normRef float64
	var normCand float64
	for i := start; i < end; i++ {
		j := i - lag
		ref := reference[i]
		cand := candidate[j]
		sum += ref * cand
		normRef += ref * ref
		normCand += cand * cand
	}
	if normRef == 0 || normCand == 0 {
		return -math.MaxFloat64
	}
	return sum / math.Sqrt(normRef*normCand)
}

func (a *App) renderMulticam(req multicamRenderRequest, send func(progressEvent)) (string, string, time.Duration, []shotSegment, float64, error) {
	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath, true)
	if err != nil {
		return "", "", 0, nil, 0, err
	}
	if backend.Cleanup != nil {
		defer backend.Cleanup()
	}

	preset := strings.TrimSpace(req.Preset)
	if preset == "" {
		preset = "medium"
	}
	crf := req.CRF
	if crf <= 0 {
		crf = 18
	}
	shotWindow := req.ShotWindowSeconds
	if shotWindow <= 0 {
		shotWindow = 1.0
	}
	minShot := req.MinShotSeconds
	if minShot <= 0 {
		minShot = 2.5
	}
	outputPath := resolveMulticamOutputPath(req.MasterAudioPath, req.OutputPath)
	stagingRoot := ensureOutputStagingRoot(filepath.Dir(outputPath))
	stagedMasterAudioPath, cleanupMasterAudio, err := stageInputPathForWindowsInDir(req.MasterAudioPath, stagingRoot)
	if err != nil {
		return "", "", 0, nil, 0, err
	}
	defer cleanupMasterAudio()

	stagedOutputPath, finalizeOutput, cleanupOutput, err := stageOutputPathForWindows(outputPath, stagingRoot)
	if err != nil {
		return "", "", 0, nil, 0, err
	}
	defer cleanupOutput()
	primaryIndex := req.PrimaryCamera - 1
	if primaryIndex < 0 || primaryIndex >= len(req.CameraPaths) {
		primaryIndex = 0
	}

	analyses := make([]multicamAnalysis, 0, len(req.CameraPaths))
	for _, path := range req.CameraPaths {
		originalPath := strings.TrimSpace(path)
		if originalPath == "" {
			continue
		}
		if err := validateExistingFile(originalPath, "cameraPath"); err != nil {
			return "", "", 0, nil, 0, err
		}
		metrics, ok := measuredMetricsForPath(originalPath, req.MeasuredCameras)
		if !ok {
			var err error
			metrics, err = a.analyzeSync(originalPath, req.MasterAudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
			if err != nil {
				return "", "", 0, nil, 0, fmt.Errorf("%s: %w", originalPath, err)
			}
		}
		renderPath, cleanupRenderPath, err := stageInputPathForWindowsInDir(originalPath, stagingRoot)
		if err != nil {
			return "", "", 0, nil, 0, err
		}
		envelope, err := a.extractEnvelope(renderPath, metrics.VideoDuration)
		if err != nil {
			cleanupRenderPath()
			return "", "", 0, nil, 0, fmt.Errorf("%s: %w", originalPath, err)
		}
		meta, err := a.probeVideoStream(renderPath)
		if err != nil {
			meta = videoStreamMeta{Width: 1920, Height: 1080, FPS: 25, Duration: metrics.VideoDuration}
		}
		if meta.Duration <= 0 {
			meta.Duration = metrics.VideoDuration
		}
		analyses = append(analyses, multicamAnalysis{
			Path:     renderPath,
			Metrics:  metrics,
			Envelope: envelope,
			Meta:     meta,
		})
		defer cleanupRenderPath()
	}
	if len(analyses) == 0 {
		return "", "", 0, nil, 0, errors.New("no valid cameras to render")
	}

	masterDuration, err := a.probeDuration(stagedMasterAudioPath)
	if err != nil {
		return "", "", 0, nil, 0, err
	}
	totalSeconds := masterDuration
	if totalSeconds <= 0 {
		return "", "", 0, nil, 0, errors.New("master audio has invalid duration")
	}
	if req.PreviewSeconds > 0 && req.PreviewSeconds < totalSeconds {
		totalSeconds = req.PreviewSeconds
	}

	editMode := strings.TrimSpace(strings.ToLower(req.EditMode))
	shots := []shotSegment(nil)
	if editMode == "ai" || editMode == "smart-ai" {
		if strings.TrimSpace(req.AssemblyAIKey) == "" {
			return "", "", 0, nil, 0, errors.New("для умного AI multicam нужен ключ AssemblyAI")
		}
		if send != nil {
			send(progressEvent{Message: "AI: diarization и speaker-based shot plan..."})
		}
		utterances, err := a.transcribeWithAssemblyAI(context.Background(), req.MasterAudioPath, req.AssemblyAIKey, send)
		if err != nil {
			return "", "", 0, nil, 0, err
		}
		shots = buildSpeakerShotPlan(analyses, utterances, totalSeconds, primaryIndex, minShot)
	}
	if len(shots) == 0 {
		shots = buildShotPlan(analyses, totalSeconds, shotWindow, minShot, primaryIndex)
	}
	if len(shots) == 0 {
		return "", "", 0, nil, 0, errors.New("failed to build shot plan")
	}
	shots, timelineTrimStart, adjustedTotalSeconds := normalizeRenderableShots(analyses, shots, totalSeconds, primaryIndex)
	if len(shots) == 0 {
		return "", "", 0, nil, 0, errors.New("no renderable multicam segments after availability normalization")
	}
	totalSeconds = adjustedTotalSeconds

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", "", 0, nil, 0, err
	}

	referenceMeta := analyses[primaryIndex].Meta
	if referenceMeta.Width <= 0 {
		referenceMeta.Width = 1920
	}
	if referenceMeta.Height <= 0 {
		referenceMeta.Height = 1080
	}
	if referenceMeta.FPS <= 0 {
		referenceMeta.FPS = 25
	}
	fpsValue := trimFloat(referenceMeta.FPS, 3)

	filterParts := make([]string, 0, len(shots)+2)
	concatInputs := make([]string, 0, len(shots))
	for i, shot := range shots {
		camera := analyses[shot.CameraIndex]
		sourceStart := shot.Start - camera.Metrics.DelaySeconds
		sourceEnd := shot.End - camera.Metrics.DelaySeconds
		if sourceStart < 0 || sourceEnd <= sourceStart {
			continue
		}

		label := fmt.Sprintf("v%d", i)
		scalePart := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=%s", referenceMeta.Width, referenceMeta.Height, referenceMeta.Width, referenceMeta.Height, fpsValue)
		segmentFilter := fmt.Sprintf("[%d:v]trim=start=%s:end=%s,%s,setpts=PTS-STARTPTS[%s]",
			shot.CameraIndex,
			trimFloat(sourceStart, 6),
			trimFloat(sourceEnd, 6),
			scalePart,
			label,
		)
		filterParts = append(filterParts, segmentFilter)
		concatInputs = append(concatInputs, fmt.Sprintf("[%s]", label))
	}
	if len(concatInputs) == 0 {
		return "", "", 0, nil, 0, errors.New("no renderable multicam segments")
	}

	filterParts = append(filterParts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vout]", strings.Join(concatInputs, ""), len(concatInputs)))
	audioIndex := len(analyses)
	filterParts = append(filterParts, fmt.Sprintf("[%d:a]atrim=start=%s:end=%s,asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[aout]", audioIndex, trimFloat(timelineTrimStart, 6), trimFloat(timelineTrimStart+totalSeconds, 6)))

	ffmpegArgs := []string{"-y"}
	for _, camera := range analyses {
		ffmpegArgs = append(ffmpegArgs, "-i", camera.Path)
	}
	ffmpegArgs = append(ffmpegArgs, "-i", stagedMasterAudioPath)
	ffmpegArgs = append(ffmpegArgs,
		"-filter_complex", strings.Join(filterParts, ";"),
		"-map", "[vout]",
		"-map", "[aout]",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
	)
	ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
	ffmpegArgs = append(ffmpegArgs, "-progress", "pipe:1", "-nostats", stagedOutputPath)
	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)
	start := time.Now()
	if send != nil {
		send(progressEvent{Message: "ffmpeg: starting multicam render..."})
	}
	if err := a.runFFmpegCommand(backend.Executable, args, totalSeconds, send); err != nil {
		return "", "", 0, nil, 0, err
	}
	if err := finalizeOutput(); err != nil {
		return "", "", 0, nil, 0, err
	}
	cleanupDirectoryIfEmpty(filepath.Join(filepath.Dir(outputPath), "aligned"))

	return outputPath, shellJoin(append([]string{backend.Executable}, args...)), time.Since(start), shots, totalSeconds, nil
}

func (a *App) extractEnvelope(path string, duration float64) ([]float64, error) {
	if duration <= 0 {
		duration = defaultAnalyzeSeconds
	}
	samples, err := a.extractPCM(path, duration)
	if err != nil {
		return nil, err
	}
	env := buildEnvelope(samples, fineWindowSamples)
	if len(env) == 0 {
		return nil, errors.New("empty envelope")
	}
	normalizeInPlace(env)
	return env, nil
}

func buildShotPlan(cameras []multicamAnalysis, totalSeconds, shotWindow, minShot float64, primaryIndex int) []shotSegment {
	if totalSeconds <= 0 || len(cameras) == 0 {
		return nil
	}
	segments := make([]shotSegment, 0)
	switchThreshold := 0.12
	for start := 0.0; start < totalSeconds; start += shotWindow {
		end := math.Min(totalSeconds, start+shotWindow)
		bestIndex := -1
		bestScore := -math.MaxFloat64
		for index, camera := range cameras {
			if !cameraAvailableAt(camera, start, end) {
				continue
			}
			score := scoreCameraActivity(camera, start, end)
			if bestIndex == -1 || score > bestScore+0.03 || (math.Abs(score-bestScore) < 0.03 && index == primaryIndex) {
				bestIndex = index
				bestScore = score
			}
		}
		if bestIndex == -1 {
			bestIndex = selectBestTimelineCamera(cameras, start, end, primaryIndex, -1)
		}
		if len(segments) == 0 {
			segments = append(segments, shotSegment{CameraIndex: bestIndex, Start: start, End: end})
			continue
		}

		current := &segments[len(segments)-1]
		currentIndex := current.CameraIndex
		currentScore := -math.MaxFloat64
		if currentIndex >= 0 && currentIndex < len(cameras) && cameraAvailableAt(cameras[currentIndex], start, end) {
			currentScore = scoreCameraActivity(cameras[currentIndex], start, end)
		}

		chosenIndex := currentIndex
		currentDuration := current.End - current.Start
		if currentScore <= -math.MaxFloat64/2 {
			chosenIndex = bestIndex
		} else if bestIndex != currentIndex && currentDuration >= minShot && bestScore > currentScore+switchThreshold {
			chosenIndex = bestIndex
		}

		if chosenIndex == current.CameraIndex {
			current.End = end
		} else {
			segments = append(segments, shotSegment{CameraIndex: chosenIndex, Start: start, End: end})
		}
	}

	return smoothShotPlan(segments, minShot, primaryIndex)
}

func cameraAvailableAt(camera multicamAnalysis, start, end float64) bool {
	return cameraCoverage(camera, start, end) >= 0.98
}

func cameraCoverage(camera multicamAnalysis, start, end float64) float64 {
	if end <= start {
		return 0
	}
	alignedStart := camera.Metrics.DelaySeconds
	alignedEnd := camera.Metrics.DelaySeconds + camera.Meta.Duration
	overlapStart := math.Max(start, alignedStart)
	overlapEnd := math.Min(end, alignedEnd)
	if overlapEnd <= overlapStart {
		return 0
	}
	return (overlapEnd - overlapStart) / (end - start)
}

func scoreCameraActivity(camera multicamAnalysis, start, end float64) float64 {
	stepSeconds := float64(fineWindowSamples) / float64(defaultSampleRate)
	sourceStart := start - camera.Metrics.DelaySeconds
	sourceEnd := end - camera.Metrics.DelaySeconds
	if sourceEnd <= 0 || sourceStart >= camera.Meta.Duration {
		return -math.MaxFloat64
	}
	if sourceStart < 0 {
		sourceStart = 0
	}
	if sourceEnd > camera.Meta.Duration {
		sourceEnd = camera.Meta.Duration
	}
	startIndex := int(math.Floor(sourceStart / stepSeconds))
	endIndex := int(math.Ceil(sourceEnd / stepSeconds))
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > len(camera.Envelope) {
		endIndex = len(camera.Envelope)
	}
	if endIndex-startIndex <= 0 {
		return -math.MaxFloat64
	}

	var sum float64
	for _, value := range camera.Envelope[startIndex:endIndex] {
		sum += math.Abs(value)
	}
	return sum / float64(endIndex-startIndex)
}

func smoothShotPlan(segments []shotSegment, minShot float64, primaryIndex int) []shotSegment {
	if len(segments) == 0 {
		return nil
	}
	for i := 1; i < len(segments)-1; i++ {
		if segments[i].End-segments[i].Start >= minShot {
			continue
		}
		if segments[i-1].CameraIndex == segments[i+1].CameraIndex {
			segments[i-1].End = segments[i+1].End
			segments = append(segments[:i], segments[i+1:]...)
			i--
			continue
		}
		if segments[i-1].End-segments[i-1].Start >= segments[i+1].End-segments[i+1].Start {
			segments[i-1].End = segments[i].End
			segments = append(segments[:i], segments[i+1:]...)
		} else {
			segments[i+1].Start = segments[i].Start
			segments = append(segments[:i], segments[i+1:]...)
		}
		i--
	}
	if len(segments) > 0 && segments[0].End-segments[0].Start < minShot {
		segments[0].CameraIndex = primaryIndex
	}
	merged := make([]shotSegment, 0, len(segments))
	for _, segment := range segments {
		if len(merged) > 0 && merged[len(merged)-1].CameraIndex == segment.CameraIndex {
			merged[len(merged)-1].End = segment.End
			continue
		}
		merged = append(merged, segment)
	}
	return merged
}

func normalizeRenderableShot(camera multicamAnalysis, shot shotSegment) (shotSegment, bool) {
	start := math.Max(shot.Start, camera.Metrics.DelaySeconds)
	end := math.Min(shot.End, camera.Metrics.DelaySeconds+camera.Meta.Duration)
	if end <= start {
		return shotSegment{}, false
	}
	shot.Start = start
	shot.End = end
	return shot, true
}

func normalizeRenderableShots(cameras []multicamAnalysis, shots []shotSegment, totalSeconds float64, primaryIndex int) ([]shotSegment, float64, float64) {
	if len(shots) == 0 {
		return nil, 0, totalSeconds
	}

	normalized := make([]shotSegment, 0, len(shots))
	for _, shot := range shots {
		if shot.CameraIndex >= 0 && shot.CameraIndex < len(cameras) {
			if adjusted, ok := normalizeRenderableShot(cameras[shot.CameraIndex], shot); ok {
				normalized = append(normalized, adjusted)
				continue
			}
		}

		alternate := selectBestTimelineCamera(cameras, shot.Start, shot.End, primaryIndex, shot.CameraIndex)
		if alternate >= 0 && alternate < len(cameras) {
			if adjusted, ok := normalizeRenderableShot(cameras[alternate], shotSegment{CameraIndex: alternate, Start: shot.Start, End: shot.End}); ok {
				normalized = append(normalized, adjusted)
			}
		}
	}
	if len(normalized) == 0 {
		return nil, 0, totalSeconds
	}

	filled := make([]shotSegment, 0, len(normalized)*2)
	filled = append(filled, normalized[0])
	for _, segment := range normalized[1:] {
		last := &filled[len(filled)-1]
		if segment.Start > last.End {
			gapStart := last.End
			gapEnd := segment.Start

			if extended, ok := normalizeRenderableShot(cameras[last.CameraIndex], shotSegment{CameraIndex: last.CameraIndex, Start: gapStart, End: gapEnd}); ok && math.Abs(extended.Start-gapStart) < 0.001 {
				last.End = extended.End
			} else if extended, ok := normalizeRenderableShot(cameras[segment.CameraIndex], shotSegment{CameraIndex: segment.CameraIndex, Start: gapStart, End: gapEnd}); ok && math.Abs(extended.End-gapEnd) < 0.001 {
				segment.Start = extended.Start
			} else {
				gapCamera := selectBestTimelineCamera(cameras, gapStart, gapEnd, last.CameraIndex, -1)
				if gapCamera >= 0 && gapCamera < len(cameras) {
					if gap, ok := normalizeRenderableShot(cameras[gapCamera], shotSegment{CameraIndex: gapCamera, Start: gapStart, End: gapEnd}); ok {
						if len(filled) > 0 && filled[len(filled)-1].CameraIndex == gap.CameraIndex && gap.Start <= filled[len(filled)-1].End+0.001 {
							if gap.End > filled[len(filled)-1].End {
								filled[len(filled)-1].End = gap.End
							}
						} else {
							filled = append(filled, gap)
						}
					}
				}
			}
		}

		if len(filled) > 0 && filled[len(filled)-1].CameraIndex == segment.CameraIndex && segment.Start <= filled[len(filled)-1].End+0.001 {
			if segment.End > filled[len(filled)-1].End {
				filled[len(filled)-1].End = segment.End
			}
			continue
		}
		filled = append(filled, segment)
	}

	trimStart := math.Max(0, filled[0].Start)
	if trimStart > 0 {
		for i := range filled {
			filled[i].Start -= trimStart
			filled[i].End -= trimStart
		}
		totalSeconds -= trimStart
	}
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	return filled, trimStart, totalSeconds
}

func buildSpeakerShotPlan(cameras []multicamAnalysis, utterances []AssemblyUtterance, totalSeconds float64, primaryIndex int, minShot float64) []shotSegment {
	cameraCount := len(cameras)
	if totalSeconds <= 0 || cameraCount <= 0 {
		return nil
	}
	type speakerStats struct {
		duration float64
	}
	stats := map[string]speakerStats{}
	for _, utterance := range utterances {
		speaker := strings.TrimSpace(utterance.Speaker)
		if speaker == "" {
			continue
		}
		duration := math.Max(0, float64(utterance.End-utterance.Start)/1000.0)
		entry := stats[speaker]
		entry.duration += duration
		stats[speaker] = entry
	}
	primarySpeaker := ""
	primaryDuration := -1.0
	for speaker, entry := range stats {
		if entry.duration > primaryDuration {
			primarySpeaker = speaker
			primaryDuration = entry.duration
		}
	}
	utteranceThreshold := math.Max(minShot, 2.0)
	speakerMap := map[string]int{}
	nextFallbackCamera := 0
	mapSpeaker := func(speaker string, start, end float64) int {
		speaker = strings.TrimSpace(speaker)
		if speaker == "" || speaker == primarySpeaker {
			return selectBestTimelineCamera(cameras, start, end, primaryIndex, -1)
		}
		duration := end - start
		if duration < utteranceThreshold {
			return selectBestTimelineCamera(cameras, start, end, primaryIndex, -1)
		}
		if mapped, ok := speakerMap[speaker]; ok && mapped >= 0 && mapped < cameraCount {
			return selectBestTimelineCamera(cameras, start, end, mapped, -1)
		}

		best := selectBestTimelineCamera(cameras, start, end, -1, primaryIndex)
		if best == primaryIndex {
			for attempts := 0; attempts < cameraCount; attempts++ {
				candidate := nextFallbackCamera % cameraCount
				nextFallbackCamera++
				if candidate != primaryIndex {
					best = candidate
					break
				}
			}
		}
		speakerMap[speaker] = best
		return best
	}

	segments := make([]shotSegment, 0, len(utterances)+2)
	cursor := 0.0
	for _, utterance := range utterances {
		start := math.Max(0, float64(utterance.Start)/1000.0)
		end := math.Min(totalSeconds, float64(utterance.End)/1000.0)
		if end <= start {
			continue
		}
		if start > cursor {
			gapCamera := selectBestTimelineCamera(cameras, cursor, start, primaryIndex, -1)
			segments = append(segments, shotSegment{CameraIndex: gapCamera, Start: cursor, End: start})
		}
		cam := mapSpeaker(utterance.Speaker, start, end)
		segments = append(segments, shotSegment{CameraIndex: cam, Start: start, End: end})
		cursor = end
	}
	if cursor < totalSeconds {
		tailCamera := selectBestTimelineCamera(cameras, cursor, totalSeconds, primaryIndex, -1)
		segments = append(segments, shotSegment{CameraIndex: tailCamera, Start: cursor, End: totalSeconds})
	}

	merged := make([]shotSegment, 0, len(segments))
	for _, segment := range segments {
		if len(merged) > 0 && merged[len(merged)-1].CameraIndex == segment.CameraIndex {
			merged[len(merged)-1].End = segment.End
			continue
		}
		merged = append(merged, segment)
	}
	return diversifyPrimaryShots(cameras, smoothShotPlan(merged, minShot, primaryIndex), primaryIndex, minShot)
}

func selectBestTimelineCamera(cameras []multicamAnalysis, start, end float64, preferredIndex, avoidIndex int) int {
	bestIndex := -1
	bestScore := -math.MaxFloat64
	for idx, camera := range cameras {
		if idx == avoidIndex {
			continue
		}
		coverage := cameraCoverage(camera, start, end)
		if coverage < 0.85 {
			continue
		}
		score := scoreCameraActivity(camera, start, end)
		if bestIndex == -1 || score > bestScore+0.02 || (math.Abs(score-bestScore) < 0.02 && idx == preferredIndex) {
			bestIndex = idx
			bestScore = score
		}
	}
	if bestIndex != -1 {
		return bestIndex
	}

	bestCoverage := -1.0
	for idx, camera := range cameras {
		if idx == avoidIndex {
			continue
		}
		coverage := cameraCoverage(camera, start, end)
		if coverage <= 0 {
			continue
		}
		score := scoreCameraActivity(camera, start, end)
		if coverage > bestCoverage+0.05 || (math.Abs(coverage-bestCoverage) < 0.05 && (bestIndex == -1 || score > bestScore+0.02 || (math.Abs(score-bestScore) < 0.02 && idx == preferredIndex))) {
			bestIndex = idx
			bestCoverage = coverage
			bestScore = score
		}
	}
	if bestIndex != -1 {
		return bestIndex
	}
	if preferredIndex >= 0 && preferredIndex < len(cameras) {
		return preferredIndex
	}
	return 0
}

func diversifyPrimaryShots(cameras []multicamAnalysis, segments []shotSegment, primaryIndex int, minShot float64) []shotSegment {
	if len(cameras) < 3 || len(segments) == 0 {
		return segments
	}

	minCutaway := math.Max(minShot, 3.0)
	longSegmentThreshold := math.Max(minShot*3, 14.0)
	diversified := make([]shotSegment, 0, len(segments)+len(segments)/2)

	for i, segment := range segments {
		duration := segment.End - segment.Start
		if segment.CameraIndex != primaryIndex || duration < longSegmentThreshold {
			diversified = append(diversified, segment)
			continue
		}

		avoidIndex := -1
		if len(diversified) > 0 {
			avoidIndex = diversified[len(diversified)-1].CameraIndex
		}
		if i+1 < len(segments) && segments[i+1].CameraIndex != primaryIndex {
			avoidIndex = segments[i+1].CameraIndex
		}

		cutawayDuration := math.Min(math.Max(duration/3.0, minCutaway), duration-minCutaway)
		if cutawayDuration < minCutaway {
			diversified = append(diversified, segment)
			continue
		}

		cutawayStart := segment.Start + (duration-cutawayDuration)/2.0
		cutawayEnd := cutawayStart + cutawayDuration
		alternateIndex := selectBestTimelineCamera(cameras, cutawayStart, cutawayEnd, primaryIndex, avoidIndex)
		if alternateIndex == primaryIndex {
			diversified = append(diversified, segment)
			continue
		}

		if cutawayStart-segment.Start >= minShot {
			diversified = append(diversified, shotSegment{CameraIndex: primaryIndex, Start: segment.Start, End: cutawayStart})
		}
		diversified = append(diversified, shotSegment{CameraIndex: alternateIndex, Start: cutawayStart, End: cutawayEnd})
		if segment.End-cutawayEnd >= minShot {
			diversified = append(diversified, shotSegment{CameraIndex: primaryIndex, Start: cutawayEnd, End: segment.End})
		}
	}

	return smoothShotPlan(diversified, minShot, primaryIndex)
}

func (a *App) renderSyncedFile(req syncRenderRequest, send func(progressEvent)) (string, string, time.Duration, error) {
	crf := req.CRF
	if crf <= 0 {
		crf = 18
	}
	preset := strings.TrimSpace(req.Preset)
	if preset == "" {
		preset = "medium"
	}
	outputPath := resolveSyncOutputPath(req.VideoPath, req.OutputPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", "", 0, err
	}
	stagingRoot := ensureOutputStagingRoot(filepath.Dir(outputPath))
	stagedVideoPath, cleanupVideo, err := stageInputPathForWindowsInDir(req.VideoPath, stagingRoot)
	if err != nil {
		return "", "", 0, err
	}
	defer cleanupVideo()

	stagedAudioPath, cleanupAudio, err := stageInputPathForWindowsInDir(req.AudioPath, stagingRoot)
	if err != nil {
		return "", "", 0, err
	}
	defer cleanupAudio()

	stagedOutputPath, finalizeOutput, cleanupOutput, err := stageOutputPathForWindows(outputPath, stagingRoot)
	if err != nil {
		return "", "", 0, err
	}
	defer cleanupOutput()

	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath, true)
	if err != nil {
		return "", "", 0, err
	}
	if backend.Cleanup != nil {
		defer backend.Cleanup()
	}

	delay := req.DelaySeconds
	var filter string
	if delay >= 0 {
		filter = fmt.Sprintf("[0:v]setpts=PTS-STARTPTS[v];[1:a]atrim=start=%.6f,asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[a]", delay)
	} else {
		filter = fmt.Sprintf("[0:v]trim=start=%.6f,setpts=PTS-STARTPTS[v];[1:a]asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[a]", math.Abs(delay))
	}
	totalSeconds := 0.0
	videoDuration, videoErr := a.probeDuration(stagedVideoPath)
	audioDuration, audioErr := a.probeDuration(stagedAudioPath)
	if videoErr == nil && audioErr == nil {
		totalSeconds = math.Min(videoDuration, audioDuration)
	}
	if req.PreviewSeconds > 0 && (totalSeconds == 0 || req.PreviewSeconds < totalSeconds) {
		totalSeconds = req.PreviewSeconds
	}

	ffmpegArgs := []string{
		"-y",
		"-i", stagedVideoPath,
		"-i", stagedAudioPath,
		"-filter_complex", filter,
		"-map", "[v]",
		"-map", "[a]",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "192k",
		"-movflags", "+faststart",
		"-shortest",
	}
	ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
	if totalSeconds > 0 {
		ffmpegArgs = append(ffmpegArgs, "-t", trimFloat(totalSeconds, 3))
	}
	ffmpegArgs = append(ffmpegArgs, "-progress", "pipe:1", "-nostats", stagedOutputPath)
	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)
	start := time.Now()
	if send != nil {
		send(progressEvent{Message: "ffmpeg: starting render..."})
	}
	if err := a.runFFmpegCommand(backend.Executable, args, totalSeconds, send); err != nil {
		return "", "", 0, err
	}
	if err := finalizeOutput(); err != nil {
		return "", "", 0, err
	}
	return outputPath, shellJoin(append([]string{backend.Executable}, args...)), time.Since(start), nil
}

func (a *App) transcribeWithAssemblyAI(ctx context.Context, audioPath, apiKey string, send func(progressEvent)) ([]AssemblyUtterance, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("AssemblyAI key is required")
	}

	tempRoot := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "analysis-temp")
	_ = os.MkdirAll(tempRoot, 0755)
	stagedAudioPath, cleanupAudio, err := stageInputPathForWindowsInDir(audioPath, tempRoot)
	if err != nil {
		return nil, err
	}
	defer cleanupAudio()
	tempWav := filepath.Join(tempRoot, fmt.Sprintf("ai_master_%d.wav", time.Now().UnixNano()))
	defer os.Remove(tempWav)

	cmd := newCommand(a.ffmpegPath, "-y", "-i", stagedAudioPath, "-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", tempWav)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("prepare wav: %s", msg)
	}
	if send != nil {
		send(progressEvent{Message: "AI: upload audio to AssemblyAI..."})
	}
	uploadURL, err := uploadAssemblyAIFile(ctx, apiKey, tempWav, send)
	if err != nil {
		return nil, err
	}

	body := fmt.Sprintf(`{"audio_url":"%s","speaker_labels":true,"speech_models":["universal-3-pro","universal-2"],"language_detection":true}`, uploadURL)
	req2, err := http.NewRequestWithContext(ctx, "POST", "https://api.assemblyai.com/v2/transcript", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Authorization", apiKey)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := (&http.Client{Timeout: 10 * time.Minute}).Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	transcriptBody, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}
	var trRes AssemblyTranscriptRes
	if err := json.Unmarshal(transcriptBody, &trRes); err != nil {
		return nil, err
	}
	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 {
		message := strings.TrimSpace(trRes.Error)
		if message == "" {
			message = strings.TrimSpace(string(transcriptBody))
		}
		if message == "" {
			message = resp2.Status
		}
		return nil, fmt.Errorf("AssemblyAI transcript start failed: %s", message)
	}
	if strings.TrimSpace(trRes.ID) == "" {
		if strings.TrimSpace(trRes.Error) != "" {
			return nil, fmt.Errorf("AssemblyAI transcript start failed: %s", strings.TrimSpace(trRes.Error))
		}
		return nil, fmt.Errorf("AssemblyAI transcript start failed: %s", strings.TrimSpace(string(transcriptBody)))
	}

	started := time.Now()
	audioDuration := trRes.AudioDuration
	for {
		if err := ctx.Err(); err != nil {
			return nil, errors.New("operation cancelled")
		}
		if send != nil {
			percent, message := estimateAssemblyAIProgress("queued", audioDuration, started)
			send(progressEvent{Percent: percent, Message: message})
		}
		select {
		case <-ctx.Done():
			return nil, errors.New("operation cancelled")
		case <-time.After(3 * time.Second):
		}

		pollReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.assemblyai.com/v2/transcript/"+trRes.ID, nil)
		if err != nil {
			return nil, err
		}
		pollReq.Header.Set("Authorization", apiKey)
		pollResp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(pollReq)
		if err != nil {
			return nil, err
		}
		var pollRes AssemblyPollRes
		err = json.NewDecoder(pollResp.Body).Decode(&pollRes)
		pollResp.Body.Close()
		if err != nil {
			return nil, err
		}
		if pollRes.AudioDuration > 0 {
			audioDuration = pollRes.AudioDuration
		}
		if send != nil {
			percent, message := estimateAssemblyAIProgress(pollRes.Status, audioDuration, started)
			send(progressEvent{Percent: percent, Message: message})
		}
		switch pollRes.Status {
		case "completed":
			return pollRes.Utterances, nil
		case "error":
			if pollRes.Error == "" {
				pollRes.Error = "AssemblyAI transcription failed"
			}
			return nil, errors.New(pollRes.Error)
		}
	}
}

func uploadAssemblyAIFile(ctx context.Context, apiKey, path string, send func(progressEvent)) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", errors.New("operation cancelled")
		}
		if send != nil && attempt > 1 {
			send(progressEvent{Message: fmt.Sprintf("AI: retrying AssemblyAI upload (%d/%d)...", attempt, maxAttempts)})
		}

		file, err := os.Open(path)
		if err != nil {
			return "", err
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.assemblyai.com/v2/upload", file)
		if err != nil {
			file.Close()
			return "", err
		}
		req.Header.Set("Authorization", apiKey)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = info.Size()

		resp, err := client.Do(req)
		file.Close()
		if err != nil {
			if attempt < maxAttempts && isRetryableAssemblyAIError(err) {
				time.Sleep(time.Duration(attempt*2) * time.Second)
				continue
			}
			return "", err
		}

		uploadBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if attempt < maxAttempts && isRetryableAssemblyAIError(readErr) {
				time.Sleep(time.Duration(attempt*2) * time.Second)
				continue
			}
			return "", readErr
		}

		var upRes AssemblyUploadRes
		if err := json.Unmarshal(uploadBody, &upRes); err != nil {
			return "", err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			message := strings.TrimSpace(upRes.Error)
			if message == "" {
				message = strings.TrimSpace(string(uploadBody))
			}
			if message == "" {
				message = resp.Status
			}
			return "", fmt.Errorf("AssemblyAI upload failed: %s", message)
		}
		if strings.TrimSpace(upRes.UploadURL) == "" {
			if strings.TrimSpace(upRes.Error) != "" {
				return "", fmt.Errorf("AssemblyAI upload failed: %s", strings.TrimSpace(upRes.Error))
			}
			return "", errors.New("AssemblyAI upload failed")
		}
		return strings.TrimSpace(upRes.UploadURL), nil
	}

	return "", errors.New("AssemblyAI upload failed")
}

func isRetryableAssemblyAIError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "eof") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "connection aborted") ||
		strings.Contains(message, "unexpected eof") ||
		strings.Contains(message, "broken pipe")
}

func estimateAssemblyAIProgress(status string, audioDuration float64, started time.Time) (float64, string) {
	status = strings.TrimSpace(strings.ToLower(status))
	elapsed := time.Since(started).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	switch status {
	case "completed":
		return 100, "AssemblyAI: completed"
	case "error":
		return 0, "AssemblyAI: failed"
	case "processing":
		if audioDuration > 0 {
			ratioBase := math.Max(audioDuration*0.85, 20)
			estimate := 32 + (elapsed/ratioBase)*58
			if estimate > 94 {
				estimate = 94
			}
			return round(estimate, 1), fmt.Sprintf("AssemblyAI: processing (~%.1f%% estimate)", round(estimate, 1))
		}
		return 55, fmt.Sprintf("AssemblyAI: processing (%ds elapsed)", int(elapsed))
	default:
		if audioDuration > 0 {
			return 18, "AssemblyAI: queued (~18% estimate)"
		}
		return 12, "AssemblyAI: queued"
	}
}

func (a *App) planShorts(req shortsPlanRequest) ([]shortSegment, string, string, error) {
	if strings.TrimSpace(req.AssemblyAIKey) == "" {
		return nil, "", "", errors.New("для Shorts нужен ключ AssemblyAI")
	}
	utterances, err := a.transcribeWithAssemblyAI(context.Background(), req.MasterAudioPath, req.AssemblyAIKey, nil)
	if err != nil {
		return nil, "", "", err
	}
	count := req.ShortsCount
	if count <= 0 {
		count = 3
	}
	segments := buildHeuristicShorts(utterances, count, len(req.CameraPaths), req.PrimaryCamera-1)
	if len(segments) == 0 {
		return nil, "", "", errors.New("не удалось построить shorts plan")
	}

	provider := strings.TrimSpace(strings.ToLower(req.AIProvider))
	note := "Собран heuristic shorts plan по diarization и длине реплик."
	if provider == "gemini" || provider == "openai" {
		if strings.TrimSpace(req.AIKey) != "" {
			refined, err := a.refineShortsWithLLM(provider, req.AIKey, req.AIPrompt, utterances, count)
			if err == nil && len(refined) > 0 {
				for i := range refined {
					if i < len(segments) {
						segments[i].Title = refined[i].Title
						segments[i].Reason = refined[i].Reason
					}
				}
				note = "Shorts plan усилен LLM-подсказками поверх diarization."
			} else {
				note = "Diarization сработал, но LLM refinement не ответил; показан fallback plan."
			}
		}
	}

	for i := range segments {
		segments[i].Command = shellJoin([]string{
			a.ffmpegPath, "-y", "-ss", trimFloat(segments[i].Start, 3), "-to", trimFloat(segments[i].End, 3),
			"-i", req.MasterAudioPath, "-c:a", "aac", fmt.Sprintf("short_%02d.m4a", i+1),
		})
	}
	return segments, note, provider, nil
}

func buildHeuristicShorts(utterances []AssemblyUtterance, count, cameraCount, primaryIndex int) []shortSegment {
	type candidate struct {
		start float64
		end   float64
		text  string
		score float64
		cam   int
	}
	candidates := make([]candidate, 0, len(utterances))
	speakerMap := map[string]int{}
	nextCamera := 0
	for _, utterance := range utterances {
		start := float64(utterance.Start) / 1000.0
		end := float64(utterance.End) / 1000.0
		if end-start < 8 || end-start > 70 {
			continue
		}
		cam, ok := speakerMap[utterance.Speaker]
		if !ok {
			if len(speakerMap) == 0 {
				cam = primaryIndex
			} else {
				cam = nextCamera
				if cam == primaryIndex {
					cam++
				}
				if cam >= cameraCount {
					cam = minInt(cameraCount-1, 1)
				}
				nextCamera++
			}
			speakerMap[utterance.Speaker] = cam
		}
		score := (end - start) + float64(len(strings.Fields(utterance.Text)))/4.0
		candidates = append(candidates, candidate{start: start, end: end, text: strings.TrimSpace(utterance.Text), score: score, cam: cam})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	if len(candidates) > count {
		candidates = candidates[:count]
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].start < candidates[j].start })
	segments := make([]shortSegment, 0, len(candidates))
	for i, item := range candidates {
		title := fmt.Sprintf("Short %d", i+1)
		if item.text != "" {
			words := strings.Fields(item.text)
			if len(words) > 8 {
				words = words[:8]
			}
			title = strings.Join(words, " ")
		}
		segments = append(segments, shortSegment{
			Title:      title,
			Start:      round(item.start, 3),
			End:        round(item.end, 3),
			Reason:     "Длинная реплика с высоким conversational weight.",
			CameraHint: item.cam + 1,
		})
	}
	return segments
}

func normalizeShortsPlanRequest(req shortsPlanRequest) shortsPlanRequest {
	if strings.TrimSpace(req.VideoPath) == "" && len(req.CameraPaths) > 0 {
		index := req.PrimaryCamera - 1
		if index < 0 || index >= len(req.CameraPaths) {
			index = 0
		}
		req.VideoPath = strings.TrimSpace(req.CameraPaths[index])
	}
	if strings.TrimSpace(req.AudioPath) == "" {
		req.AudioPath = strings.TrimSpace(req.MasterAudioPath)
	}
	return req
}

func buildShortTitleFromText(text string, index int) string {
	title := fmt.Sprintf("Clip %d", index)
	if strings.TrimSpace(text) == "" {
		return title
	}
	words := strings.Fields(text)
	if len(words) > 8 {
		words = words[:8]
	}
	title = strings.TrimSpace(strings.Join(words, " "))
	if title == "" {
		return fmt.Sprintf("Clip %d", index)
	}
	return title
}

func (a *App) buildShortPreviewCommand(videoPath string, syncDelay float64, segment shortSegment) string {
	videoStart := math.Max(0, segment.Start-syncDelay)
	videoEnd := math.Max(videoStart+0.25, segment.End-syncDelay)
	return shellJoin([]string{
		a.ffmpegPath,
		"-y",
		"-ss", trimFloat(videoStart, 3),
		"-to", trimFloat(videoEnd, 3),
		"-i", videoPath,
		"-vf", "scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920",
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "18",
		"-c:a", "aac",
		fmt.Sprintf("%s_preview_youtube-shorts.mp4", sanitizeShortLabel(segment.Title)),
	})
}

func (a *App) buildShortsPlanResponse(req shortsPlanRequest, send func(progressEvent)) (shortsPlanResponse, error) {
	req = normalizeShortsPlanRequest(req)
	ctx, cancel := a.beginCancelableTask()
	defer a.endCancelableTask(cancel)
	if strings.TrimSpace(req.AssemblyAIKey) == "" {
		return shortsPlanResponse{}, errors.New("AssemblyAI key is required for Shorts / Reels")
	}
	if send != nil {
		send(progressEvent{Message: "Shorts: validating sources..."})
	}

	transcriptSource := strings.TrimSpace(req.AudioPath)
	timelineSource := "video"
	syncDelay := 0.0
	timelineDuration := 0.0
	if transcriptSource == "" {
		transcriptSource = strings.TrimSpace(req.VideoPath)
		timelineDuration, _ = a.probeDuration(req.VideoPath)
	} else {
		if send != nil {
			send(progressEvent{Message: "Shorts: measuring sync between video and master audio..."})
		}
		metrics, err := a.analyzeSync(req.VideoPath, req.AudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
		if err != nil {
			return shortsPlanResponse{}, fmt.Errorf("failed to align video with master audio for shorts: %w", err)
		}
		syncDelay = metrics.DelaySeconds
		timelineSource = "master-audio"
		timelineDuration, _ = a.probeDuration(req.AudioPath)
	}
	if timelineDuration <= 0 {
		timelineDuration, _ = a.probeDuration(transcriptSource)
	}

	if send != nil {
		send(progressEvent{Message: "Shorts: sending audio to AssemblyAI..."})
	}
	utterances, err := a.transcribeWithAssemblyAI(ctx, transcriptSource, req.AssemblyAIKey, func(event progressEvent) {
		if send == nil {
			return
		}
		send(progressEvent{Percent: event.Percent, Message: event.Message})
	})
	if err != nil {
		return shortsPlanResponse{}, err
	}
	if send != nil {
		send(progressEvent{Message: "Shorts: assembling clip candidates..."})
	}

	count := req.ShortsCount
	if count <= 0 {
		count = 3
	}
	segments := buildHeuristicShorts(utterances, count, 1, 0)
	if len(segments) == 0 {
		return shortsPlanResponse{}, errors.New("unable to build a shorts plan from the transcript")
	}

	provider := strings.TrimSpace(strings.ToLower(req.AIProvider))
	note := "Heuristic shorts plan built from speaker timing and utterance weight."
	if provider == "gemini" || provider == "openai" {
		if strings.TrimSpace(req.AIKey) != "" {
			if send != nil {
				send(progressEvent{Message: "Shorts: refining titles with the selected AI model..."})
			}
			refined, err := a.refineShortsWithLLM(provider, req.AIKey, req.AIPrompt, utterances, count)
			if err == nil && len(refined) > 0 {
				for i := range refined {
					if i >= len(segments) {
						break
					}
					if strings.TrimSpace(refined[i].Title) != "" {
						segments[i].Title = strings.TrimSpace(refined[i].Title)
					}
					if strings.TrimSpace(refined[i].Reason) != "" {
						segments[i].Reason = strings.TrimSpace(refined[i].Reason)
					}
					if refined[i].End > refined[i].Start {
						segments[i].Start = round(refined[i].Start, 3)
						segments[i].End = round(refined[i].End, 3)
					}
				}
				note = "Shorts plan was refined with LLM suggestions on top of the transcript timing."
			} else {
				note = "Transcript plan built successfully, but LLM refinement did not return a usable result."
			}
		}
	}

	for i := range segments {
		segments[i].ID = fmt.Sprintf("clip-%02d", i+1)
		if strings.TrimSpace(segments[i].Title) == "" {
			segments[i].Title = buildShortTitleFromText("", i+1)
		}
		segments[i].Duration = round(math.Max(0, segments[i].End-segments[i].Start), 3)
		segments[i].Enabled = true
		segments[i].CameraHint = 1
		segments[i].PreviewCommand = a.buildShortPreviewCommand(req.VideoPath, syncDelay, segments[i])
		segments[i].Command = segments[i].PreviewCommand
	}
	if send != nil {
		send(progressEvent{Message: "Shorts: finalizing plan..."})
	}

	return shortsPlanResponse{
		Provider:         provider,
		Segments:         segments,
		Utterances:       utterances,
		Note:             note,
		TimelineSource:   timelineSource,
		SyncDelaySeconds: round(syncDelay, 3),
		TimelineDuration: round(timelineDuration, 3),
	}, nil
}

func (a *App) refineShortsWithLLM(provider, apiKey, prompt string, utterances []AssemblyUtterance, count int) ([]shortSegment, error) {
	type llmSegment struct {
		Title  string  `json:"title"`
		Reason string  `json:"reason"`
		Start  float64 `json:"start"`
		End    float64 `json:"end"`
	}
	transcriptLines := make([]string, 0, len(utterances))
	for _, utterance := range utterances {
		transcriptLines = append(transcriptLines, fmt.Sprintf("[%0.2f-%0.2f] %s: %s", float64(utterance.Start)/1000.0, float64(utterance.End)/1000.0, utterance.Speaker, utterance.Text))
	}
	userPrompt := strings.TrimSpace(prompt)
	if userPrompt == "" {
		userPrompt = "Find the strongest short-form highlight moments."
	}
	systemPrompt := fmt.Sprintf("Return JSON array only. Pick up to %d highlight segments for short videos. Fields: title, reason, start, end.", count)
	fullPrompt := systemPrompt + "\n\n" + userPrompt + "\n\nTranscript:\n" + strings.Join(transcriptLines, "\n")

	var body []byte
	var req *http.Request
	var err error
	switch provider {
	case "gemini":
		payload := map[string]any{
			"contents": []map[string]any{{"parts": []map[string]string{{"text": fullPrompt}}}},
		}
		body, _ = json.Marshal(payload)
		req, err = http.NewRequest("POST", "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:generateContent?key="+apiKey, bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	case "openai":
		payload := map[string]any{
			"model": "gpt-4.1-mini",
			"input": fullPrompt,
		}
		body, _ = json.Marshal(payload)
		req, err = http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	default:
		return nil, errors.New("unsupported ai provider")
	}
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	extractJSON := func(text string) string {
		start := strings.Index(text, "[")
		end := strings.LastIndex(text, "]")
		if start >= 0 && end > start {
			return text[start : end+1]
		}
		return text
	}

	text := string(raw)
	if provider == "gemini" {
		var payload struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
			return nil, errors.New("empty Gemini response")
		}
		text = payload.Candidates[0].Content.Parts[0].Text
	} else if provider == "openai" {
		var payload struct {
			Output []struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		if len(payload.Output) == 0 || len(payload.Output[0].Content) == 0 {
			return nil, errors.New("empty OpenAI response")
		}
		text = payload.Output[0].Content[0].Text
	}
	text = extractJSON(text)

	var parsed []llmSegment
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, err
	}
	segments := make([]shortSegment, 0, len(parsed))
	for _, item := range parsed {
		segments = append(segments, shortSegment{
			Title:  item.Title,
			Start:  item.Start,
			End:    item.End,
			Reason: item.Reason,
		})
	}
	return segments, nil
}

func shortRenderPresetCatalog() map[string]shortRenderPreset {
	return map[string]shortRenderPreset{
		"youtube-shorts":    {ID: "youtube-shorts", FileSuffix: "youtube-shorts", Width: 1080, Height: 1920},
		"tiktok":            {ID: "tiktok", FileSuffix: "tiktok", Width: 1080, Height: 1920},
		"instagram-reels":   {ID: "instagram-reels", FileSuffix: "instagram-reels", Width: 1080, Height: 1920},
		"square":            {ID: "square", FileSuffix: "square", Width: 1080, Height: 1080},
		"feed":              {ID: "feed", FileSuffix: "feed", Width: 1080, Height: 1350},
		"story":             {ID: "story", FileSuffix: "story", Width: 1080, Height: 1920},
		"horizontal-teaser": {ID: "horizontal-teaser", FileSuffix: "horizontal-teaser", Width: 1920, Height: 1080},
	}
}

func defaultShortRenderPresetIDs() []string {
	return []string{
		"youtube-shorts",
		"tiktok",
		"instagram-reels",
	}
}

func resolveShortRenderPresets(ids []string) []shortRenderPreset {
	catalog := shortRenderPresetCatalog()
	if len(ids) == 0 {
		return []shortRenderPreset{
			{ID: "source-original", FileSuffix: "source"},
		}
	}
	presets := make([]shortRenderPreset, 0, len(ids))
	seen := map[string]bool{}
	for _, raw := range ids {
		id := strings.TrimSpace(strings.ToLower(raw))
		preset, ok := catalog[id]
		if !ok || seen[id] {
			continue
		}
		presets = append(presets, preset)
		seen[id] = true
	}
	return presets
}

func sanitizeShortLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "clip"
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		isAlpha := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlpha {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "clip"
	}
	if len(slug) > 48 {
		slug = strings.Trim(slug[:48], "-")
	}
	if slug == "" {
		return "clip"
	}
	return slug
}

func shortOutputBaseName(segment shortSegment, order int) string {
	return fmt.Sprintf("%02d_%s", order, sanitizeShortLabel(segment.Title))
}

func buildShortVideoFilter(preset shortRenderPreset) string {
	filters := make([]string, 0, 4)
	if preset.Width > 0 && preset.Height > 0 {
		filters = append(filters,
			fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase", preset.Width, preset.Height),
			fmt.Sprintf("crop=%d:%d", preset.Width, preset.Height),
		)
	}
	filters = append(filters, "setsar=1")
	return strings.Join(filters, ",")
}

type captionChunk struct {
	Text  string
	Start float64
	End   float64
}

func shortCaptionFontSize(preset shortRenderPreset) int {
	if preset.Height > preset.Width {
		return maxInt(22, preset.Height/48)
	}
	if preset.Width == preset.Height {
		return maxInt(24, preset.Height/42)
	}
	return maxInt(24, preset.Height/36)
}

func captionLayoutProfile(preset shortRenderPreset) (wordsPerLine, wordsPerChunk, maxCharsPerLine, maxCharsPerChunk int) {
	if preset.Height > preset.Width {
		return 2, 3, 16, 22
	}
	if preset.Width == preset.Height {
		return 3, 4, 20, 28
	}
	if preset.Width >= 1600 {
		return 5, 7, 28, 46
	}
	return 4, 6, 24, 40
}

func buildTimedCaptionChunks(utterances []AssemblyUtterance, timelineStart, timelineEnd float64, wordsPerChunk, maxCharsPerChunk int) []captionChunk {
	if len(utterances) == 0 || timelineEnd <= timelineStart {
		return nil
	}
	chunks := make([]captionChunk, 0, len(utterances)*2)
	for _, utterance := range utterances {
		utteranceStart := float64(utterance.Start) / 1000.0
		utteranceEnd := float64(utterance.End) / 1000.0
		if utteranceEnd <= timelineStart || utteranceStart >= timelineEnd {
			continue
		}
		relativeStart := math.Max(0, utteranceStart-timelineStart)
		relativeEnd := math.Min(timelineEnd-timelineStart, utteranceEnd-timelineStart)
		if relativeEnd-relativeStart < 0.2 {
			continue
		}
		if len(utterance.Words) > 0 {
			wordChunks := splitTimedCaptionWordChunks(utterance.Words, timelineStart, timelineEnd, wordsPerChunk, maxCharsPerChunk)
			if len(wordChunks) > 0 {
				chunks = append(chunks, wordChunks...)
				continue
			}
		}
		parts := splitCaptionChunks(strings.TrimSpace(utterance.Text), wordsPerChunk, maxCharsPerChunk)
		chunks = append(chunks, buildEstimatedCaptionChunks(parts, relativeStart, relativeEnd)...)
	}
	return chunks
}

func buildUtteranceCaptionChunks(utterances []AssemblyUtterance, timelineStart, timelineEnd float64) []captionChunk {
	if len(utterances) == 0 || timelineEnd <= timelineStart {
		return nil
	}
	chunks := make([]captionChunk, 0, len(utterances))
	for _, utterance := range utterances {
		utteranceStart := float64(utterance.Start) / 1000.0
		utteranceEnd := float64(utterance.End) / 1000.0
		if utteranceEnd <= timelineStart || utteranceStart >= timelineEnd {
			continue
		}
		start := math.Max(0, utteranceStart-timelineStart)
		end := math.Min(timelineEnd-timelineStart, utteranceEnd-timelineStart)
		text := strings.TrimSpace(strings.ReplaceAll(utterance.Text, "\n", " "))
		if text == "" || end-start < 0.15 {
			continue
		}
		chunks = append(chunks, captionChunk{
			Text:  text,
			Start: start,
			End:   end,
		})
	}
	return chunks
}

func splitUtteranceSentences(text string) []string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if text == "" {
		return nil
	}
	var parts []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		switch r {
		case '.', '!', '?', '…':
			segment := strings.TrimSpace(current.String())
			if segment != "" {
				parts = append(parts, segment)
			}
			current.Reset()
		}
	}
	if tail := strings.TrimSpace(current.String()); tail != "" {
		parts = append(parts, tail)
	}
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.Join(strings.Fields(part), " "))
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	if len(cleaned) == 0 {
		return []string{text}
	}
	return cleaned
}

func buildSentenceCaptionChunks(utterances []AssemblyUtterance, timelineStart, timelineEnd float64) []captionChunk {
	if len(utterances) == 0 || timelineEnd <= timelineStart {
		return nil
	}
	chunks := make([]captionChunk, 0, len(utterances)*2)
	for _, utterance := range utterances {
		utteranceStart := float64(utterance.Start) / 1000.0
		utteranceEnd := float64(utterance.End) / 1000.0
		if utteranceEnd <= timelineStart || utteranceStart >= timelineEnd {
			continue
		}
		start := math.Max(0, utteranceStart-timelineStart)
		end := math.Min(timelineEnd-timelineStart, utteranceEnd-timelineStart)
		if end-start < 0.15 {
			continue
		}
		parts := splitUtteranceSentences(utterance.Text)
		if len(parts) <= 1 {
			text := strings.TrimSpace(strings.Join(strings.Fields(utterance.Text), " "))
			if text != "" {
				chunks = append(chunks, captionChunk{Text: text, Start: start, End: end})
			}
			continue
		}
		chunks = append(chunks, buildEstimatedCaptionChunks(parts, start, end)...)
	}
	return chunks
}

func splitTimedCaptionWordChunks(words []AssemblyWord, timelineStart, timelineEnd float64, wordsPerChunk, maxCharsPerChunk int) []captionChunk {
	if len(words) == 0 {
		return nil
	}
	if wordsPerChunk < 2 {
		wordsPerChunk = 2
	}
	if maxCharsPerChunk < 18 {
		maxCharsPerChunk = 18
	}
	chunks := make([]captionChunk, 0, (len(words)+wordsPerChunk-1)/wordsPerChunk)
	currentWords := make([]AssemblyWord, 0, wordsPerChunk)
	currentText := make([]string, 0, wordsPerChunk)
	flush := func() {
		if len(currentWords) == 0 {
			return
		}
		start := math.Max(0, float64(currentWords[0].Start)/1000.0-timelineStart)
		end := math.Min(timelineEnd-timelineStart, float64(currentWords[len(currentWords)-1].End)/1000.0-timelineStart)
		if end-start >= 0.12 {
			chunks = append(chunks, captionChunk{
				Text:  strings.Join(currentText, " "),
				Start: start,
				End:   end,
			})
		}
		currentWords = currentWords[:0]
		currentText = currentText[:0]
	}
	for _, word := range words {
		wordText := strings.TrimSpace(word.Text)
		if wordText == "" {
			continue
		}
		wordStart := float64(word.Start) / 1000.0
		wordEnd := float64(word.End) / 1000.0
		if wordEnd <= timelineStart || wordStart >= timelineEnd {
			continue
		}
		candidate := strings.TrimSpace(strings.Join(append(currentText, wordText), " "))
		shouldFlush := len(currentText) > 0 && (len(currentText) >= wordsPerChunk || utf8.RuneCountInString(candidate) > maxCharsPerChunk)
		if shouldFlush {
			flush()
		}
		currentWords = append(currentWords, word)
		currentText = append(currentText, wordText)
		if len(currentText) >= 2 && hasCaptionBreakPunctuation(wordText) {
			flush()
		}
	}
	flush()
	return chunks
}

func buildEstimatedCaptionChunks(parts []string, relativeStart, relativeEnd float64) []captionChunk {
	if len(parts) == 0 || relativeEnd <= relativeStart {
		return nil
	}
	chunks := make([]captionChunk, 0, len(parts))
	totalWords := 0
	wordCounts := make([]int, 0, len(parts))
	for _, part := range parts {
		count := len(strings.Fields(part))
		if count <= 0 {
			count = 1
		}
		wordCounts = append(wordCounts, count)
		totalWords += count
	}
	available := relativeEnd - relativeStart
	cursor := relativeStart
	for index, part := range parts {
		duration := available / float64(len(parts))
		if totalWords > 0 {
			duration = available * (float64(wordCounts[index]) / float64(totalWords))
		}
		if duration < 0.45 {
			duration = 0.45
		}
		start := cursor
		end := math.Min(relativeEnd, start+duration)
		if index == len(parts)-1 || end >= relativeEnd {
			end = relativeEnd
		}
		cursor = end
		if end-start < 0.15 {
			continue
		}
		chunks = append(chunks, captionChunk{
			Text:  part,
			Start: start,
			End:   end,
		})
	}
	return chunks
}

func drawtextFontArg(fontName string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	fontCandidates := map[string][]struct {
		fsPath     string
		ffmpegPath string
	}{
		"arial": {
			{fsPath: `C:\Windows\Fonts\arial.ttf`, ffmpegPath: `C\:/Windows/Fonts/arial.ttf`},
			{fsPath: `C:\Windows\Fonts\ARIAL.TTF`, ffmpegPath: `C\:/Windows/Fonts/ARIAL.TTF`},
		},
		"verdana": {
			{fsPath: `C:\Windows\Fonts\verdana.ttf`, ffmpegPath: `C\:/Windows/Fonts/verdana.ttf`},
			{fsPath: `C:\Windows\Fonts\VERDANA.TTF`, ffmpegPath: `C\:/Windows/Fonts/VERDANA.TTF`},
		},
		"tahoma": {
			{fsPath: `C:\Windows\Fonts\tahoma.ttf`, ffmpegPath: `C\:/Windows/Fonts/tahoma.ttf`},
			{fsPath: `C:\Windows\Fonts\TAHOMA.TTF`, ffmpegPath: `C\:/Windows/Fonts/TAHOMA.TTF`},
		},
		"trebuchet-ms": {
			{fsPath: `C:\Windows\Fonts\trebuc.ttf`, ffmpegPath: `C\:/Windows/Fonts/trebuc.ttf`},
			{fsPath: `C:\Windows\Fonts\TREBUC.TTF`, ffmpegPath: `C\:/Windows/Fonts/TREBUC.TTF`},
		},
		"georgia": {
			{fsPath: `C:\Windows\Fonts\georgia.ttf`, ffmpegPath: `C\:/Windows/Fonts/georgia.ttf`},
			{fsPath: `C:\Windows\Fonts\GEORGIA.TTF`, ffmpegPath: `C\:/Windows/Fonts/GEORGIA.TTF`},
		},
		"montserrat": {
			{fsPath: `C:\Windows\Fonts\Montserrat-Regular.ttf`, ffmpegPath: `C\:/Windows/Fonts/Montserrat-Regular.ttf`},
			{fsPath: `C:\Windows\Fonts\MONTSERRAT-REGULAR.TTF`, ffmpegPath: `C\:/Windows/Fonts/MONTSERRAT-REGULAR.TTF`},
			{fsPath: `C:\Windows\Fonts\Montserrat.ttf`, ffmpegPath: `C\:/Windows/Fonts/Montserrat.ttf`},
		},
		"segoe-ui": {
			{fsPath: `C:\Windows\Fonts\segoeui.ttf`, ffmpegPath: `C\:/Windows/Fonts/segoeui.ttf`},
			{fsPath: `C:\Windows\Fonts\SEGOEUI.TTF`, ffmpegPath: `C\:/Windows/Fonts/SEGOEUI.TTF`},
		},
	}
	order := []string{strings.TrimSpace(strings.ToLower(fontName)), "segoe-ui", "arial", "verdana", "tahoma", "trebuchet-ms", "georgia", "montserrat"}
	seen := map[string]bool{}
	for _, key := range order {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		for _, item := range fontCandidates[key] {
			if _, err := os.Stat(item.fsPath); err == nil {
				return fmt.Sprintf("fontfile='%s':", item.ffmpegPath)
			}
		}
	}
	return ""
}

func wrapCaptionText(text string, wordsPerLine, maxCharsPerLine int) string {
	words := strings.Fields(strings.ReplaceAll(text, "\n", " "))
	if len(words) == 0 {
		return ""
	}
	if wordsPerLine < 2 {
		wordsPerLine = 2
	}
	if maxCharsPerLine < 12 {
		maxCharsPerLine = 12
	}
	lines := make([]string, 0, 2)
	current := make([]string, 0, wordsPerLine)
	for _, word := range words {
		candidate := strings.TrimSpace(strings.Join(append(current, word), " "))
		if len(current) > 0 && (len(current) >= wordsPerLine || utf8.RuneCountInString(candidate) > maxCharsPerLine) {
			lines = append(lines, strings.Join(current, " "))
			current = []string{word}
			if len(lines) == 2 {
				break
			}
			continue
		}
		current = append(current, word)
	}
	if len(lines) < 2 && len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	consumedWords := 0
	for _, line := range lines {
		consumedWords += len(strings.Fields(line))
	}
	if consumedWords < len(words) && len(lines) > 0 {
		lines[len(lines)-1] = strings.TrimSpace(lines[len(lines)-1] + " " + strings.Join(words[consumedWords:], " "))
	}
	return strings.Join(lines, "\n")
}

func forceBalancedTwoLineCaption(text string, maxCharsPerLine int) string {
	words := strings.Fields(strings.ReplaceAll(text, "\n", " "))
	if len(words) < 2 {
		return strings.TrimSpace(text)
	}
	bestSplit := 1
	bestScore := math.MaxFloat64
	for i := 1; i < len(words); i++ {
		left := strings.Join(words[:i], " ")
		right := strings.Join(words[i:], " ")
		leftLen := utf8.RuneCountInString(left)
		rightLen := utf8.RuneCountInString(right)
		penalty := math.Abs(float64(leftLen - rightLen))
		if leftLen > maxCharsPerLine {
			penalty += float64((leftLen - maxCharsPerLine) * 4)
		}
		if rightLen > maxCharsPerLine {
			penalty += float64((rightLen - maxCharsPerLine) * 4)
		}
		if penalty < bestScore {
			bestScore = penalty
			bestSplit = i
		}
	}
	left := strings.TrimSpace(strings.Join(words[:bestSplit], " "))
	right := strings.TrimSpace(strings.Join(words[bestSplit:], " "))
	if left == "" || right == "" {
		return strings.TrimSpace(text)
	}
	return left + "\n" + right
}

func splitCaptionChunks(text string, wordsPerChunk, maxCharsPerChunk int) []string {
	words := strings.Fields(strings.ReplaceAll(text, "\n", " "))
	if len(words) == 0 {
		return nil
	}
	if wordsPerChunk < 2 {
		wordsPerChunk = 2
	}
	if maxCharsPerChunk < 18 {
		maxCharsPerChunk = 18
	}
	chunks := make([]string, 0, (len(words)+wordsPerChunk-1)/wordsPerChunk)
	current := make([]string, 0, wordsPerChunk)
	flush := func(force bool) {
		if len(current) == 0 {
			return
		}
		if !force && len(current) < 2 {
			return
		}
		chunks = append(chunks, strings.Join(current, " "))
		current = current[:0]
	}
	for _, word := range words {
		candidate := strings.TrimSpace(strings.Join(append(current, word), " "))
		shouldFlush := len(current) > 0 && (len(current) >= wordsPerChunk || utf8.RuneCountInString(candidate) > maxCharsPerChunk)
		if shouldFlush {
			flush(true)
		}
		current = append(current, word)
		if len(current) >= 2 && hasCaptionBreakPunctuation(word) {
			flush(true)
		}
	}
	flush(true)
	return chunks
}

func hasCaptionBreakPunctuation(word string) bool {
	word = strings.TrimSpace(word)
	if word == "" {
		return false
	}
	last := word[len(word)-1]
	switch last {
	case '.', ',', '!', '?', ';', ':':
		return true
	default:
		return false
	}
}

func escapeDrawtextText(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		":", "\\:",
		"'", "\\'",
		"%", "\\%",
		",", "\\,",
		";", "\\;",
		"[", "\\[",
		"]", "\\]",
		"\r", "",
	)
	return replacer.Replace(text)
}

type fullCaptionArtifacts struct {
	SRTPath     string
	TextPath    string
	ASSPath     string
	WorkASSPath string
	Cleanup     func()
}

type renderTrimWindow struct {
	VideoStart float64
	VideoEnd   float64
	AudioStart float64
	AudioEnd   float64
	Duration   float64
}

type fullSubtitleStyle struct {
	FontName         string
	BackColourASS    string
	OutlineColourASS string
	BoxColorFFMPEG   string
	OutlineSize      float64
	BorderStyle      int
}

func buildFullInterviewCaptionChunks(utterances []AssemblyUtterance, timelineDuration float64) []captionChunk {
	if timelineDuration <= 0 {
		return nil
	}
	return buildSentenceCaptionChunks(utterances, 0, timelineDuration)
}

func buildFullTranscriptText(utterances []AssemblyUtterance, timelineDuration float64) string {
	if len(utterances) == 0 || timelineDuration <= 0 {
		return ""
	}
	lines := make([]string, 0, len(utterances))
	for _, utterance := range utterances {
		start := float64(utterance.Start) / 1000.0
		end := float64(utterance.End) / 1000.0
		if end <= 0 || start >= timelineDuration {
			continue
		}
		text := strings.TrimSpace(utterance.Text)
		if text == "" {
			continue
		}
		speaker := strings.TrimSpace(utterance.Speaker)
		if speaker != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", speaker, text))
			continue
		}
		lines = append(lines, text)
	}
	return strings.TrimSpace(strings.Join(lines, "\n\n"))
}

func formatSRTTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalMS := int(math.Round(seconds * 1000))
	hours := totalMS / 3600000
	totalMS -= hours * 3600000
	minutes := totalMS / 60000
	totalMS -= minutes * 60000
	secs := totalMS / 1000
	millis := totalMS % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}

func formatASSTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalCS := int(math.Round(seconds * 100))
	hours := totalCS / 360000
	totalCS -= hours * 360000
	minutes := totalCS / 6000
	totalCS -= minutes * 6000
	secs := totalCS / 100
	centis := totalCS % 100
	return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, secs, centis)
}

func escapeASSText(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"{", "(",
		"}", ")",
		"\r", "",
		"\n", "\\N",
	)
	return replacer.Replace(strings.TrimSpace(text))
}

func buildSRTFromCaptionChunks(chunks []captionChunk) string {
	var builder strings.Builder
	index := 0
	for _, chunk := range chunks {
		text := strings.TrimSpace(chunk.Text)
		if text == "" || chunk.End-chunk.Start < 0.15 {
			continue
		}
		index++
		builder.WriteString(strconv.Itoa(index))
		builder.WriteString("\n")
		builder.WriteString(formatSRTTimestamp(chunk.Start))
		builder.WriteString(" --> ")
		builder.WriteString(formatSRTTimestamp(chunk.End))
		builder.WriteString("\n")
		builder.WriteString(strings.TrimSpace(chunk.Text))
		builder.WriteString("\n\n")
	}
	return strings.TrimSpace(builder.String()) + "\n"
}

func buildASSFromCaptionChunks(chunks []captionChunk) string {
	return buildASSFromCaptionChunksWithStyle(chunks, 1920, 1080, resolveFullSubtitleStyle("", "", 50))
}

func normalizeFontKey(fontName string) string {
	return strings.TrimSpace(strings.ToLower(fontName))
}

func normalizeSubtitleOpacity(opacity int) int {
	if opacity < 0 {
		return 0
	}
	if opacity > 100 {
		return 100
	}
	return opacity
}

func resolveFullSubtitleStyle(fontName, bgColor string, opacity int) fullSubtitleStyle {
	allowedFonts := map[string]string{
		"segoe-ui":     "Segoe UI",
		"arial":        "Arial",
		"verdana":      "Verdana",
		"tahoma":       "Tahoma",
		"trebuchet-ms": "Trebuchet MS",
		"georgia":      "Georgia",
		"montserrat":   "Montserrat",
	}
	fontKey := strings.TrimSpace(strings.ToLower(fontName))
	resolvedFont := allowedFonts[fontKey]
	if resolvedFont == "" {
		if direct, ok := allowedFonts[strings.TrimSpace(fontName)]; ok {
			resolvedFont = direct
		}
	}
	if resolvedFont == "" {
		resolvedFont = "Segoe UI"
	}
	opacity = normalizeSubtitleOpacity(opacity)
	outlineSize := 1.4
	if fontKey == "montserrat" {
		outlineSize = 2.0
	}
	borderStyle := 3
	if opacity <= 0 {
		borderStyle = 1
	}
	return fullSubtitleStyle{
		FontName:         resolvedFont,
		BackColourASS:    assBackColourFromHex(bgColor, opacity),
		OutlineColourASS: "&H00000000",
		BoxColorFFMPEG:   ffmpegBoxColorFromHex(bgColor, opacity),
		OutlineSize:      outlineSize,
		BorderStyle:      borderStyle,
	}
}

func assBackColourFromHex(hex string, opacity int) string {
	value := strings.TrimSpace(strings.TrimPrefix(hex, "#"))
	if len(value) != 6 {
		opacity = normalizeSubtitleOpacity(opacity)
		alpha := 255 - int(math.Round(float64(opacity)*255.0/100.0))
		return fmt.Sprintf("&H%02X000000", alpha)
	}
	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		opacity = normalizeSubtitleOpacity(opacity)
		alpha := 255 - int(math.Round(float64(opacity)*255.0/100.0))
		return fmt.Sprintf("&H%02X000000", alpha)
	}
	r := byte(parsed >> 16)
	g := byte(parsed >> 8)
	b := byte(parsed)
	alpha := 255 - int(math.Round(float64(normalizeSubtitleOpacity(opacity))*255.0/100.0))
	return fmt.Sprintf("&H%02X%02X%02X%02X", alpha, b, g, r)
}

func ffmpegBoxColorFromHex(hex string, opacity int) string {
	value := strings.TrimSpace(strings.TrimPrefix(hex, "#"))
	if len(value) != 6 {
		value = "000000"
	}
	if _, err := strconv.ParseUint(value, 16, 32); err != nil {
		value = "000000"
	}
	alpha := float64(normalizeSubtitleOpacity(opacity)) / 100.0
	return fmt.Sprintf("0x%s@%.2f", strings.ToUpper(value), alpha)
}

func subtitleLayoutProfile(width, height int, fontName string) (fontSize, marginH, marginV, wordsPerLine, maxCharsPerLine int) {
	isMontserrat := normalizeFontKey(fontName) == "montserrat"
	switch {
	case height > width:
		fontSize = maxInt(36, int(math.Round(float64(height)*0.035)))
		marginH = maxInt(38, int(math.Round(float64(width)*0.04)))
		marginV = maxInt(220, int(math.Round(float64(height)*0.205)))
		wordsPerLine = 4
		maxCharsPerLine = 24
		if isMontserrat {
			fontSize = maxInt(34, int(math.Round(float64(height)*0.0315)))
			marginH = maxInt(18, int(math.Round(float64(width)*0.02)))
			marginV = maxInt(210, int(math.Round(float64(height)*0.195)))
			wordsPerLine = 5
			maxCharsPerLine = 26
		}
	case width == height:
		fontSize = maxInt(34, int(math.Round(float64(height)*0.040)))
		marginH = maxInt(72, int(math.Round(float64(width)*0.10)))
		marginV = maxInt(84, int(math.Round(float64(height)*0.11)))
		wordsPerLine = 4
		maxCharsPerLine = 24
		if isMontserrat {
			fontSize = maxInt(32, int(math.Round(float64(height)*0.037)))
			maxCharsPerLine = 22
		}
	default:
		fontSize = maxInt(34, int(math.Round(float64(height)*0.038)))
		marginH = maxInt(72, int(math.Round(float64(width)*0.10)))
		marginV = maxInt(54, int(math.Round(float64(height)*0.09)))
		wordsPerLine = 5
		maxCharsPerLine = 30
		if isMontserrat {
			fontSize = maxInt(32, int(math.Round(float64(height)*0.035)))
			maxCharsPerLine = 28
		}
	}
	return fontSize, marginH, marginV, wordsPerLine, maxCharsPerLine
}

func buildASSFromCaptionChunksWithStyle(chunks []captionChunk, width, height int, style fullSubtitleStyle) string {
	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}
	fontSize, marginH, marginV, wordsPerLine, maxCharsPerLine := subtitleLayoutProfile(width, height, style.FontName)
	var builder strings.Builder
	builder.WriteString("[Script Info]\n")
	builder.WriteString("ScriptType: v4.00+\n")
	builder.WriteString("WrapStyle: 1\n")
	builder.WriteString("ScaledBorderAndShadow: yes\n")
	builder.WriteString(fmt.Sprintf("PlayResX: %d\n", width))
	builder.WriteString(fmt.Sprintf("PlayResY: %d\n", height))
	builder.WriteString("\n[V4+ Styles]\n")
	builder.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	builder.WriteString(fmt.Sprintf("Style: Default,%s,%d,&H00FFFFFF,&H000000FF,%s,%s,0,0,0,0,100,100,0,0,%d,%.1f,0,2,%d,%d,%d,1\n", style.FontName, fontSize, style.OutlineColourASS, style.BackColourASS, style.BorderStyle, style.OutlineSize, marginH, marginH, marginV))
	builder.WriteString("\n[Events]\n")
	builder.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")
	forceTwoLines := height > width && normalizeFontKey(style.FontName) == "montserrat"
	blurTag := "{\\blur0.8}"
	for _, chunk := range chunks {
		rawText := strings.TrimSpace(chunk.Text)
		if forceTwoLines {
			rawText = forceBalancedTwoLineCaption(rawText, maxCharsPerLine)
		} else {
			rawText = wrapCaptionText(rawText, wordsPerLine, maxCharsPerLine)
		}
		text := escapeASSText(rawText)
		if text == "" || chunk.End-chunk.Start < 0.15 {
			continue
		}
		builder.WriteString("Dialogue: 0,")
		builder.WriteString(formatASSTimestamp(chunk.Start))
		builder.WriteString(",")
		builder.WriteString(formatASSTimestamp(chunk.End))
		builder.WriteString(",Default,,0,0,0,,")
		builder.WriteString(blurTag)
		builder.WriteString(text)
		builder.WriteString("\n")
	}
	return builder.String()
}

func ensureEmbeddedSubtitleFontsDir() (string, error) {
	fontRoot := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "subtitle-fonts")
	if err := os.MkdirAll(fontRoot, 0755); err != nil {
		return "", err
	}
	fontPath := filepath.Join(fontRoot, "Montserrat-Regular.ttf")
	if _, err := os.Stat(fontPath); err == nil {
		return fontRoot, nil
	}
	fontData, err := embeddedFontFiles.ReadFile("fonts/Montserrat-Regular.ttf")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(fontPath, fontData, 0644); err != nil {
		return "", err
	}
	return fontRoot, nil
}

func buildSubtitleFilterExpr(subtitlePath string) (string, error) {
	filter := fmt.Sprintf("setsar=1,subtitles='%s'", escapeFFmpegFilterPath(subtitlePath))
	fontsDir, err := ensureEmbeddedSubtitleFontsDir()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(fontsDir) != "" {
		filter += fmt.Sprintf(":fontsdir='%s'", escapeFFmpegFilterPath(fontsDir))
	}
	return filter, nil
}

func buildFullCaptionArtifacts(outputDir, videoPath string, utterances []AssemblyUtterance, timelineDuration float64, meta videoStreamMeta, style fullSubtitleStyle) (fullCaptionArtifacts, error) {
	baseName := fullCaptionsBaseName(videoPath)
	chunks := buildFullInterviewCaptionChunks(utterances, timelineDuration)
	if len(chunks) == 0 {
		return fullCaptionArtifacts{}, errors.New("unable to build subtitle chunks for the full interview")
	}
	transcriptText := buildFullTranscriptText(utterances, timelineDuration)
	if strings.TrimSpace(transcriptText) == "" {
		return fullCaptionArtifacts{}, errors.New("unable to build transcript text for the full interview")
	}

	srtData := buildSRTFromCaptionChunks(chunks)
	assData := buildASSFromCaptionChunksWithStyle(chunks, meta.Width, meta.Height, style)
	textData := transcriptText + "\n"

	srtPath := filepath.Join(outputDir, baseName+".srt")
	textPath := filepath.Join(outputDir, baseName+".txt")
	assPath := filepath.Join(outputDir, baseName+".ass")

	if err := os.WriteFile(srtPath, []byte(srtData), 0644); err != nil {
		return fullCaptionArtifacts{}, err
	}
	if err := os.WriteFile(textPath, []byte(textData), 0644); err != nil {
		return fullCaptionArtifacts{}, err
	}
	if err := os.WriteFile(assPath, []byte(assData), 0644); err != nil {
		return fullCaptionArtifacts{}, err
	}

	tempRoot := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "analysis-temp")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return fullCaptionArtifacts{}, err
	}
	tempDir, err := os.MkdirTemp(tempRoot, "full-captions-")
	if err != nil {
		return fullCaptionArtifacts{}, err
	}
	workASSPath := filepath.Join(tempDir, "captions.ass")
	if err := os.WriteFile(workASSPath, []byte(assData), 0644); err != nil {
		_ = os.RemoveAll(tempDir)
		return fullCaptionArtifacts{}, err
	}

	return fullCaptionArtifacts{
		SRTPath:     srtPath,
		TextPath:    textPath,
		ASSPath:     assPath,
		WorkASSPath: workASSPath,
		Cleanup: func() {
			_ = os.RemoveAll(tempDir)
		},
	}, nil
}

func buildShortCaptionArtifacts(utterances []AssemblyUtterance, segmentStart, segmentEnd float64, preset shortRenderPreset, style fullSubtitleStyle) (fullCaptionArtifacts, error) {
	var chunks []captionChunk
	if preset.Height > preset.Width && normalizeFontKey(style.FontName) == "montserrat" {
		chunks = buildSentenceCaptionChunks(utterances, segmentStart, segmentEnd)
	} else {
		_, wordsPerChunk, _, maxCharsPerChunk := captionLayoutProfile(preset)
		chunks = buildTimedCaptionChunks(utterances, segmentStart, segmentEnd, wordsPerChunk, maxCharsPerChunk)
	}
	if len(chunks) == 0 {
		return fullCaptionArtifacts{}, errors.New("unable to build subtitle chunks for this clip")
	}
	tempRoot := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "analysis-temp")
	if err := os.MkdirAll(tempRoot, 0755); err != nil {
		return fullCaptionArtifacts{}, err
	}
	tempDir, err := os.MkdirTemp(tempRoot, "short-captions-")
	if err != nil {
		return fullCaptionArtifacts{}, err
	}
	workASSPath := filepath.Join(tempDir, "clip.ass")
	assData := buildASSFromCaptionChunksWithStyle(chunks, preset.Width, preset.Height, style)
	if err := os.WriteFile(workASSPath, []byte(assData), 0644); err != nil {
		_ = os.RemoveAll(tempDir)
		return fullCaptionArtifacts{}, err
	}
	return fullCaptionArtifacts{
		WorkASSPath: workASSPath,
		Cleanup: func() {
			_ = os.RemoveAll(tempDir)
		},
	}, nil
}

func escapeFFmpegFilterPath(path string) string {
	replacer := strings.NewReplacer(
		"\\", "/",
		":", "\\:",
		"'", "\\'",
		"[", "\\[",
		"]", "\\]",
		",", "\\,",
		";", "\\;",
	)
	return replacer.Replace(path)
}

func computeRenderTrimWindow(segmentStart, segmentEnd, syncDelay, mediaDuration float64) (renderTrimWindow, error) {
	videoStart := math.Max(0, segmentStart-syncDelay)
	videoEnd := math.Max(videoStart+0.25, segmentEnd-syncDelay)
	if mediaDuration > 0 {
		videoEnd = math.Min(videoEnd, mediaDuration)
	}
	if videoEnd-videoStart < 0.25 {
		return renderTrimWindow{}, errors.New("clip is too short after sync trimming")
	}
	audioStart := math.Max(0, videoStart+syncDelay)
	audioEnd := math.Max(audioStart+0.25, videoEnd+syncDelay)
	duration := videoEnd - videoStart
	if audioEnd-audioStart > 0 {
		duration = math.Min(duration, audioEnd-audioStart)
	}
	return renderTrimWindow{
		VideoStart: videoStart,
		VideoEnd:   videoEnd,
		AudioStart: audioStart,
		AudioEnd:   audioEnd,
		Duration:   duration,
	}, nil
}

func (a *App) renderShorts(req shortsRenderRequest, send func(progressEvent)) (shortsRenderResponse, error) {
	outputDir := strings.TrimSpace(req.OutputDir)
	if outputDir == "" {
		outputDir = filepath.Join(filepath.Dir(req.VideoPath), strings.TrimSuffix(filepath.Base(req.VideoPath), filepath.Ext(req.VideoPath))+"_shorts")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return shortsRenderResponse{}, err
	}

	presets := resolveShortRenderPresets(req.Formats)
	if len(presets) == 0 {
		return shortsRenderResponse{}, errors.New("at least one shorts format is required")
	}

	segments := make([]shortSegment, 0, len(req.Segments))
	for _, segment := range req.Segments {
		if segment.Enabled {
			segments = append(segments, segment)
		}
	}
	if len(segments) == 0 {
		return shortsRenderResponse{}, errors.New("select at least one clip to render")
	}
	if strings.EqualFold(strings.TrimSpace(req.CaptionsMode), "burned-in") && len(req.Utterances) == 0 {
		return shortsRenderResponse{}, errors.New("captions require transcript utterances from Build plan")
	}

	crf := req.CRF
	if crf <= 0 {
		crf = 18
	}
	presetName := strings.TrimSpace(req.Preset)
	if presetName == "" {
		presetName = "medium"
	}

	stagingRoot := ensureOutputStagingRoot(outputDir)
	stagedVideoPath, cleanupVideo, err := stageInputPathForWindowsInDir(req.VideoPath, stagingRoot)
	if err != nil {
		return shortsRenderResponse{}, err
	}
	defer cleanupVideo()

	stagedAudioPath := ""
	cleanupAudio := func() {}
	if strings.TrimSpace(req.AudioPath) != "" {
		stagedAudioPath, cleanupAudio, err = stageInputPathForWindowsInDir(req.AudioPath, stagingRoot)
		if err != nil {
			return shortsRenderResponse{}, err
		}
		defer cleanupAudio()
	}

	meta, err := a.probeVideoStream(stagedVideoPath)
	if err != nil {
		return shortsRenderResponse{}, err
	}

	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath, true)
	if err != nil {
		return shortsRenderResponse{}, err
	}
	if backend.Cleanup != nil {
		defer backend.Cleanup()
	}

	started := time.Now()
	files := make([]shortsRenderedFile, 0, len(segments)*len(presets))
	failed := make([]string, 0)
	totalJobs := len(segments) * len(presets)
	job := 0
	for index, segment := range segments {
		baseName := shortOutputBaseName(segment, index+1)
		for _, renderPreset := range presets {
			job++
			if send != nil {
				send(progressEvent{
					Percent: round((float64(job-1)/float64(totalJobs))*100, 1),
					Message: fmt.Sprintf("Shorts: rendering %d/%d -> %s (%s)", job, totalJobs, segment.Title, renderPreset.ID),
				})
			}
			file, err := a.renderShortFile(stagedVideoPath, stagedAudioPath, outputDir, stagingRoot, meta, segment, renderPreset, req, baseName, crf, presetName, backend, send)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s / %s: %v", segment.Title, renderPreset.ID, err))
				continue
			}
			files = append(files, file)
		}
	}

	planPath, planErr := saveShortsPlanFile(outputDir, req, files, failed)
	if planErr != nil {
		failed = append(failed, fmt.Sprintf("plan.json: %v", planErr))
	}
	if len(files) == 0 {
		if len(failed) > 0 {
			return shortsRenderResponse{}, errors.New(strings.Join(failed, "\n"))
		}
		return shortsRenderResponse{}, errors.New("shorts render produced no files")
	}

	return shortsRenderResponse{
		OutputDir:     outputDir,
		PlanPath:      planPath,
		Files:         files,
		Duration:      time.Since(started).String(),
		RenderedCount: len(files),
		Failed:        failed,
	}, nil
}

func fullCaptionsBaseName(videoPath string) string {
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	if strings.TrimSpace(base) == "" {
		base = "interview"
	}
	return base + "_subtitled"
}

func (a *App) renderFullCaptions(req fullCaptionsRenderRequest, send func(progressEvent)) (fullCaptionsRenderResponse, error) {
	ctx, cancel := a.beginCancelableTask()
	defer a.endCancelableTask(cancel)

	if strings.TrimSpace(req.AssemblyAIKey) == "" {
		return fullCaptionsRenderResponse{}, errors.New("AssemblyAI key is required")
	}
	if !strings.EqualFold(strings.TrimSpace(req.CaptionsMode), "burned-in") {
		return fullCaptionsRenderResponse{}, errors.New("captions must be enabled")
	}

	outputDir := strings.TrimSpace(req.OutputDir)
	if outputDir == "" {
		outputDir = filepath.Join(filepath.Dir(req.VideoPath), strings.TrimSuffix(filepath.Base(req.VideoPath), filepath.Ext(req.VideoPath))+"_shorts")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fullCaptionsRenderResponse{}, err
	}

	transcriptSource := strings.TrimSpace(req.AudioPath)
	transcriptSourceLabel := "video"
	syncDelay := 0.0
	timelineDuration := 0.0
	videoDurationHint := 0.0
	if send != nil {
		send(progressEvent{Message: "Captions: validating full interview sources..."})
	}
	if transcriptSource == "" {
		transcriptSource = strings.TrimSpace(req.VideoPath)
		timelineDuration, _ = a.probeDuration(req.VideoPath)
	} else {
		transcriptSourceLabel = "master-audio"
		if send != nil {
			send(progressEvent{Message: "Captions: measuring sync between video and master audio..."})
		}
		metrics, err := a.analyzeSync(req.VideoPath, req.AudioPath, defaultAnalyzeSeconds, defaultMaxLagSeconds)
		if err != nil {
			return fullCaptionsRenderResponse{}, fmt.Errorf("failed to align video with master audio for captions: %w", err)
		}
		syncDelay = metrics.DelaySeconds
		timelineDuration = metrics.AudioDuration
		videoDurationHint = metrics.VideoDuration
		if timelineDuration <= 0 {
			timelineDuration, _ = a.probeDuration(req.AudioPath)
		}
	}
	if timelineDuration <= 0 {
		timelineDuration, _ = a.probeDuration(transcriptSource)
	}

	if send != nil {
		send(progressEvent{Message: "Captions: requesting transcript from AssemblyAI..."})
	}
	utterances, err := a.transcribeWithAssemblyAI(ctx, transcriptSource, req.AssemblyAIKey, func(event progressEvent) {
		if send == nil {
			return
		}
		send(progressEvent{Percent: event.Percent, Message: event.Message})
	})
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}
	if len(utterances) == 0 {
		return fullCaptionsRenderResponse{}, errors.New("AssemblyAI returned no transcript utterances")
	}

	stagingRoot := ensureOutputStagingRoot(outputDir)
	stagedVideoPath, cleanupVideo, err := stageInputPathForWindowsInDir(req.VideoPath, stagingRoot)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}
	defer cleanupVideo()

	stagedAudioPath := ""
	cleanupAudio := func() {}
	if strings.TrimSpace(req.AudioPath) != "" {
		stagedAudioPath, cleanupAudio, err = stageInputPathForWindowsInDir(req.AudioPath, stagingRoot)
		if err != nil {
			return fullCaptionsRenderResponse{}, err
		}
		defer cleanupAudio()
	}

	meta, err := a.probeVideoStream(stagedVideoPath)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}
	if timelineDuration <= 0 {
		timelineDuration = meta.Duration
	}
	if timelineDuration <= 0 && videoDurationHint > 0 {
		timelineDuration = videoDurationHint
	}
	if timelineDuration <= 0 && stagedAudioPath != "" {
		timelineDuration, _ = a.probeDuration(stagedAudioPath)
	}
	if timelineDuration <= 0 {
		return fullCaptionsRenderResponse{}, errors.New("unable to determine full interview duration")
	}
	trimWindow, err := computeRenderTrimWindow(0, timelineDuration, syncDelay, meta.Duration)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}

	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath, true)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}
	if backend.Cleanup != nil {
		defer backend.Cleanup()
	}

	crf := req.CRF
	if crf <= 0 {
		crf = 18
	}
	presetName := strings.TrimSpace(req.Preset)
	if presetName == "" {
		presetName = "medium"
	}

	if send != nil {
		send(progressEvent{Message: "Captions: building subtitle files..."})
	}
	style := resolveFullSubtitleStyle(req.SubtitleFont, req.SubtitleBgColor, req.SubtitleBgOpacity)
	artifacts, err := buildFullCaptionArtifacts(outputDir, req.VideoPath, utterances, trimWindow.Duration, meta, style)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}
	if artifacts.Cleanup != nil {
		defer artifacts.Cleanup()
	}

	if send != nil {
		send(progressEvent{Message: "Captions: rendering full interview subtitles..."})
	}
	started := time.Now()
	outputPath, err := a.renderFullCaptionsFile(
		stagedVideoPath,
		stagedAudioPath,
		outputDir,
		stagingRoot,
		meta,
		trimWindow,
		artifacts.WorkASSPath,
		fullCaptionsBaseName(req.VideoPath),
		crf,
		presetName,
		backend,
		send,
	)
	if err != nil {
		return fullCaptionsRenderResponse{}, err
	}

	return fullCaptionsRenderResponse{
		OutputPath:       outputPath,
		Duration:         time.Since(started).String(),
		TranscriptSource: transcriptSourceLabel,
		SRTPath:          artifacts.SRTPath,
		TextPath:         artifacts.TextPath,
		ASSPath:          artifacts.ASSPath,
	}, nil
}

func (a *App) renderFullCaptionsFile(stagedVideoPath, stagedAudioPath, outputDir, stagingRoot string, meta videoStreamMeta, trimWindow renderTrimWindow, subtitlePath, baseName string, crf int, presetName string, backend executionPlan, send func(progressEvent)) (string, error) {
	outputPath := filepath.Join(outputDir, baseName+".mp4")
	stagedOutputPath, finalizeOutput, cleanupOutput, err := stageOutputPathForWindows(outputPath, stagingRoot)
	if err != nil {
		return "", err
	}
	defer cleanupOutput()

	subtitleFilter, err := buildSubtitleFilterExpr(subtitlePath)
	if err != nil {
		return "", err
	}
	ffmpegArgs := []string{"-y"}
	totalSeconds := trimWindow.Duration
	if stagedAudioPath == "" {
		ffmpegArgs = append(ffmpegArgs,
			"-ss", trimFloat(trimWindow.VideoStart, 3),
			"-to", trimFloat(trimWindow.VideoEnd, 3),
			"-i", stagedVideoPath,
			"-map", "0:v:0",
			"-map", "0:a?",
			"-vf", subtitleFilter,
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, presetName)...)
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+faststart",
			"-shortest",
			"-progress", "pipe:1",
			"-nostats",
			stagedOutputPath,
		)
	} else {
		filterComplex := fmt.Sprintf(
			"[0:v]trim=start=%s:end=%s,setpts=PTS-STARTPTS,%s[v];[1:a]atrim=start=%s:end=%s,asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[a]",
			trimFloat(trimWindow.VideoStart, 3),
			trimFloat(trimWindow.VideoEnd, 3),
			subtitleFilter,
			trimFloat(trimWindow.AudioStart, 3),
			trimFloat(trimWindow.AudioEnd, 3),
		)
		ffmpegArgs = append(ffmpegArgs,
			"-i", stagedVideoPath,
			"-i", stagedAudioPath,
			"-filter_complex", filterComplex,
			"-map", "[v]",
			"-map", "[a]",
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, presetName)...)
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+faststart",
			"-shortest",
			"-progress", "pipe:1",
			"-nostats",
			stagedOutputPath,
		)
	}

	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)
	if err := a.runFFmpegCommand(backend.Executable, args, totalSeconds, send); err != nil {
		return "", err
	}
	if err := finalizeOutput(); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (a *App) renderShortFile(stagedVideoPath, stagedAudioPath, outputDir, stagingRoot string, meta videoStreamMeta, segment shortSegment, renderPreset shortRenderPreset, req shortsRenderRequest, baseName string, crf int, presetName string, backend executionPlan, send func(progressEvent)) (shortsRenderedFile, error) {
	videoStart := math.Max(0, segment.Start-req.SyncDelaySeconds)
	videoEnd := math.Max(videoStart+0.25, segment.End-req.SyncDelaySeconds)
	if meta.Duration > 0 {
		videoEnd = math.Min(videoEnd, meta.Duration)
	}
	if videoEnd-videoStart < 0.25 {
		return shortsRenderedFile{}, errors.New("clip is too short after sync trimming")
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.mp4", baseName, renderPreset.FileSuffix))
	if segment.ID == "full-interview" && renderPreset.ID == "source-original" {
		outputPath = filepath.Join(outputDir, baseName+".mp4")
	}
	stagedOutputPath, finalizeOutput, cleanupOutput, err := stageOutputPathForWindows(outputPath, stagingRoot)
	if err != nil {
		return shortsRenderedFile{}, err
	}
	defer cleanupOutput()

	effectivePreset := renderPreset
	if effectivePreset.Width <= 0 || effectivePreset.Height <= 0 {
		effectivePreset.Width = meta.Width
		effectivePreset.Height = meta.Height
		if effectivePreset.Width <= 0 || effectivePreset.Height <= 0 {
			effectivePreset.Width = 1920
			effectivePreset.Height = 1080
		}
	}
	style := resolveFullSubtitleStyle(req.SubtitleFont, req.SubtitleBgColor, req.SubtitleBgOpacity)
	videoFilter := buildShortVideoFilter(effectivePreset)
	if strings.EqualFold(strings.TrimSpace(req.CaptionsMode), "burned-in") && len(req.Utterances) > 0 {
		captionArtifacts, err := buildShortCaptionArtifacts(req.Utterances, segment.Start, segment.End, effectivePreset, style)
		if err != nil {
			return shortsRenderedFile{}, err
		}
		if captionArtifacts.Cleanup != nil {
			defer captionArtifacts.Cleanup()
		}
		subtitleFilter, err := buildSubtitleFilterExpr(captionArtifacts.WorkASSPath)
		if err != nil {
			return shortsRenderedFile{}, err
		}
		videoFilter = strings.TrimSpace(videoFilter + "," + subtitleFilter)
	}
	ffmpegArgs := []string{"-y"}
	totalSeconds := videoEnd - videoStart
	if stagedAudioPath == "" {
		ffmpegArgs = append(ffmpegArgs,
			"-ss", trimFloat(videoStart, 3),
			"-to", trimFloat(videoEnd, 3),
			"-i", stagedVideoPath,
			"-map", "0:v:0",
			"-map", "0:a?",
		)
		ffmpegArgs = append(ffmpegArgs, "-vf", videoFilter)
		ffmpegArgs = append(ffmpegArgs, "-pix_fmt", "yuv420p")
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, presetName)...)
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+faststart",
			"-shortest",
			"-progress", "pipe:1",
			"-nostats",
			stagedOutputPath,
		)
	} else {
		audioStart := math.Max(0, videoStart+req.SyncDelaySeconds)
		audioEnd := math.Max(audioStart+0.25, videoEnd+req.SyncDelaySeconds)
		totalSeconds = math.Min(totalSeconds, audioEnd-audioStart)
		filterComplex := fmt.Sprintf(
			"[0:v]trim=start=%s:end=%s,setpts=PTS-STARTPTS,%s[v];[1:a]atrim=start=%s:end=%s,asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[a]",
			trimFloat(videoStart, 3),
			trimFloat(videoEnd, 3),
			videoFilter,
			trimFloat(audioStart, 3),
			trimFloat(audioEnd, 3),
		)
		ffmpegArgs = append(ffmpegArgs,
			"-i", stagedVideoPath,
			"-i", stagedAudioPath,
		)
		ffmpegArgs = append(ffmpegArgs, "-filter_complex", filterComplex)
		ffmpegArgs = append(ffmpegArgs,
			"-map", "[v]",
			"-map", "[a]",
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, presetName)...)
		ffmpegArgs = append(ffmpegArgs,
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+faststart",
			"-shortest",
			"-progress", "pipe:1",
			"-nostats",
			stagedOutputPath,
		)
	}

	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)
	if err := a.runFFmpegCommand(backend.Executable, args, totalSeconds, send); err != nil {
		return shortsRenderedFile{}, err
	}
	if err := finalizeOutput(); err != nil {
		return shortsRenderedFile{}, err
	}
	return shortsRenderedFile{
		SegmentID: segment.ID,
		Title:     segment.Title,
		Format:    renderPreset.ID,
		Output:    outputPath,
		Start:     round(segment.Start, 3),
		End:       round(segment.End, 3),
	}, nil
}

func saveShortsPlanFile(outputDir string, req shortsRenderRequest, files []shortsRenderedFile, failed []string) (string, error) {
	planFileName := "plan.json"
	if len(req.Segments) == 1 && req.Segments[0].ID == "full-interview" {
		planFileName = "full_interview_plan.json"
	}
	planPath := filepath.Join(outputDir, planFileName)
	payload := map[string]any{
		"videoPath":        req.VideoPath,
		"audioPath":        req.AudioPath,
		"outputDir":        outputDir,
		"segments":         req.Segments,
		"formats":          req.Formats,
		"captionsMode":     req.CaptionsMode,
		"syncDelaySeconds": req.SyncDelaySeconds,
		"files":            files,
		"failed":           failed,
		"savedAt":          time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(planPath, data, 0644); err != nil {
		return "", err
	}
	return planPath, nil
}

func (a *App) probeDuration(path string) (float64, error) {
	cmd := newCommand(a.ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return 0, errors.New(msg)
	}
	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return 0, errors.New("ffprobe returned empty duration")
	}
	duration, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func (a *App) probeVideoStream(path string) (videoStreamMeta, error) {
	cmd := newCommand(a.ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate:stream_tags=rotate:stream_side_data=rotation:format=duration",
		"-of", "json",
		path,
	)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return videoStreamMeta{}, errors.New(msg)
	}

	var payload struct {
		Streams []struct {
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			Tags       struct {
				Rotate string `json:"rotate"`
			} `json:"tags"`
			SideDataList []struct {
				Rotation float64 `json:"rotation"`
			} `json:"side_data_list"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return videoStreamMeta{}, err
	}
	if len(payload.Streams) == 0 {
		return videoStreamMeta{}, errors.New("ffprobe returned no video stream")
	}

	meta := videoStreamMeta{
		Width:  payload.Streams[0].Width,
		Height: payload.Streams[0].Height,
		FPS:    parseFFprobeRate(payload.Streams[0].RFrameRate),
	}
	if payload.Streams[0].Tags.Rotate != "" {
		if rotation, err := strconv.ParseFloat(payload.Streams[0].Tags.Rotate, 64); err == nil {
			meta.Rotation = rotation
		}
	}
	if meta.Rotation == 0 && len(payload.Streams[0].SideDataList) > 0 {
		meta.Rotation = payload.Streams[0].SideDataList[0].Rotation
	}
	if int(math.Abs(meta.Rotation))%180 == 90 {
		meta.Width, meta.Height = meta.Height, meta.Width
	}
	if payload.Format.Duration != "" {
		if duration, err := strconv.ParseFloat(payload.Format.Duration, 64); err == nil {
			meta.Duration = duration
		}
	}
	if meta.FPS <= 0 {
		meta.FPS = 25
	}
	return meta, nil
}

func (a *App) ensureTools() error {
	if a.ffmpegPath == "" {
		return errors.New("ffmpeg not found in PATH")
	}
	if a.ffprobePath == "" {
		return errors.New("ffprobe not found in PATH")
	}
	return nil
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (a *App) writeError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, apiError{Error: message})
}

func streamJSON(w http.ResponseWriter, fn func(send func(progressEvent))) {
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	send := func(event progressEvent) {
		_ = encoder.Encode(event)
		flusher.Flush()
	}
	fn(send)
	time.Sleep(150 * time.Millisecond)
}

func (a *App) runFFmpegCommand(executable string, args []string, totalSeconds float64, send func(progressEvent)) error {
	cmd := newCommand(executable, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.currentCmd = cmd
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		if a.currentCmd == cmd {
			a.currentCmd = nil
		}
		a.mu.Unlock()
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	progressDone := make(chan struct{})
	var lastMessage string
	go func() {
		scanner := bufio.NewScanner(stdout)
		progress := map[string]string{}
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			progress[parts[0]] = parts[1]
			if parts[0] == "progress" {
				if send != nil {
					send(parseFFmpegProgress(progress, totalSeconds))
				}
				progress = map[string]string{}
			}
		}
	}()
	go func() {
		defer close(progressDone)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if event, ok := parseFFmpegStderrProgress(line, totalSeconds); ok {
				lastMessage = event.Message
				if send != nil {
					send(event)
				}
				continue
			}
			if shouldIgnoreFFmpegLogLine(line) {
				continue
			}
			lastMessage = line
			if send != nil {
				send(progressEvent{Message: line})
			}
		}
	}()

	err = cmd.Wait()
	<-progressDone
	if err != nil {
		if lastMessage != "" {
			return errors.New(lastMessage)
		}
		return err
	}
	return nil
}

func parseFFmpegProgress(values map[string]string, totalSeconds float64) progressEvent {
	outTime := parseFFmpegProgressTime(values["out_time_ms"], values["out_time"])
	percent := 0.0
	if totalSeconds > 0 {
		percent = math.Min(100, (outTime/totalSeconds)*100)
	}
	timeValue := values["out_time"]
	if timeValue == "" {
		timeValue = "00:00:00.000000"
	}
	speedValue := values["speed"]
	if speedValue == "" {
		speedValue = "-"
	}
	message := fmt.Sprintf("ffmpeg %.1f%% | time=%s | speed=%s", percent, timeValue, speedValue)
	return progressEvent{
		Percent: percent,
		Message: message,
	}
}

func parseFFmpegProgressTime(outTimeMS, outTime string) float64 {
	if outTimeMS != "" {
		if ms, err := strconv.ParseFloat(outTimeMS, 64); err == nil {
			return ms / 1000000
		}
	}
	parts := strings.Split(outTime, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)
	return hours*3600 + minutes*60 + seconds
}

func parseFFmpegStderrProgress(line string, totalSeconds float64) (progressEvent, bool) {
	if !strings.Contains(line, "time=") {
		return progressEvent{}, false
	}
	timeValue := extractFFmpegKV(line, "time")
	if timeValue == "" {
		return progressEvent{}, false
	}
	outTime := parseFFmpegProgressTime("", timeValue)
	percent := 0.0
	if totalSeconds > 0 {
		percent = math.Min(100, (outTime/totalSeconds)*100)
	}
	speedValue := extractFFmpegKV(line, "speed")
	if speedValue == "" {
		speedValue = "-"
	}
	return progressEvent{
		Percent: percent,
		Message: fmt.Sprintf("ffmpeg %.1f%% | time=%s | speed=%s", percent, timeValue, speedValue),
	}, true
}

func extractFFmpegKV(line, key string) string {
	index := strings.Index(line, key+"=")
	if index < 0 {
		return ""
	}
	value := strings.TrimSpace(line[index+len(key)+1:])
	if value == "" {
		return ""
	}
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return strings.TrimSpace(fields[0])
}

func shouldIgnoreFFmpegLogLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return true
	}
	if strings.HasPrefix(lower, "video:") || strings.Contains(lower, "muxing overhead") {
		return true
	}
	return false
}

func validateExistingFile(path, field string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !strings.Contains(path, `\`) && !strings.Contains(path, `/`) {
		return fmt.Errorf("%s must be a full file path, not just a file name", field)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s does not exist", field)
	}
	if info.IsDir() {
		return fmt.Errorf("%s must point to a file", field)
	}
	return nil
}

func describeDelay(delay float64) string {
	if math.Abs(delay) < 0.010 {
		return "Почти идеально: сдвиг меньше 10 мс."
	}
	if delay > 0 {
		return fmt.Sprintf("Мастер-аудио стартует раньше видео примерно на %.0f мс. Для точного sync нужно подрезать начало внешнего аудио.", delay*1000)
	}
	return fmt.Sprintf("Видео стартует раньше мастер-аудио примерно на %.0f мс. Для точного sync нужно подрезать начало видео.", math.Abs(delay*1000))
}

func buildRenderSummary(delay float64) string {
	if delay >= 0 {
		return "Точный рендер будет выполнять `atrim` для внешнего аудио и полное перекодирование видео вместо `-c:v copy`."
	}
	return "Точный рендер будет выполнять `trim/setpts` для видео и полное перекодирование, чтобы не зависеть от keyframe."
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func round(value float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(value*pow) / pow
}

func trimFloat(value float64, digits int) string {
	return strconv.FormatFloat(round(value, digits), 'f', -1, 64)
}

func resolveSyncOutputPath(videoPath, rawOutput string) string {
	outputPath := strings.TrimSpace(rawOutput)
	ext := filepath.Ext(videoPath)
	if ext == "" {
		ext = ".mp4"
	}
	defaultName := buildSyncOutputName(videoPath, ext)
	if outputPath == "" {
		return filepath.Join(filepath.Dir(videoPath), defaultName)
	}
	if looksLikeDirectoryPath(outputPath) {
		return filepath.Join(outputPath, defaultName)
	}
	return outputPath
}

func resolveMulticamOutputPath(masterAudioPath, rawOutput string) string {
	outputPath := strings.TrimSpace(rawOutput)
	defaultName := strings.TrimSuffix(filepath.Base(masterAudioPath), filepath.Ext(masterAudioPath)) + "_multicam.mp4"
	if outputPath == "" {
		return filepath.Join(filepath.Dir(masterAudioPath), defaultName)
	}
	if looksLikeDirectoryPath(outputPath) {
		return filepath.Join(outputPath, defaultName)
	}
	return outputPath
}

func stageInputPathForWindows(path string) (string, func(), error) {
	if runtime.GOOS != "windows" || !containsNonASCII(path) {
		return path, func() {}, nil
	}

	if shortPath := getWindowsShortPath(path); shortPath != "" && !containsNonASCII(shortPath) {
		return shortPath, func() {}, nil
	}

	stagingRoot := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "analysis-temp")
	if err := os.MkdirAll(stagingRoot, 0755); err != nil {
		return "", nil, err
	}

	tempDir, err := os.MkdirTemp(stagingRoot, "autosyncstudio-input-")
	if err != nil {
		return "", nil, err
	}

	targetPath := filepath.Join(tempDir, "input"+filepath.Ext(path))
	if err := copyFile(path, targetPath); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", nil, err
	}

	return targetPath, func() { _ = os.RemoveAll(tempDir) }, nil
}

func stageInputPathForWindowsInDir(path, stagingRoot string) (string, func(), error) {
	if runtime.GOOS != "windows" || !containsNonASCII(path) {
		return path, func() {}, nil
	}
	if shortPath := getWindowsShortPath(path); shortPath != "" && !containsNonASCII(shortPath) {
		return shortPath, func() {}, nil
	}
	if err := os.MkdirAll(stagingRoot, 0755); err != nil {
		return "", nil, err
	}
	tempDir, err := os.MkdirTemp(stagingRoot, "input-")
	if err != nil {
		return "", nil, err
	}
	targetPath := filepath.Join(tempDir, "input"+filepath.Ext(path))
	if err := copyFile(path, targetPath); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", nil, err
	}
	return targetPath, func() {
		_ = os.RemoveAll(tempDir)
		cleanupDirectoryIfEmpty(stagingRoot)
	}, nil
}

func ensureOutputStagingRoot(outputDir string) string {
	return filepath.Join(outputDir, ".autosync-temp")
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

func appSettingsPath() string {
	return filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "settings.json")
}

func loadAppSettings() (appSettings, error) {
	path := appSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return appSettings{}, nil
		}
		return appSettings{}, err
	}
	var settings appSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return appSettings{}, err
	}
	settings.AssemblyAIKey, err = decryptStoredSecret(settings.AssemblyAIKey)
	if err != nil {
		return appSettings{}, err
	}
	settings.GeminiAIKey, err = decryptStoredSecret(settings.GeminiAIKey)
	if err != nil {
		return appSettings{}, err
	}
	settings.OpenAIKey, err = decryptStoredSecret(settings.OpenAIKey)
	if err != nil {
		return appSettings{}, err
	}
	settings.AIKey, err = decryptStoredSecret(settings.AIKey)
	if err != nil {
		return appSettings{}, err
	}
	if settings.GeminiAIKey == "" && settings.AIKey != "" {
		settings.GeminiAIKey = settings.AIKey
	}
	if settings.OpenAIKey == "" && settings.AIKey != "" {
		settings.OpenAIKey = settings.AIKey
	}
	return settings, nil
}

func saveAppSettings(settings appSettings) error {
	path := appSettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	storedSettings := settings
	var err error
	storedSettings.AssemblyAIKey, err = encryptStoredSecret(settings.AssemblyAIKey)
	if err != nil {
		return err
	}
	storedSettings.GeminiAIKey, err = encryptStoredSecret(settings.GeminiAIKey)
	if err != nil {
		return err
	}
	storedSettings.OpenAIKey, err = encryptStoredSecret(settings.OpenAIKey)
	if err != nil {
		return err
	}
	storedSettings.AIKey = ""
	data, err := json.MarshalIndent(storedSettings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func stageOutputPathForWindows(path, stagingRoot string) (string, func() error, func(), error) {
	if runtime.GOOS != "windows" {
		return path, func() error { return nil }, func() {}, nil
	}
	if !containsNonASCII(path) {
		return path, func() error { return nil }, func() {}, nil
	}
	parentDir := filepath.Dir(path)
	fileName := filepath.Base(path)
	if shortParent := getWindowsShortPath(parentDir); shortParent != "" && !containsNonASCII(shortParent) && !containsNonASCII(fileName) {
		directPath := filepath.Join(shortParent, fileName)
		return directPath, func() error { return nil }, func() {}, nil
	}
	if err := os.MkdirAll(stagingRoot, 0755); err != nil {
		return "", nil, nil, err
	}

	tempDir, err := os.MkdirTemp(stagingRoot, "output-")
	if err != nil {
		return "", nil, nil, err
	}

	stagedPath := filepath.Join(tempDir, "output"+filepath.Ext(path))
	finalize := func() error {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		return copyFile(stagedPath, path)
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
		cleanupDirectoryIfEmpty(stagingRoot)
	}
	return stagedPath, finalize, cleanup, nil
}

func buildSyncOutputName(videoPath, ext string) string {
	baseName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	return fmt.Sprintf("%s_%s_sync%s", baseName, fileTimestampTag(videoPath), ext)
}

func fileTimestampTag(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now().Format("20060102_150405")
	}
	return info.ModTime().Format("20060102_150405")
}

func cleanupDirectoryIfEmpty(path string) {
	entries, err := os.ReadDir(path)
	if err != nil || len(entries) != 0 {
		return
	}
	_ = os.Remove(path)
}

func looksLikeDirectoryPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if strings.HasSuffix(path, `\`) || strings.HasSuffix(path, `/`) {
		return true
	}
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}
	return filepath.Ext(path) == ""
}

func containsNonASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Close()
}

func windowsPickFile(kind string) (string, error) {
	filter := "All files (*.*)|*.*"
	multiselect := "$false"
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "video":
		filter = "Video files (*.mp4;*.mov;*.mkv;*.mxf;*.avi)|*.mp4;*.mov;*.mkv;*.mxf;*.avi|All files (*.*)|*.*"
	case "audio", "master-audio":
		filter = "Audio files (*.wav;*.mp3;*.m4a;*.aac;*.flac)|*.wav;*.mp3;*.m4a;*.aac;*.flac|All files (*.*)|*.*"
	case "camera-multi", "cameras":
		filter = "Video files (*.mp4;*.mov;*.mkv;*.mxf;*.avi)|*.mp4;*.mov;*.mkv;*.mxf;*.avi|All files (*.*)|*.*"
		multiselect = "$true"
	}

	script := fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms\n$dialog = New-Object System.Windows.Forms.OpenFileDialog\n$dialog.Filter = '%s'\n$dialog.Multiselect = %s\nif ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {\n  $selected = if ($dialog.Multiselect) { [string]::Join(\"`n\", $dialog.FileNames) } else { $dialog.FileName }\n  $bytes = [System.Text.Encoding]::UTF8.GetBytes($selected)\n  Write-Output ([Convert]::ToBase64String($bytes))\n}", strings.ReplaceAll(filter, `'`, `''`), multiselect)

	cmd := newCommand("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-STA", "-Command", script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func windowsPickDirectory() (string, error) {
	script := `Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Filter = 'Folders|*.none'
$dialog.CheckFileExists = $false
$dialog.CheckPathExists = $true
$dialog.ValidateNames = $false
$dialog.FileName = 'Выбрать папку'
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  $selected = Split-Path -Parent $dialog.FileName
  if ([string]::IsNullOrWhiteSpace($selected)) {
    $selected = $dialog.FileName
  }
  $bytes = [System.Text.Encoding]::UTF8.GetBytes($selected)
  Write-Output ([Convert]::ToBase64String($bytes))
}`

	cmd := newCommand("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-STA", "-Command", script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func windowsPickSave(kind, currentPath string) (string, error) {
	filter := "MP4 files (*.mp4)|*.mp4|All files (*.*)|*.*"
	defaultExt := "mp4"
	fileName := ""
	initialDir := ""

	currentPath = strings.TrimSpace(currentPath)
	if currentPath != "" {
		if looksLikeDirectoryPath(currentPath) {
			initialDir = currentPath
		} else {
			initialDir = filepath.Dir(currentPath)
			fileName = filepath.Base(currentPath)
		}
	}

	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "multicam-output":
		if fileName == "" {
			fileName = "multicam_result.mp4"
		}
	default:
		if fileName == "" {
			fileName = "result.mp4"
		}
	}

	script := fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms\n$dialog = New-Object System.Windows.Forms.SaveFileDialog\n$dialog.Filter = '%s'\n$dialog.DefaultExt = '%s'\n$dialog.AddExtension = $true\n$dialog.OverwritePrompt = $true\n$dialog.FileName = '%s'\nif ('%s' -ne '') { $dialog.InitialDirectory = '%s' }\nif ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {\n  $bytes = [System.Text.Encoding]::UTF8.GetBytes($dialog.FileName)\n  Write-Output ([Convert]::ToBase64String($bytes))\n}",
		strings.ReplaceAll(filter, `'`, `''`),
		strings.ReplaceAll(defaultExt, `'`, `''`),
		strings.ReplaceAll(fileName, `'`, `''`),
		strings.ReplaceAll(initialDir, `'`, `''`),
		strings.ReplaceAll(initialDir, `'`, `''`),
	)

	cmd := newCommand("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-STA", "-Command", script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func buildCameraAlignPlan(cameraPath string, delaySeconds float64, outputDir, preset string, crf int, confidence float64, backend executionPlan) multicamExportPlan {
	baseName := strings.TrimSuffix(filepath.Base(cameraPath), filepath.Ext(cameraPath))
	targetDir := outputDir
	if targetDir == "" {
		targetDir = filepath.Dir(cameraPath)
	}
	outputPath := filepath.Join(targetDir, baseName+"_aligned.mp4")

	ffmpegArgs := []string{"-y", "-i", cameraPath}
	strategy := ""
	if delaySeconds >= 0 {
		filter := fmt.Sprintf("tpad=start_duration=%.6f:start_mode=add:color=black,setpts=PTS-STARTPTS", delaySeconds)
		ffmpegArgs = append(ffmpegArgs,
			"-vf", filter,
			"-an",
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
		ffmpegArgs = append(ffmpegArgs, outputPath)
		strategy = "Камера стартует позже мастера: команда добавляет черный lead-in через tpad и сохраняет video-only mezzanine."
	} else {
		filter := fmt.Sprintf("trim=start=%.6f,setpts=PTS-STARTPTS", math.Abs(delaySeconds))
		ffmpegArgs = append(ffmpegArgs,
			"-vf", filter,
			"-an",
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
		ffmpegArgs = append(ffmpegArgs, outputPath)
		strategy = "Камера стартует раньше мастера: команда подрезает начало и делает точный video-only mezzanine."
	}

	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)

	return multicamExportPlan{
		Path:         cameraPath,
		DelaySeconds: round(delaySeconds, 3),
		DelayMs:      int(math.Round(delaySeconds * 1000)),
		Confidence:   round(confidence, 3),
		OutputPath:   outputPath,
		Strategy:     strategy,
		Command:      shellJoin(append([]string{backend.Executable}, args...)),
	}
}

func (a *App) renderAlignedCamera(inputPath, outputPath string, delaySeconds float64, preset string, crf int, backend executionPlan) error {
	ffmpegArgs := []string{"-y", "-i", inputPath}
	if delaySeconds >= 0 {
		filter := fmt.Sprintf("tpad=start_duration=%.6f:start_mode=add:color=black,setpts=PTS-STARTPTS", delaySeconds)
		ffmpegArgs = append(ffmpegArgs,
			"-vf", filter,
			"-an",
			"-pix_fmt", "yuv420p",
		)
	} else {
		filter := fmt.Sprintf("trim=start=%.6f,setpts=PTS-STARTPTS", math.Abs(delaySeconds))
		ffmpegArgs = append(ffmpegArgs,
			"-vf", filter,
			"-an",
			"-pix_fmt", "yuv420p",
		)
	}
	ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
	ffmpegArgs = append(ffmpegArgs, outputPath)

	args := append([]string{}, backend.PrefixArgs...)
	args = append(args, ffmpegArgs...)
	if backend.Cleanup != nil {
		defer backend.Cleanup()
	}
	return a.runFFmpegCommand(backend.Executable, args, 0, nil)
}

func (a *App) resolveExecutionPlan(mode, remoteAddress, remoteSecret, remoteClientPath string, ephemeralRemoteConfig bool) (executionPlan, error) {
	normalized := strings.TrimSpace(strings.ToLower(mode))
	if normalized == "" {
		normalized = "cpu"
	}
	switch normalized {
	case "cpu", "local-cpu":
		if a.ffmpegPath == "" {
			return executionPlan{}, errors.New("ffmpeg not found in PATH")
		}
		return executionPlan{Mode: "cpu", Executable: a.ffmpegPath}, nil
	case "gpu", "local-gpu":
		if a.ffmpegPath == "" {
			return executionPlan{}, errors.New("ffmpeg not found in PATH")
		}
		if !a.supportsNVENC() {
			return executionPlan{}, errors.New("local GPU mode requires ffmpeg with h264_nvenc support")
		}
		return executionPlan{Mode: "gpu", Executable: a.ffmpegPath}, nil
	case "remote", "ffmpeg-over-ip", "remote-gpu":
		clientPath := strings.TrimSpace(remoteClientPath)
		if clientPath == "" && runtime.GOOS == "windows" {
			if tools, err := windowsbundle.EnsureStudioTools(); err == nil && tools.ClientPath != "" {
				clientPath = tools.ClientPath
			}
		}
		if clientPath == "" {
			clientPath = findBinary("ffmpeg-over-ip-client")
		}
		if clientPath == "" && runtime.GOOS == "windows" {
			if tools, err := windowsbundle.EnsureStudioTools(); err == nil && tools.ClientPath != "" {
				clientPath = tools.ClientPath
			}
		}
		if clientPath == "" {
			return executionPlan{}, errors.New("remote mode requires ffmpeg-over-ip-client рядом с программой или в PATH")
		}
		if strings.TrimSpace(remoteAddress) == "" {
			return executionPlan{}, errors.New("remoteAddress is required for ffmpeg-over-ip mode")
		}
		if strings.TrimSpace(remoteSecret) == "" {
			return executionPlan{}, errors.New("remoteSecret is required for ffmpeg-over-ip mode")
		}
		configPath := ""
		cleanup := func() {}
		if ephemeralRemoteConfig {
			var err error
			configPath, cleanup, err = writeTemporaryFFmpegOverIPClientConfig(strings.TrimSpace(remoteAddress), strings.TrimSpace(remoteSecret))
			if err != nil {
				return executionPlan{}, err
			}
		} else {
			var err error
			configPath, err = writeFFmpegOverIPClientConfig(strings.TrimSpace(remoteAddress), strings.TrimSpace(remoteSecret))
			if err != nil {
				return executionPlan{}, err
			}
		}
		prefixArgs := []string{"--config", configPath}
		return executionPlan{
			Mode:       "remote",
			Executable: clientPath,
			PrefixArgs: prefixArgs,
			Cleanup:    cleanup,
		}, nil
	default:
		return executionPlan{}, fmt.Errorf("unknown executionMode: %s", mode)
	}
}

func (a *App) inspectBackendStatus(req backendStatusRequest) backendStatusResponse {
	mode := strings.TrimSpace(strings.ToLower(req.ExecutionMode))
	if mode == "" {
		mode = "cpu"
	}

	switch mode {
	case "cpu", "local-cpu":
		ready := a.ffmpegPath != ""
		return backendStatusResponse{
			Mode:          "cpu",
			OverallStatus: ternaryStatus(ready, "ok", "error"),
			ModeLabel:     "Local CPU",
			ClientStatus:  ternaryText(ready, "ffmpeg ready", "ffmpeg missing"),
			ServerStatus:  "No remote connection required",
			BackendReady:  ready,
			ClientFound:   ready,
			Message:       ternaryText(ready, "Local CPU backend is ready", "ffmpeg not found in PATH"),
		}
	case "gpu", "local-gpu":
		ffmpegReady := a.ffmpegPath != ""
		nvencReady := ffmpegReady && a.supportsNVENC()
		return backendStatusResponse{
			Mode:          "gpu",
			OverallStatus: ternaryStatus(nvencReady, "ok", "error"),
			ModeLabel:     "Local GPU",
			ClientStatus:  ternaryText(ffmpegReady, "ffmpeg ready", "ffmpeg missing"),
			ServerStatus:  ternaryText(nvencReady, "NVENC ready", "NVENC unavailable"),
			BackendReady:  nvencReady,
			ClientFound:   ffmpegReady,
			Message:       ternaryText(nvencReady, "Local GPU backend is ready", "Local GPU mode requires ffmpeg with h264_nvenc support"),
		}
	default:
		clientPath := strings.TrimSpace(req.RemoteClientPath)
		if clientPath == "" && runtime.GOOS == "windows" {
			if tools, err := windowsbundle.EnsureStudioTools(); err == nil && tools.ClientPath != "" {
				clientPath = tools.ClientPath
			}
		}
		if clientPath == "" {
			clientPath = findBinary("ffmpeg-over-ip-client")
		}
		if clientPath == "" && runtime.GOOS == "windows" {
			if tools, err := windowsbundle.EnsureStudioTools(); err == nil && tools.ClientPath != "" {
				clientPath = tools.ClientPath
			}
		}

		address := strings.TrimSpace(req.RemoteAddress)
		clientFound := clientPath != ""
		serverReachable := false
		serverStatus := "Address is required"
		overallStatus := "error"
		message := "Remote backend is not ready"

		switch {
		case address == "":
			serverStatus = "Address is required"
		case !isDialableTCPAddress(address):
			serverStatus = "Address must be host:port"
		default:
			serverReachable = canReachTCP(address, 1500*time.Millisecond)
			serverStatus = ternaryText(serverReachable, "Server reachable", "Server is offline or blocked")
		}

		secretReady := strings.TrimSpace(req.RemoteSecret) != ""
		clientStatus := ternaryText(clientFound, "Managed client ready", "Managed client missing")
		if !secretReady {
			clientStatus = "Secret is required"
		}

		backendReady := clientFound && secretReady && serverReachable
		if backendReady {
			overallStatus = "ok"
			message = "Remote backend is ready"
		} else if clientFound || serverReachable || secretReady {
			overallStatus = "warn"
			message = "Remote backend needs attention"
		}

		return backendStatusResponse{
			Mode:            "remote",
			OverallStatus:   overallStatus,
			ModeLabel:       "Remote ffmpeg-over-ip",
			ClientStatus:    clientStatus,
			ServerStatus:    serverStatus,
			BackendReady:    backendReady,
			ServerReachable: serverReachable,
			ClientFound:     clientFound,
			ResolvedClient:  clientPath,
			ResolvedAddress: address,
			Message:         message,
		}
	}
}

func isDialableTCPAddress(address string) bool {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil {
		return false
	}
	if strings.TrimSpace(host) == "" {
		return false
	}
	portValue, err := strconv.Atoi(port)
	return err == nil && portValue >= 1 && portValue <= 65535
}

func canReachTCP(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func ternaryText(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func ternaryStatus(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func comparablePath(path string) string {
	normalized := filepath.Clean(strings.TrimSpace(path))
	if runtime.GOOS == "windows" {
		return strings.ToLower(normalized)
	}
	return normalized
}

func measuredMetricsForPath(path string, measured []multicamCameraResult) (syncMetrics, bool) {
	target := comparablePath(path)
	for _, item := range measured {
		if comparablePath(item.Path) != target {
			continue
		}
		if item.Duration <= 0 {
			break
		}
		return syncMetrics{
			DelaySeconds:  item.DelaySeconds,
			Confidence:    item.Confidence,
			VideoDuration: item.Duration,
		}, true
	}
	return syncMetrics{}, false
}

func videoCodecForMode(mode string) string {
	if mode == "gpu" || mode == "remote" {
		return "h264_nvenc"
	}
	return "libx264"
}

func (a *App) supportsNVENC() bool {
	if a.ffmpegPath == "" {
		return false
	}
	cmd := newCommand(a.ffmpegPath, "-hide_banner", "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_nvenc")
}

func videoPresetForMode(mode, requested string) string {
	if mode == "gpu" || mode == "remote" {
		switch requested {
		case "slow", "medium":
			return "p5"
		case "fast":
			return "p4"
		case "veryfast":
			return "p3"
		default:
			return "p5"
		}
	}
	return requested
}

func videoEncodeArgsForMode(mode string, crf int, preset string) []string {
	if mode == "gpu" || mode == "remote" {
		return []string{
			"-c:v", videoCodecForMode(mode),
			"-preset", videoPresetForMode(mode, preset),
			"-rc", "vbr",
			"-cq", strconv.Itoa(crf),
			"-b:v", "0",
		}
	}
	return []string{
		"-c:v", videoCodecForMode(mode),
		"-preset", videoPresetForMode(mode, preset),
		"-crf", strconv.Itoa(crf),
	}
}

func writeFFmpegOverIPClientConfig(address, secret string) (string, error) {
	content := fmt.Sprintf("{\n  \"address\": %q,\n  \"authSecret\": %q\n}\n", address, secret)
	root := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime")
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(root, "autosync.ffmpeg-over-ip.client.jsonc")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", err
	}
	return path, nil
}

func writeTemporaryFFmpegOverIPClientConfig(address, secret string) (string, func(), error) {
	content := fmt.Sprintf("{\n  \"address\": %q,\n  \"authSecret\": %q\n}\n", address, secret)
	root := filepath.Join(runtimeWorkspaceRoot(), ".autosync-runtime", "temp")
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", func() {}, err
	}
	file, err := os.CreateTemp(root, "ffmpeg-over-ip-*.jsonc")
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	if _, err := file.WriteString(content); err != nil {
		file.Close()
		_ = os.Remove(path)
		return "", func() {}, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", func() {}, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		_ = os.Remove(path)
		return "", func() {}, err
	}
	return path, func() {
		_ = os.Remove(path)
	}, nil
}

func shellJoin(parts []string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, shellQuote(part))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func parseFFprobeRate(value string) float64 {
	if value == "" {
		return 0
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		f, _ := strconv.ParseFloat(value, 64)
		return f
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return 0
	}
	return num / den
}
