//go:build !windows

package wsl

import "fmt"

func DetectPi(string) (PiDetection, error) {
	return PiDetection{}, fmt.Errorf("WSL Pi detection is supported only on Windows")
}
