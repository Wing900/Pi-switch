//go:build darwin

package pi

import (
	"os/exec"
)

func openCommandInTerminal(command string, workingDir string) error {
	return exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+command+`"`).Start()
}
