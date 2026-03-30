package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
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

	"autosyncstudio/internal/bundles"
	windowsbundle "autosyncstudio/third_party/windows"
)

//go:embed index.html main.js
var staticFiles embed.FS

const serverAppAddr = "127.0.0.1:8521"

type App struct {
	addr string

	mu           sync.Mutex
	serverConfig serverConfig
	serverCmd    *exec.Cmd
	startedAt    time.Time
	lastExit     string
	logs         []string
}

type serverConfig struct {
	ServerBinary string   `json:"serverBinary"`
	FFmpegPath   string   `json:"ffmpegPath"`
	Address      string   `json:"address"`
	AuthSecret   string   `json:"authSecret"`
	LogMode      string   `json:"logMode"`
	Debug        bool     `json:"debug"`
	Rewrites     []string `json:"rewrites"`
}

type apiError struct {
	Error string `json:"error"`
}

type statusResponse struct {
	AppName           string                   `json:"appName"`
	Address           string                   `json:"address"`
	Running           bool                     `json:"running"`
	PID               int                      `json:"pid"`
	Uptime            string                   `json:"uptime"`
	LastExit          string                   `json:"lastExit"`
	ServerConfig      serverConfig             `json:"serverConfig"`
	ServerConfigPath  string                   `json:"serverConfigPath"`
	ConnectedClients  []clientInfo             `json:"connectedClients"`
	ActiveJobs        []jobInfo                `json:"activeJobs"`
	LogTail           []string                 `json:"logTail"`
	BundledPlatform   string                   `json:"bundledPlatform"`
	BundledComponents []bundles.NamedComponent `json:"bundledComponents"`
}

type clientInfo struct {
	RemoteAddress string `json:"remoteAddress"`
	State         string `json:"state"`
}

type jobInfo struct {
	PID         int    `json:"pid"`
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
}

func NewApp() *App {
	cfg := serverConfig{
		ServerBinary: "ffmpeg-over-ip-server.exe",
		FFmpegPath:   "ffmpeg.exe",
		Address:      "0.0.0.0:5050",
		AuthSecret:   "change-me",
		LogMode:      "stdout",
		Debug:        true,
		Rewrites: []string{
			`["h264_nvenc","h264_qsv"]`,
			`["hevc_nvenc","hevc_qsv"]`,
		},
	}
	if runtime.GOOS == "windows" {
		if tools, err := windowsbundle.EnsureServerTools(); err == nil {
			if tools.ServerBinary != "" {
				cfg.ServerBinary = tools.ServerBinary
			}
			if tools.FFmpegPath != "" {
				cfg.FFmpegPath = tools.FFmpegPath
			}
		}
	}
	return &App{
		addr:         serverAppAddr,
		serverConfig: cfg,
	}
}

func (a *App) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleIndex)
	mux.HandleFunc("/main.js", a.handleMainJS)
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/start", a.handleStart)
	mux.HandleFunc("/api/stop", a.handleStop)
	mux.HandleFunc("/api/config", a.handleConfig)

	log.Printf("AutoSync Render Server UI is ready at http://%s\n", a.addr)
	return http.ListenAndServe(a.addr, mux)
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

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	a.writeJSON(w, http.StatusOK, a.currentStatus())
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var cfg serverConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateServerConfig(cfg); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	a.mu.Lock()
	a.serverConfig = cfg
	a.mu.Unlock()

	a.writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *App) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var cfg serverConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validateServerConfig(cfg); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.startServer(cfg); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, a.currentStatus())
}

func (a *App) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := a.stopServer(); err != nil {
		a.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.writeJSON(w, http.StatusOK, a.currentStatus())
}

func (a *App) startServer(cfg serverConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.serverCmd != nil && a.serverCmd.Process != nil {
		return errors.New("server is already running")
	}

	configPath, err := writeServerConfig(cfg)
	if err != nil {
		return err
	}

	cmd := exec.Command(cfg.ServerBinary, "--config", configPath)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	a.serverConfig = cfg
	a.serverCmd = cmd
	a.startedAt = time.Now()
	a.lastExit = ""
	a.appendLog(fmt.Sprintf("[%s] Server starting with PID %d", time.Now().Format(time.RFC3339), cmd.Process.Pid))

	go a.consumePipe(stdoutPipe, "stdout")
	go a.consumePipe(stderrPipe, "stderr")
	go a.waitServerProcess(cmd)
	return nil
}

func (a *App) stopServer() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.serverCmd == nil || a.serverCmd.Process == nil {
		return nil
	}
	err := a.serverCmd.Process.Kill()
	if err == nil {
		a.appendLog(fmt.Sprintf("[%s] Stop requested", time.Now().Format(time.RFC3339)))
	}
	return err
}

func (a *App) waitServerProcess(cmd *exec.Cmd) {
	err := cmd.Wait()
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		a.lastExit = err.Error()
		a.appendLog(fmt.Sprintf("[%s] Server exited with error: %v", time.Now().Format(time.RFC3339), err))
	} else {
		a.lastExit = "clean exit"
		a.appendLog(fmt.Sprintf("[%s] Server exited cleanly", time.Now().Format(time.RFC3339)))
	}
	if a.serverCmd == cmd {
		a.serverCmd = nil
	}
}

func (a *App) consumePipe(pipe io.ReadCloser, source string) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		a.mu.Lock()
		a.appendLog(fmt.Sprintf("[%s] %s: %s", time.Now().Format(time.RFC3339), source, scanner.Text()))
		a.mu.Unlock()
	}
}

func (a *App) appendLog(line string) {
	a.logs = append(a.logs, line)
	if len(a.logs) > 300 {
		a.logs = a.logs[len(a.logs)-300:]
	}
}

func (a *App) currentStatus() statusResponse {
	a.mu.Lock()
	cfg := a.serverConfig
	var pid int
	running := false
	var uptime string
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		pid = a.serverCmd.Process.Pid
		running = true
		uptime = time.Since(a.startedAt).Round(time.Second).String()
	}
	lastExit := a.lastExit
	logTail := append([]string(nil), a.logs...)
	a.mu.Unlock()

	var clients []clientInfo
	var jobs []jobInfo
	if running {
		clients = detectConnectedClients(cfg.Address, pid)
		jobs = detectActiveJobs(pid)
	}

	return statusResponse{
		AppName:           "AutoSync Render Server",
		Address:           a.addr,
		Running:           running,
		PID:               pid,
		Uptime:            uptime,
		LastExit:          lastExit,
		ServerConfig:      cfg,
		ServerConfigPath:  serverConfigPath(),
		ConnectedClients:  clients,
		ActiveJobs:        jobs,
		LogTail:           tailStrings(logTail, 80),
		BundledPlatform:   "windows-amd64",
		BundledComponents: bundles.ComponentsForPlatform("windows-amd64"),
	}
}

func writeServerConfig(cfg serverConfig) (string, error) {
	rewriteEntries := make([]string, 0, len(cfg.Rewrites))
	for _, rewrite := range cfg.Rewrites {
		rewrite = strings.TrimSpace(rewrite)
		if rewrite == "" {
			continue
		}
		rewriteEntries = append(rewriteEntries, "    "+rewrite)
	}
	rewriteBlock := ""
	if len(rewriteEntries) > 0 {
		rewriteBlock = strings.Join(rewriteEntries, ",\n")
	}

	content := fmt.Sprintf("{\n  \"log\": %q,\n  \"debug\": %t,\n  \"address\": %q,\n  \"authSecret\": %q,\n  \"ffmpegPath\": %q,\n  \"rewrites\": [\n%s\n  ]\n}\n",
		cfg.LogMode,
		cfg.Debug,
		cfg.Address,
		cfg.AuthSecret,
		cfg.FFmpegPath,
		rewriteBlock,
	)
	path := serverConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", err
	}
	return path, nil
}

func serverConfigPath() string {
	return filepath.Join(os.TempDir(), "autosync.ffmpeg-over-ip.server.jsonc")
}

func validateServerConfig(cfg serverConfig) error {
	if strings.TrimSpace(cfg.ServerBinary) == "" {
		return errors.New("serverBinary is required")
	}
	if strings.TrimSpace(cfg.FFmpegPath) == "" {
		return errors.New("ffmpegPath is required")
	}
	if strings.TrimSpace(cfg.Address) == "" {
		return errors.New("address is required")
	}
	if strings.TrimSpace(cfg.AuthSecret) == "" {
		return errors.New("authSecret is required")
	}
	if strings.TrimSpace(cfg.LogMode) == "" {
		cfg.LogMode = "stdout"
	}
	return nil
}

func detectConnectedClients(address string, serverPID int) []clientInfo {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		if strings.Count(address, ":") == 1 {
			parts := strings.Split(address, ":")
			host = parts[0]
			port = parts[1]
		} else {
			return nil
		}
	}
	_ = host

	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		return nil
	}

	var clients []clientInfo
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		localAddr := fields[1]
		remoteAddr := fields[2]
		state := fields[3]
		pidField := fields[len(fields)-1]
		pid, _ := strconv.Atoi(pidField)
		if pid != serverPID {
			continue
		}
		if !strings.HasSuffix(localAddr, ":"+port) {
			continue
		}
		if strings.EqualFold(state, "ESTABLISHED") {
			clients = append(clients, clientInfo{RemoteAddress: remoteAddr, State: state})
		}
	}
	sort.Slice(clients, func(i, j int) bool { return clients[i].RemoteAddress < clients[j].RemoteAddress })
	return clients
}

func detectActiveJobs(serverPID int) []jobInfo {
	psScript := fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'; $pid=%d; Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -eq $pid } | Select-Object ProcessId,Name,CommandLine | ConvertTo-Json -Compress`, serverPID)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		return nil
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil
	}

	type psProcess struct {
		ProcessID   int    `json:"ProcessId"`
		Name        string `json:"Name"`
		CommandLine string `json:"CommandLine"`
	}

	var single psProcess
	var many []psProcess
	jobs := make([]jobInfo, 0)
	if strings.HasPrefix(output, "{") {
		if err := json.Unmarshal([]byte(output), &single); err == nil && single.ProcessID != 0 {
			jobs = append(jobs, jobInfo{PID: single.ProcessID, Name: single.Name, CommandLine: single.CommandLine})
		}
	} else {
		if err := json.Unmarshal([]byte(output), &many); err == nil {
			for _, item := range many {
				if item.ProcessID == 0 {
					continue
				}
				jobs = append(jobs, jobInfo{PID: item.ProcessID, Name: item.Name, CommandLine: item.CommandLine})
			}
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].PID < jobs[j].PID })
	return jobs
}

func tailStrings(values []string, max int) []string {
	if len(values) <= max {
		return values
	}
	return values[len(values)-max:]
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (a *App) writeError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, apiError{Error: message})
}
