//go:build !windows

package wsl

func Detect() (Detection, error) {
	return Detection{Distros: []string{}}, nil
}
