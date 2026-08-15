//go:build windows

package wsl

import (
	"os/exec"
	"syscall"
)

const piDetectionCreateNoWindow uint32 = 0x08000000

var piDetectionCommand = exec.Command

func DetectPi(distro string) (PiDetection, error) {
	return detectPi(distro, Detect, probePi)
}

func probePi(distro string) ([]byte, error) {
	cmd := piDetectionCommand("wsl.exe", "--distribution", distro, "--exec", "sh", "-c", piDetectionScript)
	// Prevent a console window for this noninteractive WSL query.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: piDetectionCreateNoWindow}
	return cmd.Output()
}
