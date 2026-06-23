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

func OpenCommandInTerminal(command string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("powershell", "-NoExit", "-Command", command).Start()
	case "darwin":
		return exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+command+`"`).Start()
	default:
		return exec.Command("x-terminal-emulator", "-e", command).Start()
	}
}
