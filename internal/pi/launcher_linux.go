//go:build linux

package pi

import (
	"os/exec"
)

func openCommandInTerminal(command string, workingDir string) error {
	cmd := exec.Command("x-terminal-emulator", "-e", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	return cmd.Start()
}
