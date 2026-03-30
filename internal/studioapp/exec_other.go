//go:build !windows

package studioapp

import "os/exec"

func applyWindowsCommandAttrs(cmd *exec.Cmd) {}

func getWindowsShortPath(path string) string { return "" }
