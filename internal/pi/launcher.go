package pi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type LaunchPreview struct {
	Command   string   `json:"command"`
	Checklist []string `json:"checklist"`
}

func BuildCommand(piCommand, providerID, modelID string) string {
	return piCommand + " --provider " + providerID + " --model " + modelID
}

func writeLaunchLog(workingDir, command string, err error) {
	logDir := filepath.Join(os.TempDir(), "piswitch-logs")
	_ = os.MkdirAll(logDir, 0o755)
	logPath := filepath.Join(logDir, fmt.Sprintf("launch-%s.log", time.Now().Format("20060102-150405")))
	content := fmt.Sprintf("time: %s\nworkingDir: %s\ncommand: %s\nerror: %v\n",
		time.Now().Format(time.RFC3339), workingDir, command, err)
	_ = os.WriteFile(logPath, []byte(content), 0o644)
}

func OpenCommandInTerminal(command string, workingDir string) error {
	var err error
	defer func() { writeLaunchLog(workingDir, command, err) }()

	switch runtime.GOOS {
	case "windows":
		// 新开 cmd 窗口, /k 保持窗口不关闭
		cmd := exec.Command("cmd", "/c", "start", "\"Pi Agent\"", "cmd", "/k", command)
		if workingDir != "" {
			cmd.Dir = workingDir
		}
		err = cmd.Start()
		return err
	case "darwin":
		err = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+command+`"`).Start()
		return err
	default:
		cmd := exec.Command("x-terminal-emulator", "-e", command)
		if workingDir != "" {
			cmd.Dir = workingDir
		}
		err = cmd.Start()
		return err
	}
}
