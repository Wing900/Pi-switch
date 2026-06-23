//go:build windows

package pi

import (
	"os/exec"
	"syscall"
)

func openCommandInTerminal(command string, workingDir string) error {
	cmd := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000200, // DETACHED_PROCESS
	}
	return cmd.Start()
}
