//go:build !windows

package wsl

import "fmt"

func ReadPiConfigDocuments(string) (PiConfigDocuments, error) {
	return PiConfigDocuments{}, fmt.Errorf("WSL Pi configuration reading is supported only on Windows")
}
