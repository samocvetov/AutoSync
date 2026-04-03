//go:build !windows

package renderserverapp

import "os/exec"

func applyWindowsCommandAttrs(cmd *exec.Cmd) {}

func killCommandTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
