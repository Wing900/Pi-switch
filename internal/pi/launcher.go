package pi

import (
	"os/exec"
	"runtime"
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
		cmd := exec.Command("powershell", "-NoExit", "-Command", command)
		if workingDir != "" {
			cmd.Dir = workingDir
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
