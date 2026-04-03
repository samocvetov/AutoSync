//go:build windows

package renderserverapp

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func applyWindowsCommandAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func killCommandTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	taskkill := exec.Command("taskkill", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F")
	applyWindowsCommandAttrs(taskkill)
	output, err := taskkill.CombinedOutput()
	if err == nil {
		return nil
	}
	text := strings.ToLower(string(output))
	if strings.Contains(text, "not found") || strings.Contains(text, "there is no running instance") {
		return nil
	}
	return err
}
