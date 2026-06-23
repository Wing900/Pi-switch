package pi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"piswitch/internal/system"
)

type DefaultSettings struct {
	DefaultProvider      string `json:"defaultProvider"`
	DefaultModel         string `json:"defaultModel"`
	DefaultThinkingLevel string `json:"defaultThinkingLevel"`
}

func ReadDefaults(path string) (DefaultSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultSettings{}, nil
		}
		return DefaultSettings{}, err
	}
	var payload DefaultSettings
	if err := json.Unmarshal(data, &payload); err != nil {
		return DefaultSettings{}, err
	}
	return payload, nil
}

func MergeDefaults(path string, defaults DefaultSettings) error {
	if err := system.BackupFile(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &payload)
	}
	payload["defaultProvider"] = defaults.DefaultProvider
	payload["defaultModel"] = defaults.DefaultModel
	payload["defaultThinkingLevel"] = defaults.DefaultThinkingLevel

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
