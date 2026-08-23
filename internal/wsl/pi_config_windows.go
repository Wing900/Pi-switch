//go:build windows

package wsl

import (
	"os/exec"
	"syscall"
)

const piConfigCreateNoWindow uint32 = 0x08000000

var piConfigCommand = exec.Command

func ReadPiConfigDocuments(distro string) (PiConfigDocuments, error) {
	return readPiConfigDocuments(distro, DetectPi, readPiConfigFile)
}

func readPiConfigFile(distro, linuxPath string) (string, error) {
	cmd := piConfigCommand("wsl.exe", "--distribution", distro, "--exec", "cat", linuxPath)
	// Prevent a console window for this noninteractive WSL read.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: piConfigCreateNoWindow}
	return consumePiConfigCommand(cmd, piConfigDocumentLimit)
}
