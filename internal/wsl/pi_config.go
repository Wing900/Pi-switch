package wsl

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"unicode/utf8"
)

const piConfigDocumentLimit int64 = 1 << 23

type PiConfigDocuments struct {
	Distro         string `json:"distro"`
	SettingsExists bool   `json:"settingsExists"`
	ModelsExists   bool   `json:"modelsExists"`
	SettingsJSON   string `json:"settingsJson"`
	ModelsJSON     string `json:"modelsJson"`
}

type piConfigDetector func(string) (PiDetection, error)
type piConfigFileReader func(distro, linuxPath string) (string, error)

func readPiConfigDocuments(distro string, detect piConfigDetector, readFile piConfigFileReader) (PiConfigDocuments, error) {
	detection, err := detect(distro)
	if err != nil {
		return PiConfigDocuments{}, fmt.Errorf("detect WSL Pi configuration: %w", err)
	}

	result := PiConfigDocuments{
		Distro:         detection.Distro,
		SettingsExists: detection.SettingsExists,
		ModelsExists:   detection.ModelsExists,
	}

	if detection.SettingsExists {
		settingsPath := path.Join(detection.PiHome, "agent", "settings.json")
		content, err := readFile(distro, settingsPath)
		if err != nil {
			return PiConfigDocuments{}, fmt.Errorf("read WSL Pi document %q: %w", settingsPath, err)
		}
		result.SettingsJSON = content
	}

	if detection.ModelsExists {
		modelsPath := path.Join(detection.PiHome, "agent", "models.json")
		content, err := readFile(distro, modelsPath)
		if err != nil {
			return PiConfigDocuments{}, fmt.Errorf("read WSL Pi document %q: %w", modelsPath, err)
		}
		result.ModelsJSON = content
	}

	return result, nil
}

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("document exceeds %d bytes", limit)
	}
	return data, nil
}

func decodePiConfigDocument(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", fmt.Errorf("document is not valid UTF-8")
	}
	return string(data), nil
}

func consumePiConfigCommand(cmd *exec.Cmd, limit int64) (string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("open document stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start document read: %w", err)
	}

	data, readErr := readBounded(stdout, limit)
	if readErr != nil {
		_ = stdout.Close()
		_ = cmd.Wait()
		return "", readErr
	}
	if err := cmd.Wait(); err != nil {
		return "", err
	}

	text, err := decodePiConfigDocument(data)
	if err != nil {
		return "", err
	}
	return text, nil
}
