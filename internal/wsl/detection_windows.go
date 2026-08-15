//go:build windows

package wsl

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
)

const createNoWindow uint32 = 0x08000000

func Detect() (Detection, error) {
	unavailable := Detection{Distros: []string{}}

	cmd := exec.Command("wsl.exe", "--list", "--quiet")
	// Prevent a console window for this noninteractive WSL query.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
	output, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return unavailable, nil
		}
		return unavailable, fmt.Errorf("查询 WSL 发行版失败: %w", err)
	}

	distros, err := parseDistroList(output)
	if err != nil {
		return unavailable, fmt.Errorf("无法解析 WSL 发行版列表: %w", err)
	}
	return Detection{Detected: true, Distros: distros}, nil
}
