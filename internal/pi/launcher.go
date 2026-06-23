package pi

import (
	"os/exec"
	"runtime"
	"syscall"
)

type LaunchPreview struct {
	Command   string   `json:"command"`
	Checklist []string `json:"checklist"`
}

func BuildCommand(piCommand, providerID, modelID string) string {
	return piCommand + " --provider " + providerID + " --model " + modelID
}

func OpenCommandInTerminal(command string, workingDir string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "powershell", "-NoExit", "-Command", command)
		if workingDir != "" {
			cmd.Dir = workingDir
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000200, // DETACHED_PROCESS
		}
		return cmd.Start()
	case "darwin":
		return exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+command+`"`).Start()
	default:
		cmd := exec.Command("x-terminal-emulator", "-e", command)
		if workingDir != "" {
			cmd.Dir = workingDir
		}
		return cmd.Start()
	}
}
