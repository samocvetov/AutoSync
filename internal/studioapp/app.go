package studioapp

import (
	"bufio"
	"bytes"
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

	"autosyncstudio/internal/appmeta"
	"autosyncstudio/internal/bundles"
	windowsbundle "autosyncstudio/third_party/windows"
)

//go:embed index.html main.js
var staticFiles embed.FS

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
}

type progressEvent struct {
	Percent    float64               `json:"percent,omitempty"`
	Message    string                `json:"message,omitempty"`
	Done       bool                  `json:"done,omitempty"`
	Error      string                `json:"error,omitempty"`
	OutputPath string                `json:"outputPath,omitempty"`
	Duration   string                `json:"duration,omitempty"`
	Command    string                `json:"command,omitempty"`
	Shots      []multicamShotSummary `json:"shots,omitempty"`
	TotalTime  float64               `json:"totalTime,omitempty"`
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
	AIKey         string `json:"aiKey,omitempty"`
}

type systemInfoResponse struct {
	Name              string                   `json:"name"`
	Version           string                   `json:"version"`
	Address           string                   `json:"address"`
	FFmpegPath        string                   `json:"ffmpegPath"`
	FFprobePath       string                   `json:"ffprobePath"`
	BundledPlatform   string                   `json:"bundledPlatform"`
	BundledComponents []bundles.NamedComponent `json:"bundledComponents"`
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
	MasterAudioPath  string   `json:"masterAudioPath"`
	CameraPaths      []string `json:"cameraPaths"`
	AnalyzeSeconds   float64  `json:"analyzeSeconds"`
	MaxLagSeconds    float64  `json:"maxLagSeconds"`
	OutputDir        string   `json:"outputDir"`
	CRF              int      `json:"crf"`
	Preset           string   `json:"preset"`
	ExecutionMode    string   `json:"executionMode"`
	RemoteAddress    string   `json:"remoteAddress"`
	RemoteSecret     string   `json:"remoteSecret"`
	RemoteClientPath string   `json:"remoteClientPath"`
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
	MasterAudioPath   string   `json:"masterAudioPath"`
	CameraPaths       []string `json:"cameraPaths"`
	AnalyzeSeconds    float64  `json:"analyzeSeconds"`
	MaxLagSeconds     float64  `json:"maxLagSeconds"`
	OutputPath        string   `json:"outputPath"`
	PreviewSeconds    float64  `json:"previewSeconds"`
	CRF               int      `json:"crf"`
	Preset            string   `json:"preset"`
	ExecutionMode     string   `json:"executionMode"`
	RemoteAddress     string   `json:"remoteAddress"`
	RemoteSecret      string   `json:"remoteSecret"`
	RemoteClientPath  string   `json:"remoteClientPath"`
	ShotWindowSeconds float64  `json:"shotWindowSeconds"`
	MinShotSeconds    float64  `json:"minShotSeconds"`
	PrimaryCamera     int      `json:"primaryCamera"`
	EditMode          string   `json:"editMode"`
	AssemblyAIKey     string   `json:"assemblyAiKey"`
	AIProvider        string   `json:"aiProvider"`
	AIKey             string   `json:"aiKey"`
	AIPrompt          string   `json:"aiPrompt"`
}

type multicamRenderResponse struct {
	OutputPath   string                `json:"outputPath"`
	Duration     string                `json:"duration"`
	Command      string                `json:"command"`
	Shots        []multicamShotSummary `json:"shots"`
	TotalSeconds float64               `json:"totalSeconds"`
}

type shortsPlanRequest struct {
	MasterAudioPath string   `json:"masterAudioPath"`
	CameraPaths     []string `json:"cameraPaths"`
	AnalyzeSeconds  float64  `json:"analyzeSeconds"`
	MaxLagSeconds   float64  `json:"maxLagSeconds"`
	PrimaryCamera   int      `json:"primaryCamera"`
	AssemblyAIKey   string   `json:"assemblyAiKey"`
	AIProvider      string   `json:"aiProvider"`
	AIKey           string   `json:"aiKey"`
	AIPrompt        string   `json:"aiPrompt"`
	ShortsCount     int      `json:"shortsCount"`
}

type shortSegment struct {
	Title      string  `json:"title"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Reason     string  `json:"reason"`
	CameraHint int     `json:"cameraHint"`
	Command    string  `json:"command"`
}

type shortsPlanResponse struct {
	Provider string         `json:"provider"`
	Segments []shortSegment `json:"segments"`
	Note     string         `json:"note"`
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
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

type AssemblyUtterance struct {
	Speaker string `json:"speaker"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Text    string `json:"text"`
}

type AssemblyPollRes struct {
	Status     string              `json:"status"`
	Error      string              `json:"error"`
	Utterances []AssemblyUtterance `json:"utterances"`
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
	})
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
		req.AIKey = strings.TrimSpace(req.AIKey)
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
	a.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		a.writeJSON(w, http.StatusOK, map[string]string{"status": "idle"})
		return
	}
	if err := cmd.Process.Kill(); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
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

	planBackend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath)
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

		metrics, err := a.analyzeSync(path, req.MasterAudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
		if err != nil {
			a.writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: %v", path, err))
			return
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
	if err := validateExistingFile(req.MasterAudioPath, "masterAudioPath"); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.CameraPaths) == 0 {
		a.writeError(w, http.StatusBadRequest, "cameraPaths must contain at least one path")
		return
	}

	segments, note, provider, err := a.planShorts(req)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, shortsPlanResponse{
		Provider: provider,
		Segments: segments,
		Note:     note,
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
	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath)
	if err != nil {
		return "", "", 0, nil, 0, err
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
		metrics, err := a.analyzeSync(originalPath, req.MasterAudioPath, req.AnalyzeSeconds, req.MaxLagSeconds)
		if err != nil {
			return "", "", 0, nil, 0, fmt.Errorf("%s: %w", originalPath, err)
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
		utterances, err := a.transcribeWithAssemblyAI(req.MasterAudioPath, req.AssemblyAIKey, send)
		if err != nil {
			return "", "", 0, nil, 0, err
		}
		shots = buildSpeakerShotPlan(utterances, totalSeconds, len(analyses), primaryIndex, minShot)
	}
	if len(shots) == 0 {
		shots = buildShotPlan(analyses, totalSeconds, shotWindow, minShot, primaryIndex)
	}
	if len(shots) == 0 {
		return "", "", 0, nil, 0, errors.New("failed to build shot plan")
	}

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
		if sourceEnd <= 0 {
			continue
		}

		var segmentFilter string
		label := fmt.Sprintf("v%d", i)
		duration := shot.End - shot.Start
		scalePart := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=%s", referenceMeta.Width, referenceMeta.Height, referenceMeta.Width, referenceMeta.Height, fpsValue)
		if sourceStart < 0 {
			leadIn := math.Abs(sourceStart)
			segmentFilter = fmt.Sprintf("[%d:v]trim=start=0:end=%s,%s,setpts=PTS-STARTPTS,tpad=start_duration=%s:start_mode=clone,trim=start=0:end=%s[%s]",
				shot.CameraIndex,
				trimFloat(sourceEnd, 6),
				scalePart,
				trimFloat(leadIn, 6),
				trimFloat(duration, 6),
				label,
			)
		} else {
			segmentFilter = fmt.Sprintf("[%d:v]trim=start=%s:end=%s,%s,setpts=PTS-STARTPTS[%s]",
				shot.CameraIndex,
				trimFloat(sourceStart, 6),
				trimFloat(sourceEnd, 6),
				scalePart,
				label,
			)
		}
		filterParts = append(filterParts, segmentFilter)
		concatInputs = append(concatInputs, fmt.Sprintf("[%s]", label))
	}
	if len(concatInputs) == 0 {
		return "", "", 0, nil, 0, errors.New("no renderable multicam segments")
	}

	filterParts = append(filterParts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vout]", strings.Join(concatInputs, ""), len(concatInputs)))
	audioIndex := len(analyses)
	filterParts = append(filterParts, fmt.Sprintf("[%d:a]atrim=start=0:end=%s,asetpts=PTS-STARTPTS,aresample=async=1:first_pts=0[aout]", audioIndex, trimFloat(totalSeconds, 6)))

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
			bestIndex = primaryIndex
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
	alignedStart := camera.Metrics.DelaySeconds
	alignedEnd := camera.Metrics.DelaySeconds + camera.Meta.Duration
	return end > alignedStart && start < alignedEnd
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

func buildSpeakerShotPlan(utterances []AssemblyUtterance, totalSeconds float64, cameraCount, primaryIndex int, minShot float64) []shotSegment {
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
	secondaryCamera := primaryIndex
	for idx := 0; idx < cameraCount; idx++ {
		if idx != primaryIndex {
			secondaryCamera = idx
			break
		}
	}
	utteranceThreshold := math.Max(minShot, 2.0)
	mapSpeaker := func(speaker string, duration float64) int {
		speaker = strings.TrimSpace(speaker)
		if speaker == "" || speaker == primarySpeaker {
			return primaryIndex
		}
		if duration < utteranceThreshold {
			return primaryIndex
		}
		return secondaryCamera
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
			segments = append(segments, shotSegment{CameraIndex: primaryIndex, Start: cursor, End: start})
		}
		cam := mapSpeaker(utterance.Speaker, end-start)
		segments = append(segments, shotSegment{CameraIndex: cam, Start: start, End: end})
		cursor = end
	}
	if cursor < totalSeconds {
		segments = append(segments, shotSegment{CameraIndex: primaryIndex, Start: cursor, End: totalSeconds})
	}

	merged := make([]shotSegment, 0, len(segments))
	for _, segment := range segments {
		if len(merged) > 0 && merged[len(merged)-1].CameraIndex == segment.CameraIndex {
			merged[len(merged)-1].End = segment.End
			continue
		}
		merged = append(merged, segment)
	}
	return smoothShotPlan(merged, minShot, primaryIndex)
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

	backend, err := a.resolveExecutionPlan(req.ExecutionMode, req.RemoteAddress, req.RemoteSecret, req.RemoteClientPath)
	if err != nil {
		return "", "", 0, err
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

func (a *App) transcribeWithAssemblyAI(audioPath, apiKey string, send func(progressEvent)) ([]AssemblyUtterance, error) {
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

	fData, err := os.ReadFile(tempWav)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api.assemblyai.com/v2/upload", bytes.NewReader(fData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", apiKey)
	resp, err := (&http.Client{Timeout: 10 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	uploadBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var upRes AssemblyUploadRes
	if err := json.Unmarshal(uploadBody, &upRes); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(upRes.Error)
		if message == "" {
			message = strings.TrimSpace(string(uploadBody))
		}
		if message == "" {
			message = resp.Status
		}
		return nil, fmt.Errorf("AssemblyAI upload failed: %s", message)
	}
	if strings.TrimSpace(upRes.UploadURL) == "" {
		if strings.TrimSpace(upRes.Error) != "" {
			return nil, fmt.Errorf("AssemblyAI upload failed: %s", strings.TrimSpace(upRes.Error))
		}
		return nil, errors.New("AssemblyAI upload failed")
	}

	body := fmt.Sprintf(`{"audio_url":"%s","speaker_labels":true,"speech_models":["universal-3-pro","universal-2"],"language_detection":true}`, upRes.UploadURL)
	req2, err := http.NewRequest("POST", "https://api.assemblyai.com/v2/transcript", strings.NewReader(body))
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
	for {
		if send != nil {
			send(progressEvent{Message: fmt.Sprintf("AI: AssemblyAI thinking... %ds", int(time.Since(started).Seconds()))})
		}
		time.Sleep(3 * time.Second)

		pollReq, err := http.NewRequest("GET", "https://api.assemblyai.com/v2/transcript/"+trRes.ID, nil)
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

func (a *App) planShorts(req shortsPlanRequest) ([]shortSegment, string, string, error) {
	if strings.TrimSpace(req.AssemblyAIKey) == "" {
		return nil, "", "", errors.New("для Shorts нужен ключ AssemblyAI")
	}
	utterances, err := a.transcribeWithAssemblyAI(req.MasterAudioPath, req.AssemblyAIKey, nil)
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
	return settings, nil
}

func saveAppSettings(settings appSettings) error {
	path := appSettingsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
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
		filter := fmt.Sprintf("tpad=start_duration=%.6f:start_mode=clone,setpts=PTS-STARTPTS", delaySeconds)
		ffmpegArgs = append(ffmpegArgs,
			"-vf", filter,
			"-an",
			"-pix_fmt", "yuv420p",
		)
		ffmpegArgs = append(ffmpegArgs, videoEncodeArgsForMode(backend.Mode, crf, preset)...)
		ffmpegArgs = append(ffmpegArgs, outputPath)
		strategy = "Камера стартует позже мастера: команда добавляет lead-in через tpad и сохраняет video-only mezzanine."
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

func (a *App) resolveExecutionPlan(mode, remoteAddress, remoteSecret, remoteClientPath string) (executionPlan, error) {
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
		return executionPlan{Mode: "gpu", Executable: a.ffmpegPath}, nil
	case "remote", "ffmpeg-over-ip", "remote-gpu":
		clientPath := strings.TrimSpace(remoteClientPath)
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
		configPath, err := writeFFmpegOverIPClientConfig(strings.TrimSpace(remoteAddress), strings.TrimSpace(remoteSecret))
		if err != nil {
			return executionPlan{}, err
		}
		prefixArgs := []string{"--config", configPath}
		return executionPlan{
			Mode:       "remote",
			Executable: clientPath,
			PrefixArgs: prefixArgs,
		}, nil
	default:
		return executionPlan{}, fmt.Errorf("unknown executionMode: %s", mode)
	}
}

func videoCodecForMode(mode string) string {
	if mode == "gpu" || mode == "remote" {
		return "h264_nvenc"
	}
	return "libx264"
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
