//go:build windows

package studioapp

import (
	"os/exec"
	"syscall"
	"unsafe"
)

func applyWindowsCommandAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetShortPathNameW = kernel32.NewProc("GetShortPathNameW")
)

func getWindowsShortPath(path string) string {
	src, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return ""
	}

	size, _, _ := procGetShortPathNameW.Call(
		uintptr(unsafe.Pointer(src)),
		0,
		0,
	)
	if size == 0 {
		return ""
	}

	buf := make([]uint16, size)
	result, _, _ := procGetShortPathNameW.Call(
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if result == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:result])
}
