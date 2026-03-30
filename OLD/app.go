package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed ffmpeg/ffmpeg.exe
var ffmpegLocalBinary []byte

//go:embed ffmpeg-over-ip-server/ffmpeg.exe
var ffmpegServerBinary []byte

//go:embed ffmpeg-over-ip-server/ffprobe.exe
var ffprobeBinary []byte

//go:embed ffmpeg-over-ip-client/ffmpeg-over-ip-client.exe
var ffmpegOverIPClientBinary []byte

//go:embed ffmpeg-over-ip-server/ffmpeg-over-ip-server.exe
var ffmpegOverIPServerBinary []byte

type App struct {
	ctx              context.Context
	processCtx       context.Context
	cancelFn         context.CancelFunc
	ffmpegPath       string
	ffprobePath      string
	ffmpegClientPath string
	ffmpegServerPath string
	serverCmd        *exec.Cmd
	
	botCtx           context.Context
	botCancel        context.CancelFunc
	
	pendingFiles     map[int64]string
	pendingFilesMu   sync.Mutex
}

func NewApp() *App { 
	return &App{
		pendingFiles: make(map[int64]string),
	} 
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.processCtx, a.cancelFn = context.WithCancel(ctx)
	
	tempDir := os.TempDir()
	
	a.ffmpegPath = filepath.Join(tempDir, "autosync_ffmpeg.exe")
	os.WriteFile(a.ffmpegPath, ffmpegLocalBinary, 0755)

	a.ffprobePath = filepath.Join(tempDir, "ffprobe.exe")
	os.WriteFile(a.ffprobePath, ffprobeBinary, 0755)

	a.ffmpegClientPath = filepath.Join(tempDir, "ffmpeg-over-ip.exe")
	os.WriteFile(a.ffmpegClientPath, ffmpegOverIPClientBinary, 0755)

	a.ffmpegServerPath = filepath.Join(tempDir, "ffmpeg-over-ip-server.exe")
	os.WriteFile(a.ffmpegServerPath, ffmpegOverIPServerBinary, 0755)
	
	serverFfmpegPath := filepath.Join(tempDir, "ffmpeg.exe")
	os.WriteFile(serverFfmpegPath, ffmpegServerBinary, 0755)
}

func (a *App) shutdown(ctx context.Context) {
	a.CancelProcess()
	a.StopServer()
	a.StopTgBot()
}

func (a *App) CancelProcess() {
	if a.cancelFn != nil { a.cancelFn() }
	exec.Command("taskkill", "/F", "/IM", "autosync_ffmpeg.exe", "/T").Run()
	exec.Command("taskkill", "/F", "/IM", "ffmpeg-over-ip.exe", "/T").Run()
}

type AssemblyUploadRes struct { UploadURL string `json:"upload_url"` }
type AssemblyTranscriptRes struct { ID string `json:"id"` }
type AssemblyUtterance struct { Speaker string `json:"speaker"`; Start int `json:"start"`; End int `json:"end"`; Text string `json:"text"` }
type AssemblyPollRes struct { 
	Status     string              `json:"status"`
	Error      string              `json:"error"`
	Utterances []AssemblyUtterance `json:"utterances"`
}
type Chunk struct { Cam int; StartA float64; EndA float64 }

func (a *App) getVideoDuration(filePath string) float64 { cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, "-i", filePath); cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}; var stderr bytes.Buffer; cmd.Stderr = &stderr; cmd.Run(); re := regexp.MustCompile(`Duration:\s+(\d{2}):(\d{2}):([\d\.]+)`); m := re.FindStringSubmatch(stderr.String()); if len(m) == 4 { h, _ := strconv.ParseFloat(m[1], 64); min, _ := strconv.ParseFloat(m[2], 64); s, _ := strconv.ParseFloat(m[3], 64); return h*3600 + min*60 + s }; return 0.1 }
func (a *App) getVideoFPS(filePath string) int { cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, "-i", filePath); cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}; var stderr bytes.Buffer; cmd.Stderr = &stderr; cmd.Run(); re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s+fps`); m := re.FindStringSubmatch(stderr.String()); if len(m) > 1 { f, _ := strconv.ParseFloat(m[1], 64); return int(math.Round(f)) }; return 25 }
func (a *App) getEnvelope(filePath string) ([]float64, error) { cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, "-v", "error", "-i", filePath, "-vn", "-sn", "-dn", "-ac", "1", "-ar", "8000", "-f", "s16le", "pipe:1"); cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}; var out bytes.Buffer; cmd.Stdout = &out; if err := cmd.Run(); err != nil { return nil, err }; audioData := out.Bytes(); chunkSize := 80; numSamples := len(audioData) / 2; envelope := make([]float64, 0, numSamples/chunkSize+1); var sum float64; var count int; for i := 0; i < len(audioData)-1; i += 2 { sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2])); sum += math.Abs(float64(sample)); count++; if count == chunkSize { envelope = append(envelope, sum/float64(chunkSize)); sum = 0; count = 0 } }; var totalSum float64; for _, val := range envelope { totalSum += val }; mean := totalSum / float64(len(envelope)); for i := range envelope { envelope[i] -= mean }; return envelope, nil }
func findDelay(envA, envV []float64) float64 { if len(envA) == 0 || len(envV) == 0 { return 0 }; step := 10; lenA_low := len(envA) / step; envA_low := make([]float64, lenA_low); for i := 0; i < lenA_low; i++ { sum := 0.0; for j := 0; j < step; j++ { sum += envA[i*step+j] }; envA_low[i] = sum / float64(step) }; lenV_low := len(envV) / step; envV_low := make([]float64, lenV_low); for i := 0; i < lenV_low; i++ { sum := 0.0; for j := 0; j < step; j++ { sum += envV[i*step+j] }; envV_low[i] = sum / float64(step) }; maxCorrLow := -1e10; bestDelayLow := 0; startK_low := -(lenV_low - 1); endK_low := lenA_low - 1; for k := startK_low; k <= endK_low; k++ { startI := 0; if k > 0 { startI = k }; endI := lenA_low; if lenV_low+k < lenA_low { endI = lenV_low + k }; var sum float64; for i := startI; i < endI; i++ { sum += envA_low[i] * envV_low[i-k] }; if sum > maxCorrLow { maxCorrLow = sum; bestDelayLow = k } }; approxDelay := bestDelayLow * step; window := 200; maxCorr := -1e10; bestDelay := approxDelay; startK := approxDelay - window; if startK < -(len(envV)-1) { startK = -(len(envV)-1) }; endK := approxDelay + window; if endK > len(envA)-1 { endK = len(envA)-1 }; for k := startK; k <= endK; k++ { startI := 0; if k > 0 { startI = k }; endI := len(envA); if len(envV)+k < len(envA) { endI = len(envV) + k }; var sum float64; for i := startI; i < endI; i++ { sum += envA[i] * envV[i-k] }; if sum > maxCorr { maxCorr = sum; bestDelay = k } }; return float64(bestDelay) / 100.0 }

func (a *App) SelectVideo() string { s, _ := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Выберите видео"}); return s }
func (a *App) SelectMultipleVideos() []string { s, _ := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{Title: "Выберите файлы"}); return s }
func (a *App) SelectAudio() string { s, _ := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Выберите аудио"}); return s }
func (a *App) SelectDirectory() string { s, _ := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Выберите папку"}); return s }

func (a *App) getVideoRes(filePath string) (int, int) {
	cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, "-i", filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()
	output := stderr.String()

	w, h := 1080, 1920
	reSize := regexp.MustCompile(`Video:.*?,.*?(\d{3,4})x(\d{3,4})`)
	mSize := reSize.FindStringSubmatch(output)
	if len(mSize) == 3 {
		w, _ = strconv.Atoi(mSize[1])
		h, _ = strconv.Atoi(mSize[2])
	}
	return w, h
}

// ==========================================
// ОСНОВНОЙ МОНТАЖ И МАГИЯ AI
// ==========================================
func (a *App) FastSync(videoPath, audioPath, outDir string) string {
	a.processCtx, a.cancelFn = context.WithCancel(a.ctx)
	defer a.cancelFn()

	t0 := time.Now()
	sendProgress := func(p int, m string) { runtime.EventsEmit(a.ctx, "fastsync_progress", map[string]interface{}{"percent": p, "message": m}) }
	os.MkdirAll(outDir, 0755)

	sendProgress(5, "Анализ аудиоволн (ищем задержку)...")
	vEnv, err := a.getEnvelope(videoPath)
	if err != nil { return "❌ Ошибка видео: " + err.Error() }
	aEnv, err := a.getEnvelope(audioPath)
	if err != nil { return "❌ Ошибка аудио: " + err.Error() }

	delay := findDelay(aEnv, vEnv)
	totalDur := a.getVideoDuration(videoPath)
	if totalDur <= 0 { totalDur = 1.0 }

	sendProgress(20, "Начинаем склейку (без пережатия видео)...")
	outPath := filepath.Join(outDir, fmt.Sprintf("FAST_SYNC_%s.mp4", time.Now().Format("15-04-05")))
	
	var cmdF []string
	if delay >= 0 {
		cmdF = []string{"-y", "-i", videoPath, "-ss", fmt.Sprintf("%.3f", delay), "-i", audioPath, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy", "-c:a", "aac", "-b:a", "192k", "-shortest", outPath}
	} else {
		cmdF = []string{"-y", "-ss", fmt.Sprintf("%.3f", math.Abs(delay)), "-i", videoPath, "-i", audioPath, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "copy", "-c:a", "aac", "-b:a", "192k", "-shortest", outPath}
	}

	cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, cmdF...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stderrPipe, _ := cmd.StderrPipe()
	
	if errStart := cmd.Start(); errStart != nil { return "❌ Ошибка запуска FFmpeg: " + errStart.Error() }

	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 { return 0, nil, nil }
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 { return i + 1, data[0:i], nil }
		if atEOF { return len(data), data, nil }
		return 0, nil, nil
	})

	timeRe := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d\.]+)`)
	for scanner.Scan() {
		if a.processCtx.Err() != nil { return "❌ Отменено" }
		line := scanner.Text()
		matches := timeRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			h, _ := strconv.ParseFloat(matches[1], 64); m, _ := strconv.ParseFloat(matches[2], 64); s, _ := strconv.ParseFloat(matches[3], 64)
			progress := (h*3600.0 + m*60.0 + s) / totalDur
			if progress > 1.0 { progress = 1.0 } else if progress < 0.0 { progress = 0.0 }
			sendProgress(20+int(progress*80), fmt.Sprintf("Копирование потоков: %d%%", int(progress*100)))
		}
	}
	
	if errRun := cmd.Wait(); errRun != nil { return "❌ Ошибка FFmpeg при склейке: " + errRun.Error() }

	sendProgress(100, "Готово!")
	return fmt.Sprintf("✅ Идеальный звук наложен!\n📁 Сохранено в: %s\n⏱ Заняло: %.1f сек", outDir, time.Since(t0).Seconds())
}

// ==========================================
// 1. ИДЕАЛЬНАЯ ВЕРТИКАЛЬНАЯ НАРЕЗКА (БЕЗ ЗАВИСАНИЙ)
// ==========================================
func (a *App) SplitWideCamera(widePath string, v1Path string, outDir string, useGPU bool) string {
	a.processCtx, a.cancelFn = context.WithCancel(a.ctx)
	defer a.cancelFn()
	os.MkdirAll(outDir, 0755)

	sendProgress := func(p int, m string) { runtime.EventsEmit(a.ctx, "sync_progress", map[string]interface{}{"percent": p, "message": m}) }

	sendProgress(0, "Чтение 'ДНК' Камеры 1...")
	
	cmdProbe := exec.CommandContext(a.processCtx, a.ffmpegPath, "-i", v1Path)
	cmdProbe.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var stderr bytes.Buffer
	cmdProbe.Stderr = &stderr
	cmdProbe.Run()
	outStr := stderr.String()

	physW, physH := 1080, 1920
	reSize := regexp.MustCompile(`Video:.*?,.*?(\d{3,4})x(\d{3,4})`)
	mSize := reSize.FindStringSubmatch(outStr)
	if len(mSize) == 3 {
		physW, _ = strconv.Atoi(mSize[1])
		physH, _ = strconv.Atoi(mSize[2])
	}

	vCodec := "libx264"
	if strings.Contains(outStr, "Video: hevc") || strings.Contains(outStr, "Video: h265") { 
		vCodec = "libx265" 
	}
	
	pixFmt := "yuv420p"
	if strings.Contains(outStr, "yuv420p10") { 
		pixFmt = "yuv420p10le" 
	} else if strings.Contains(outStr, "yuvj420p") { 
		pixFmt = "yuvj420p" 
	}

	rot := "0"
	if strings.Contains(outStr, "rotate") || strings.Contains(outStr, "displaymatrix: rotation") {
		if strings.Contains(outStr, "-90") || strings.Contains(outStr, "270") { rot = "270" } else if strings.Contains(outStr, "90") { rot = "90" }
	}

	if physW > physH && rot == "0" { rot = "90" }

	logicalW, logicalH := physW, physH
	if rot == "90" || rot == "270" { logicalW, logicalH = physH, physW }
	
	logicalW = (logicalW / 2) * 2
	logicalH = (logicalH / 2) * 2

	fps := a.getVideoFPS(v1Path)
	if fps <= 0 { fps = 25 }
	fpsStr := strconv.Itoa(fps)

	// 🔥 ФИКС ЗАВИСАНИЙ: Указываем частоту ключевых кадров (каждые 0.5 секунды)
	keyint := strconv.Itoa(fps / 2)
	if fps <= 0 { keyint = "15" }

	sendProgress(5, "Анализ Общего плана...")
	dur := a.getVideoDuration(widePath)
	if dur <= 0 { dur = 1.0 }

	outLeft := filepath.Join(outDir, "Cam3_Left.mp4")
	outRight := filepath.Join(outDir, "Cam4_Right.mp4")

	filterLeft := fmt.Sprintf("crop=iw/2:ih:0:0,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", logicalW, logicalH, logicalW, logicalH)
	filterRight := fmt.Sprintf("crop=iw/2:ih:iw/2:0,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", logicalW, logicalH, logicalW, logicalH)

	if rot == "90" {
		filterLeft += ",transpose=1"
		filterRight += ",transpose=1"
	} else if rot == "270" {
		filterLeft += ",transpose=2"
		filterRight += ",transpose=2"
	}
	filterLeft += ",setsar=1"
	filterRight += ",setsar=1"

	timeRe := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d\.]+)`)

	runCmdWithProgress := func(cmdF []string, startPercent, endPercent int, label string) error {
		cmd := exec.CommandContext(a.processCtx, a.ffmpegPath, cmdF...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		stderrPipe, _ := cmd.StderrPipe()
		cmd.Start()

		scanner := bufio.NewScanner(stderrPipe)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 { return 0, nil, nil }
			if i := bytes.IndexAny(data, "\r\n"); i >= 0 { return i + 1, data[0:i], nil }
			if atEOF { return len(data), data, nil }
			return 0, nil, nil
		})

		for scanner.Scan() {
			if a.processCtx.Err() != nil { return fmt.Errorf("Отменено пользователем") }
			line := scanner.Text()
			matches := timeRe.FindStringSubmatch(line)
			if len(matches) == 4 {
				h, _ := strconv.ParseFloat(matches[1], 64); m, _ := strconv.ParseFloat(matches[2], 64); s, _ := strconv.ParseFloat(matches[3], 64)
				t := h*3600.0 + m*60.0 + s
				prog := t / dur
				if prog > 1.0 { prog = 1.0 } else if prog < 0.0 { prog = 0.0 }
				currentP := startPercent + int(prog*float64(endPercent-startPercent))
				sendProgress(currentP, fmt.Sprintf("%s: %d%%", label, int(prog*100)))
			}
		}
		return cmd.Wait()
	}

	// 🔥 МАГИЯ: Добавлено "-g" и "-keyint_min", чтобы видео больше никогда не зависало!
	sendProgress(5, fmt.Sprintf("Рендер 1 из 2 (Клон: %s, %s)...", vCodec, pixFmt))
	cmdL := []string{"-y", "-i", widePath, "-map_metadata", "-1", "-vf", filterLeft, "-c:v", vCodec, "-preset", "ultrafast", "-crf", "22", "-threads", "0", "-pix_fmt", pixFmt, "-r", fpsStr, "-g", keyint, "-keyint_min", keyint}
	if rot != "0" {
		cmdL = append(cmdL, "-metadata:s:v:0", "rotate="+rot)
	} else {
		cmdL = append(cmdL, "-metadata:s:v:0", "rotate=0")
	}
	cmdL = append(cmdL, "-c:a", "aac", "-b:a", "192k", "-ac", "2", "-ar", "48000", outLeft)

	sendProgress(50, fmt.Sprintf("Рендер 2 из 2 (Клон: %s, %s)...", vCodec, pixFmt))
	cmdR := []string{"-y", "-i", widePath, "-map_metadata", "-1", "-vf", filterRight, "-c:v", vCodec, "-preset", "ultrafast", "-crf", "22", "-threads", "0", "-pix_fmt", pixFmt, "-r", fpsStr, "-g", keyint, "-keyint_min", keyint}
	if rot != "0" {
		cmdR = append(cmdR, "-metadata:s:v:0", "rotate="+rot)
	} else {
		cmdR = append(cmdR, "-metadata:s:v:0", "rotate=0")
	}
	cmdR = append(cmdR, "-c:a", "aac", "-b:a", "192k", "-ac", "2", "-ar", "48000", outRight)

	if err := runCmdWithProgress(cmdL, 5, 50, "Рендер Левой части"); err != nil { return "❌ Ошибка рендера: " + err.Error() }
	if err := runCmdWithProgress(cmdR, 50, 100, "Рендер Правой части"); err != nil { return "❌ Ошибка рендера: " + err.Error() }

	sendProgress(100, "Успешно завершено!")
	return "✅ Общий план разрезан! Созданы 100% технические клоны Камеры 1 (с частыми ключами для идеальной склейки)."
}

// ==========================================
// 2. ИДЕАЛЬНАЯ СБОРКА МУЛЬТИКАМА
// ==========================================
func (a *App) RunSync(v1Path, aPath, v2Path, v3LeftPath, v4RightPath, apiKey string, mainCam int, crf int, outDir string, testDuration int) string {
	a.processCtx, a.cancelFn = context.WithCancel(a.ctx)
	defer a.cancelFn()

	t0 := time.Now()
	sendProgress := func(p int, m string) { runtime.EventsEmit(a.ctx, "sync_progress", map[string]interface{}{"percent": p, "message": m}) }
	os.MkdirAll(outDir, 0755)

	sendProgress(5, "Анализ файлов...")
	v1Env, err := a.getEnvelope(v1Path)
	if err != nil { return "❌ Ошибка видео 1: " + err.Error() }
	aEnv, err := a.getEnvelope(aPath)
	if err != nil { return "❌ Ошибка аудио: " + err.Error() }
	masterDur := float64(len(aEnv)) / 100.0

	cmdProbe := exec.CommandContext(a.processCtx, a.ffmpegPath, "-i", v1Path)
	cmdProbe.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var stderr bytes.Buffer
	cmdProbe.Stderr = &stderr
	cmdProbe.Run()
	outStr := stderr.String()
	
	rot := "0"
	if strings.Contains(outStr, "rotate") || strings.Contains(outStr, "displaymatrix: rotation") {
		if strings.Contains(outStr, "-90") || strings.Contains(outStr, "270") { rot = "270" } else if strings.Contains(outStr, "90") { rot = "90" }
	}

	delay1 := findDelay(aEnv, v1Env)
	var delay2, delay3 float64
	if v2Path != "" { v2Env, _ := a.getEnvelope(v2Path); delay2 = findDelay(aEnv, v2Env) }
	if v3LeftPath != "" { v3Env, _ := a.getEnvelope(v3LeftPath); delay3 = findDelay(aEnv, v3Env) }

	masterStart := delay1
	if v2Path != "" && delay2 < masterStart { masterStart = delay2 }
	if v3LeftPath != "" && delay3 < masterStart { masterStart = delay3 }
	if masterStart < 0 { masterStart = 0 }

	if testDuration > 0 && masterDur > (masterStart+float64(testDuration)) { masterDur = masterStart + float64(testDuration) }
	trimDur := masterDur - masterStart

	sendProgress(10, "Подготовка WAV для ИИ...")
	tempWav := filepath.Join(os.TempDir(), fmt.Sprintf("ai_master_%d.wav", time.Now().UnixNano()))
	exec.CommandContext(a.processCtx, a.ffmpegPath, "-y", "-ss", fmt.Sprintf("%.3f", masterStart), "-t", fmt.Sprintf("%.3f", trimDur), "-i", aPath, "-vn", "-ar", "16000", "-ac", "1", tempWav).Run()
	defer os.Remove(tempWav) 
	
	fData, _ := os.ReadFile(tempWav)
	req, _ := http.NewRequestWithContext(a.processCtx, "POST", "https://api.assemblyai.com/v2/upload", bytes.NewReader(fData))
	req.Header.Set("Authorization", apiKey)
	resp, _ := (&http.Client{}).Do(req)
	var upRes AssemblyUploadRes
	json.NewDecoder(resp.Body).Decode(&upRes)
	resp.Body.Close()

	sendProgress(25, "AssemblyAI переводит голос в текст...")
	reqBody := fmt.Sprintf(`{"audio_url":"%s","speaker_labels":true,"language_code":"ru","speech_models":["universal-2"]}`, upRes.UploadURL)
	req2, _ := http.NewRequestWithContext(a.processCtx, "POST", "https://api.assemblyai.com/v2/transcript", strings.NewReader(reqBody))
	req2.Header.Set("Authorization", apiKey)
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := (&http.Client{}).Do(req2)
	var trRes AssemblyTranscriptRes
	json.NewDecoder(resp2.Body).Decode(&trRes)
	resp2.Body.Close()

	waitSec := 0
	var cloudUtterances []AssemblyUtterance
	for {
		select { case <-a.processCtx.Done(): return "❌ Отменено"; case <-time.After(3 * time.Second): }
		waitSec += 3
		sendProgress(25, fmt.Sprintf("AssemblyAI думает... (прошло %d сек)", waitSec))

		r, _ := http.NewRequestWithContext(a.processCtx, "GET", "https://api.assemblyai.com/v2/transcript/"+trRes.ID, nil)
		r.Header.Set("Authorization", apiKey)
		re, err := (&http.Client{}).Do(r)
		if err != nil { continue }
		var pollRes AssemblyPollRes
		json.NewDecoder(re.Body).Decode(&pollRes)
		re.Body.Close()
		
		if pollRes.Status == "completed" { cloudUtterances = pollRes.Utterances; break }
	}

	sendProgress(40, "Синхронизация камер и умный фильтр...")
	spMap := make(map[string]int)
	next := 1
	for _, u := range cloudUtterances {
		if _, ok := spMap[u.Speaker]; !ok { spMap[u.Speaker] = next; next++; if next > 2 { next = 2 } }
	}

	type interval struct { s, e float64; cam int }
	var mainIntervals, guestIntervals []interval
	for _, u := range cloudUtterances {
		s, e := float64(u.Start)/1000.0, float64(u.End)/1000.0
		if spMap[u.Speaker] == mainCam { mainIntervals = append(mainIntervals, interval{s - 0.1, e + 0.3, mainCam}) } else { guestIntervals = append(guestIntervals, interval{s - 0.1, e + 0.1, spMap[u.Speaker]}) }
	}

	merge := func(arr []interval, gap float64) []interval {
		if len(arr) == 0 { return nil }
		sort.Slice(arr, func(i, j int) bool { return arr[i].s < arr[j].s })
		res := []interval{arr[0]}
		for i := 1; i < len(arr); i++ {
			l := &res[len(res)-1]
			if arr[i].s <= l.e+gap { if arr[i].e > l.e { l.e = arr[i].e } } else { res = append(res, arr[i]) }
		}
		return res
	}
	mainIntervals = merge(mainIntervals, 0.3)
	guestIntervals = merge(guestIntervals, 0.1)

	var evs []float64
	for _, v := range mainIntervals { evs = append(evs, v.s, v.e) }
	for _, v := range guestIntervals { evs = append(evs, v.s, v.e) }
	evs = append(evs, masterStart, masterDur)
	if delay1 > masterStart { evs = append(evs, delay1) }
	if v2Path != "" && delay2 > masterStart { evs = append(evs, delay2) }
	if v3LeftPath != "" && delay3 > masterStart { evs = append(evs, delay3) }
	sort.Float64s(evs)

	type Chunk struct { Cam int; StartA float64; EndA float64 }
	var baseCuts []Chunk
	for i := 0; i < len(evs)-1; i++ {
		t1, t2 := evs[i], evs[i+1]
		if t1 < masterStart { t1 = masterStart }
		if t2 > masterDur { t2 = masterDur } 
		if t1 >= t2 { continue } 
		
		mid := (t1 + t2) / 2.0
		mainAct, guestAct := false, false
		for _, v := range mainIntervals { if mid >= v.s && mid <= v.e { mainAct = true; break } }
		for _, v := range guestIntervals { if mid >= v.s && mid <= v.e { guestAct = true; break } }

		canUse1 := mid >= delay1
		canUse2 := v2Path != "" && mid >= delay2

		cam := 1
		if v2Path == "" {
			cam = 1
		} else {
			if mainCam == 1 { if mainAct && canUse1 { cam = 1 } else if guestAct && canUse2 { cam = 2 } else if canUse1 { cam = 1 } else if canUse2 { cam = 2 } } else { if mainAct && canUse2 { cam = 2 } else if guestAct && canUse1 { cam = 1 } else if canUse2 { cam = 2 } else if canUse1 { cam = 1 } }
		}

		if len(baseCuts) > 0 && baseCuts[len(baseCuts)-1].Cam == cam { baseCuts[len(baseCuts)-1].EndA = t2 } else { if t2-t1 < 0.2 && len(baseCuts) > 0 { baseCuts[len(baseCuts)-1].EndA = t2 } else { baseCuts = append(baseCuts, Chunk{cam, t1, t2}) } }
	}

	var finalCuts []Chunk
	for i := 0; i < len(baseCuts); i++ {
		c := baseCuts[i]
		dur := c.EndA - c.StartA

		hasMidShots := v3LeftPath != "" && v4RightPath != ""

		if hasMidShots && dur > 5.0 && i < len(baseCuts)-1 && baseCuts[i+1].Cam != c.Cam {
			if time.Now().UnixNano()%100 < 15 {
				cutPoint := c.EndA - 3.0
				finalCuts = append(finalCuts, Chunk{c.Cam, c.StartA, cutPoint})
				camMid := 31 
				if c.Cam == 2 { camMid = 32 } 
				finalCuts = append(finalCuts, Chunk{camMid, cutPoint, c.EndA})
				continue
			}
		}
		finalCuts = append(finalCuts, c)
	}

	sendProgress(50, "Рендер основного видео (Идеальная база)...")
	
	fpsInt := a.getVideoFPS(v1Path)
	if fpsInt <= 0 { fpsInt = 25 }
	fpsF := float64(fpsInt)
	frameDur := 1.0 / fpsF

	var sb strings.Builder
	for _, c := range finalCuts {
		d, p := delay1, v1Path
		if c.Cam == 2 { d, p = delay2, v2Path }
		if c.Cam == 31 { d, p = delay3, v3LeftPath } 
		if c.Cam == 32 { d, p = delay3, v4RightPath } 
		
		inP, outP := c.StartA-d, c.EndA-d
		if inP < 0 { inP = 0 }
		
		inP = math.Round(inP/frameDur) * frameDur
		outP = math.Round(outP/frameDur) * frameDur
		if outP <= inP { outP = inP + frameDur }

		sb.WriteString(fmt.Sprintf("file '%s'\ninpoint %.3f\noutpoint %.3f\n", strings.ReplaceAll(p, "\\", "/"), inP, outP))
	}
	lst := filepath.Join(os.TempDir(), "lst.txt")
	os.WriteFile(lst, []byte(sb.String()), 0644)

	mainOutPath := filepath.Join(outDir, fmt.Sprintf("RESULT_%s.mp4", time.Now().Format("15-04-05")))
	
	cmdF := []string{"-y", "-f", "concat", "-safe", "0", "-i", lst, "-ss", fmt.Sprintf("%.3f", masterStart), "-i", aPath, "-map", "0:v:0", "-map", "1:a:0", "-shortest", "-c:v", "libx264", "-preset", "superfast", "-crf", strconv.Itoa(crf), "-r", strconv.Itoa(fpsInt), "-vsync", "1", "-c:a", "aac", "-b:a", "192k"}
	if rot != "0" {
		cmdF = append(cmdF, "-metadata:s:v:0", "rotate="+rot)
	} else {
		cmdF = append(cmdF, "-metadata:s:v:0", "rotate=0")
	}
	if testDuration > 0 { cmdF = append(cmdF, "-t", strconv.Itoa(testDuration)) }
	cmdF = append(cmdF, mainOutPath)
	
	renderCmd := exec.CommandContext(a.processCtx, a.ffmpegPath, cmdF...)
	renderCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stderrPipe, _ := renderCmd.StderrPipe()
	renderCmd.Start()

	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 { return 0, nil, nil }
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 { return i + 1, data[0:i], nil }
		if atEOF { return len(data), data, nil }
		return 0, nil, nil
	})

	timeRe := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d\.]+)`)
	for scanner.Scan() {
		if a.processCtx.Err() != nil { return "❌ Отменено" }
		line := scanner.Text()
		matches := timeRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			h, _ := strconv.ParseFloat(matches[1], 64); m, _ := strconv.ParseFloat(matches[2], 64); s, _ := strconv.ParseFloat(matches[3], 64)
			progress := (h*3600.0 + m*60.0 + s) / trimDur
			if progress > 1.0 { progress = 1.0 } else if progress < 0.0 { progress = 0.0 }
			sendProgress(50+int(progress*50), fmt.Sprintf("Рендер основы: %d%%", int(progress*100)))
		}
	}
	renderCmd.Wait()

	sendProgress(100, "Успешно завершено!")
	if testDuration > 0 { return fmt.Sprintf("✅ ТЕСТ ГОТОВ (%d сек)\n📁 Сохранено в: %s\n⏱ Заняло: %.1f сек", testDuration, outDir, time.Since(t0).Seconds()) }
	return fmt.Sprintf("✅ Склейка завершена (Склеек: %d)\n📁 Сохранено в: %s\n⏱ Заняло: %.1f сек", len(finalCuts), outDir, time.Since(t0).Seconds())
}

// ==========================================
// СЖАТИЕ, СКЛЕЙКА И СЕРВЕР
// ==========================================

func (a *App) MergeVideos(v1 string, v2 string, v3 string, outDir string) string {
	a.processCtx, a.cancelFn = context.WithCancel(a.ctx)
	defer a.cancelFn()

	t0 := time.Now()
	sendProgress := func(p int, m string) { runtime.EventsEmit(a.ctx, "merge_progress", map[string]interface{}{"percent": p, "message": m}) }

	os.MkdirAll(outDir, 0755)

	sendProgress(5, "Чтение длительности видео...")
	dur1 := a.getVideoDuration(v1)
	dur2 := a.getVideoDuration(v2)
	dur3 := 0.0
	if v3 != "" { dur3 = a.getVideoDuration(v3) }
	totalDur := dur1 + dur2 + dur3

	sendProgress(10, "Подготовка к склейке...")
	lst := filepath.Join(os.TempDir(), "merge_list.txt")
	
	sb := fmt.Sprintf("file '%s'\nfile '%s'\n", strings.ReplaceAll(v1, "\\", "/"), strings.ReplaceAll(v2, "\\", "/"))
	if v3 != "" { sb += fmt.Sprintf("file '%s'\n", strings.ReplaceAll(v3, "\\", "/")) }
	os.WriteFile(lst, []byte(sb), 0644)

	outPath := filepath.Join(outDir, fmt.Sprintf("MERGED_%s.mp4", time.Now().Format("15-04-05")))
	cmdF := []string{"-y", "-fflags", "+genpts", "-f", "concat", "-safe", "0", "-i", lst, "-c:v", "copy", "-c:a", "aac", "-async", "1", outPath}

	sendProgress(20, "Запуск процесса объединения...")
	renderCmd := exec.CommandContext(a.processCtx, a.ffmpegPath, cmdF...)
	renderCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stderrPipe, _ := renderCmd.StderrPipe()
	renderCmd.Start()

	scanner := bufio.NewScanner(stderrPipe)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 { return 0, nil, nil }
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 { return i + 1, data[0:i], nil }
		if atEOF { return len(data), data, nil }
		return 0, nil, nil
	})

	timeRe := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d\.]+)`)
	for scanner.Scan() {
		if a.processCtx.Err() != nil { return "❌ Процесс отменен пользователем!" }
		line := scanner.Text()
		matches := timeRe.FindStringSubmatch(line)
		if len(matches) == 4 {
			h, _ := strconv.ParseFloat(matches[1], 64); m, _ := strconv.ParseFloat(matches[2], 64); s, _ := strconv.ParseFloat(matches[3], 64)
			progress := (h*3600.0 + m*60.0 + s) / totalDur
			if progress > 1.0 { progress = 1.0 } else if progress < 0.0 { progress = 0.0 }
			sendProgress(20+int(progress*80), fmt.Sprintf("Склейка финала: %d%%", int(progress*100)))
		}
	}
	renderCmd.Wait()
	
	if a.processCtx.Err() != nil { return "❌ Процесс отменен!" }
	sendProgress(100, "Видео успешно склеены!")
	return fmt.Sprintf("✅ Готово!\n📁 Сохранено в: %s\n⏱ Заняло: %.1f сек", outDir, time.Since(t0).Seconds())
}

func (a *App) CompressBatch(videoPaths []string, crf int, resolution string, crop bool, audioOnly bool, useGPU bool, useServer bool, serverIP string, serverSecret string, outDir string) string {
	a.processCtx, a.cancelFn = context.WithCancel(a.ctx)
	defer a.cancelFn()

	t0 := time.Now()
	sendProgress := func(p int, m string) { runtime.EventsEmit(a.ctx, "compress_progress", map[string]interface{}{"percent": p, "message": m}) }

	os.MkdirAll(outDir, 0755)
	totalFiles := len(videoPaths)
	if totalFiles == 0 { return "❌ Файлы не выбраны" }

	for i, videoPath := range videoPaths {
		if a.processCtx.Err() != nil { return "❌ Процесс отменен!" }

		sendProgress(5, fmt.Sprintf("Анализ файла %d из %d...", i+1, totalFiles))
		totalDur := a.getVideoDuration(videoPath)
		fps := a.getVideoFPS(videoPath)
		if fps <= 0 { fps = 25 }

		baseName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
		var outPath string
		var cmdF []string
		
		if useGPU {
			cmdF = []string{"-y", "-hwaccel", "auto", "-i", videoPath}
		} else {
			cmdF = []string{"-y", "-i", videoPath}
		}

		if audioOnly {
			outPath = filepath.Join(outDir, baseName+"_AUDIO.mp3")
			cmdF = append(cmdF, "-vn", "-c:a", "libmp3lame", "-q:a", "2", outPath)
		} else {
			outPath = filepath.Join(outDir, baseName+"_COMPRESSED.mp4")

			if useGPU {
				cmdF = append(cmdF, "-c:v", "h264_nvenc", "-preset", "hq", "-rc", "vbr", "-cq", strconv.Itoa(crf), "-b:v", "0")
			} else {
				cmdF = append(cmdF, "-c:v", "libx264", "-crf", strconv.Itoa(crf), "-preset", "superfast")
			}

			var filters []string
			if crop { filters = append(filters, "crop=ih*(9/16):ih") }
			if resolution == "1080" {
				filters = append(filters, "scale='if(gt(iw,ih),1920,-2)':'if(gt(iw,ih),-2,1920)'")
			} else if resolution == "720" {
				filters = append(filters, "scale='if(gt(iw,ih),1280,-2)':'if(gt(iw,ih),-2,1280)'")
			}
			
			filters = append(filters, "pad='ceil(iw/2)*2':'ceil(ih/2)*2'", "format=yuv420p")
			
			if len(filters) > 0 { cmdF = append(cmdF, "-vf", strings.Join(filters, ",")) }
			
			cmdF = append(cmdF, "-c:a", "aac", "-b:a", "192k", outPath)
		}

		var exePath string
		if useServer {
			exePath = a.ffmpegClientPath 
			cleanIP := strings.ReplaceAll(serverIP, "ws://", "")
			cleanIP = strings.ReplaceAll(cleanIP, "wss://", "")
			cleanIP = strings.ReplaceAll(cleanIP, "http://", "")
			cleanIP = strings.ReplaceAll(cleanIP, "https://", "")
			
			configJSON := fmt.Sprintf(`{"address": "%s", "authSecret": "%s"}`, cleanIP, serverSecret)
			configPath1 := filepath.Join(os.TempDir(), "ffmpeg-over-ip.client.jsonc")
			configPath2 := filepath.Join(os.TempDir(), ".ffmpeg-over-ip.client.jsonc")
			os.WriteFile(configPath1, []byte(configJSON), 0644)
			os.WriteFile(configPath2, []byte(configJSON), 0644)
		} else {
			exePath = a.ffmpegPath
		}

		cmdStr := strings.Join(cmdF, " ")
		if useServer {
			sendProgress(10, fmt.Sprintf("☁️ [СЕРВЕР] ffmpeg %s", cmdStr))
		} else {
			sendProgress(10, fmt.Sprintf("💻 [ЛОКАЛЬНО] ffmpeg %s", cmdStr))
		}
		
		renderCmd := exec.CommandContext(a.processCtx, exePath, cmdF...)
		renderCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		
		stdoutPipe, _ := renderCmd.StdoutPipe()
		stderrPipe, _ := renderCmd.StderrPipe()
		
		var fullLog strings.Builder
		var logMutex sync.Mutex 
		timeRe := regexp.MustCompile(`time=(\d{2}):(\d{2}):([\d\.]+)`)

		processLine := func(line string) {
			if a.processCtx.Err() != nil { return }
			
			logMutex.Lock()
			fullLog.WriteString(line + "\n")
			logMutex.Unlock()

			matches := timeRe.FindStringSubmatch(line)
			if len(matches) == 4 {
				h, _ := strconv.ParseFloat(matches[1], 64)
				m, _ := strconv.ParseFloat(matches[2], 64)
				s, _ := strconv.ParseFloat(matches[3], 64)
				progress := (h*3600.0 + m*60.0 + s) / totalDur
				if progress > 1.0 { progress = 1.0 } else if progress < 0.0 { progress = 0.0 }
				
				overallProgress := int(((float64(i) + progress) / float64(totalFiles)) * 100)
				sendProgress(overallProgress, fmt.Sprintf("Файл %d из %d (%d%%)", i+1, totalFiles, int(progress*100)))
			}
		}

		splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 { return 0, nil, nil }
			if j := bytes.IndexAny(data, "\r\n"); j >= 0 { return j + 1, data[0:j], nil }
			if atEOF { return len(data), data, nil }
			return 0, nil, nil
		}

		renderCmd.Start()

		go func() {
			outScanner := bufio.NewScanner(stdoutPipe)
			outScanner.Split(splitFunc)
			for outScanner.Scan() {
				processLine(outScanner.Text())
			}
		}()

		errScanner := bufio.NewScanner(stderrPipe)
		errScanner.Split(splitFunc)
		for errScanner.Scan() {
			processLine(errScanner.Text())
		}
		
		err := renderCmd.Wait()
		if err != nil && a.processCtx.Err() == nil {
			logMutex.Lock()
			logStr := fullLog.String()
			logMutex.Unlock()
			
			if len(logStr) > 500 { logStr = "..." + logStr[len(logStr)-500:] }
			if logStr == "" { logStr = "Нет ответа от программы-клиента." }

			return fmt.Sprintf("❌ Ошибка при сжатии %s:\n%s", filepath.Base(videoPath), logStr)
		}
	}

	sendProgress(100, "Пакетное сжатие завершено!")
	return fmt.Sprintf("✅ Готово!\n📁 Обработано файлов: %d\n⏱ Заняло: %.1f сек", len(videoPaths), time.Since(t0).Seconds())
}

func (a *App) StartServer(port string, secret string) string {
	a.StopServer()
	configJSON := fmt.Sprintf(`{"address": "0.0.0.0:%s", "authSecret": "%s"}`, port, secret)
	configPath := filepath.Join(os.TempDir(), "ffmpeg-over-ip.server.jsonc")
	os.WriteFile(configPath, []byte(configJSON), 0644)
	a.serverCmd = exec.Command(a.ffmpegServerPath, "-config", configPath)
	a.serverCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdoutPipe, _ := a.serverCmd.StdoutPipe()
	stderrPipe, _ := a.serverCmd.StderrPipe()
	if err := a.serverCmd.Start(); err != nil { return "❌ Ошибка запуска ноды: " + err.Error() }
	go func() { scanner := bufio.NewScanner(stdoutPipe); for scanner.Scan() { runtime.EventsEmit(a.ctx, "server_log", map[string]interface{}{"message": scanner.Text()}) } }()
	go func() { scanner := bufio.NewScanner(stderrPipe); for scanner.Scan() { runtime.EventsEmit(a.ctx, "server_log", map[string]interface{}{"message": scanner.Text()}) } }()
	return "✅ Сервер запущен (Порт: " + port + ")"
}

func (a *App) StopServer() {
	if a.serverCmd != nil && a.serverCmd.Process != nil { a.serverCmd.Process.Kill() }
	exec.Command("taskkill", "/F", "/IM", "ffmpeg-over-ip-server.exe", "/T").Run()
}

func (a *App) StartTgBot(token string, aiKey string) string {
	a.StopTgBot()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil { return "❌ Ошибка: Неверный Токен бота!" }

	a.botCtx, a.botCancel = context.WithCancel(context.Background())
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	go func() {
		for {
			select {
			case <-a.botCtx.Done():
				bot.StopReceivingUpdates()
				return
			case update := <-updates:
				if update.Message != nil {
					if update.Message.Video != nil || update.Message.Document != nil {
						go a.askVideoFormat(bot, update.Message)
					} else if update.Message.Voice != nil {
						go a.handleTGVoice(bot, update.Message, aiKey)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "👋 Привет!\n\n🎥 Отправь видео -> Сделаю Кружок или Стикер.\n🎙️ Отправь голосовое -> Переведу в текст.")
						bot.Send(msg)
					}
				} else if update.CallbackQuery != nil {
					bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "")) 
					go a.handleCallback(bot, update.CallbackQuery)
				}
			}
		}
	}()

	return "✅ Бот [" + bot.Self.UserName + "] запущен и ждет файлы!"
}

func (a *App) StopTgBot() {
	if a.botCancel != nil { a.botCancel() }
}

func (a *App) handleTGVoice(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, aiKey string) {
	chatID := msg.Chat.ID
	
	if aiKey == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка: Вставьте ключ AssemblyAI."))
		return
	}

	statusMsg, _ := bot.Send(tgbotapi.NewMessage(chatID, "🎧 1/4 Получаю ссылку на аудио..."))
	runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": "🎧 Расшифровка голосового от " + msg.From.UserName})

	url, err := bot.GetFileDirectURL(msg.Voice.FileID)
	if err != nil {
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Ошибка (Телеграм не отдал файл): " + err.Error()))
		return 
	}

	bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "🎧 2/4 Скачиваю голосовое на сервер..."))
	resp, err := http.Get(url)
	if err != nil { 
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Ошибка скачивания аудио: " + err.Error()))
		return 
	}
	defer resp.Body.Close()
	voiceData, _ := io.ReadAll(resp.Body)

	bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "🎧 3/4 Загружаю аудио в нейросеть..."))

	req, _ := http.NewRequestWithContext(a.botCtx, "POST", "https://api.assemblyai.com/v2/upload", bytes.NewReader(voiceData))
	req.Header.Set("Authorization", aiKey)
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Ошибка связи с AssemblyAI: " + err.Error()))
		return
	}
	
	if res.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		res.Body.Close()
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("❌ ИИ отклонил загрузку. Код: %d\nОтвет: %s", res.StatusCode, string(bodyBytes))))
		return
	}

	var upRes AssemblyUploadRes
	json.NewDecoder(res.Body).Decode(&upRes)
	res.Body.Close()

	if upRes.UploadURL == "" {
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ ИИ не вернул ссылку на загруженный файл."))
		return
	}

	bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "🎧 4/4 ИИ расшифровывает (обычно 5-10 сек)..."))

	reqBody := fmt.Sprintf(`{"audio_url":"%s","language_code":"ru","speech_models":["universal-2"]}`, upRes.UploadURL)
	req2, _ := http.NewRequestWithContext(a.botCtx, "POST", "https://api.assemblyai.com/v2/transcript", strings.NewReader(reqBody))
	req2.Header.Set("Authorization", aiKey)
	req2.Header.Set("Content-Type", "application/json")
	res2, err2 := (&http.Client{}).Do(req2)
	
	if err2 != nil {
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Ошибка команды расшифровки: " + err2.Error()))
		return
	}
	
	if res2.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(res2.Body)
		res2.Body.Close()
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, fmt.Sprintf("❌ ИИ отклонил команду. Код: %d\nОтвет: %s", res2.StatusCode, string(bodyBytes))))
		return
	}

	var trRes AssemblyTranscriptRes
	json.NewDecoder(res2.Body).Decode(&trRes)
	res2.Body.Close()

	if trRes.ID == "" {
		bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ ИИ не создал задачу (нет ID)."))
		return
	}

	attempts := 0
	for {
		attempts++
		if attempts > 30 {
			bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Таймаут. ИИ завис и думает дольше 60 секунд."))
			return
		}

		select {
		case <-a.botCtx.Done():
			return
		case <-time.After(2 * time.Second):
		}
		
		r, _ := http.NewRequestWithContext(a.botCtx, "GET", "https://api.assemblyai.com/v2/transcript/"+trRes.ID, nil)
		r.Header.Set("Authorization", aiKey)
		re, err := (&http.Client{}).Do(r)
		if err != nil { continue }
		
		var pollRes struct {
			Status string `json:"status"`
			Error  string `json:"error"`
			Text   string `json:"text"`
		}
		json.NewDecoder(re.Body).Decode(&pollRes)
		re.Body.Close()

		if pollRes.Status == "completed" {
			finalText := fmt.Sprintf("📝 **Текст:**\n\n%s", pollRes.Text)
			bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, finalText))
			runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": "✅ Голосовое успешно расшифровано!"})
			break
		} else if pollRes.Status == "error" {
			bot.Send(tgbotapi.NewEditMessageText(chatID, statusMsg.MessageID, "❌ Внутренняя ошибка ИИ: "+pollRes.Error))
			break
		}
	}
}

func (a *App) askVideoFormat(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	var fileID string
	var fileSize int

	if msg.Video != nil {
		fileID = msg.Video.FileID
		fileSize = msg.Video.FileSize
	} else if msg.Document != nil {
		fileID = msg.Document.FileID
		fileSize = msg.Document.FileSize
	}

	if fileSize > 20971520 {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка: Файл весит больше 20 МБ."))
		return
	}

	a.pendingFilesMu.Lock()
	a.pendingFiles[chatID] = fileID
	a.pendingFilesMu.Unlock()

	reply := tgbotapi.NewMessage(chatID, "🎥 Видео получено! Что из него сделать?")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 Кружочек", "circle"),
			tgbotapi.NewInlineKeyboardButtonData("👾 Стикер", "sticker"),
		),
	)
	bot.Send(reply)
}

func (a *App) handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	action := query.Data 

	a.pendingFilesMu.Lock()
	fileID, exists := a.pendingFiles[chatID]
	delete(a.pendingFiles, chatID) 
	a.pendingFilesMu.Unlock()

	if !exists {
		bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "❌ Видео устарело. Отправьте заново."))
		return
	}

	bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "⏳ Скачиваю видео..."))

	url, err := bot.GetFileDirectURL(fileID)
	if err != nil { return }

	inPath := filepath.Join(os.TempDir(), fmt.Sprintf("tg_in_%d.mp4", time.Now().UnixNano()))
	ext := ".mp4"
	if action == "sticker" { ext = ".webm" }
	outPath := filepath.Join(os.TempDir(), fmt.Sprintf("tg_out_%d%s", time.Now().UnixNano(), ext))
	
	defer os.Remove(inPath)
	defer os.Remove(outPath)

	resp, err := http.Get(url)
	if err != nil { return }
	defer resp.Body.Close()
	outFile, _ := os.Create(inPath)
	io.Copy(outFile, resp.Body)
	outFile.Close()

	userName := query.From.UserName

	if action == "sticker" {
		bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "👾 Рендерю Telegram-Стикер (VP9, 512x512, без звука)..."))
		runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": fmt.Sprintf("👾 Рендер стикера от %s...", userName)})
		
		cmd := exec.Command(a.ffmpegPath, "-y", "-i", inPath, "-vf", "crop='min(iw,ih)':'min(iw,ih)',scale=512:512", "-c:v", "libvpx-vp9", "-b:v", "256k", "-an", "-t", "2.9", "-r", "30", outPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Run()

		bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "🚀 Отправляю стикер..."))
		sticker := tgbotapi.NewSticker(chatID, tgbotapi.FilePath(outPath))
		bot.Send(sticker)
		runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": "✅ Стикер успешно отправлен!"})

	} else {
		bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "🔄 Рендерю кружочек..."))
		runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": fmt.Sprintf("🔄 Рендер кружочка от %s...", userName)})

		cmd := exec.Command(a.ffmpegPath, "-y", "-i", inPath, "-vf", "crop='min(iw,ih)':'min(iw,ih)',scale=384:384,format=yuv420p", "-c:v", "libx264", "-preset", "superfast", "-crf", "24", "-c:a", "aac", "-b:a", "128k", "-t", "59", outPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Run()

		bot.Send(tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, "🚀 Отправляю кружок обратно..."))
		vNote := tgbotapi.NewVideoNote(chatID, 384, tgbotapi.FilePath(outPath))
		bot.Send(vNote)
		runtime.EventsEmit(a.ctx, "tg_log", map[string]interface{}{"message": "✅ Кружок успешно отправлен!"})
	}
	
	bot.Send(tgbotapi.NewDeleteMessage(chatID, query.Message.MessageID))
}