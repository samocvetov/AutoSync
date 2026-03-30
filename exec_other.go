//go:build !windows

package main

import "os/exec"

func applyWindowsCommandAttrs(cmd *exec.Cmd) {}

func getWindowsShortPath(path string) string { return "" }
