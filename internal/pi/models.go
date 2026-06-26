package pi

import (
	"encoding/json"
	"os"
	"path/filepath"

	"piswitch/internal/provider"
	"piswitch/internal/system"
)

type providerPayload struct {
	BaseURL string               `json:"baseUrl"`
	API     string               `json:"api"`
	APIKey  string               `json:"apiKey"`
	Models  []provider.ModelInfo `json:"models"`
}

func WriteModels(path string, cfg provider.Config) error {
	return WriteAllModels(path, []provider.Config{cfg})
}

func WriteAllModels(path string, providersConfig []provider.Config) error {
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

	providers := map[string]any{}
	for _, cfg := range providersConfig {
		key := cfg.APIKeyLiteral
		if cfg.APIKeyEnv != "" {
			key = "$" + cfg.APIKeyEnv
		}

		nextProvider := providerPayload{
			BaseURL: cfg.BaseURL,
			API:     cfg.API,
			APIKey:  key,
			Models:  cfg.Models,
		}
		encodedProvider, err := json.Marshal(nextProvider)
		if err != nil {
			return err
		}

		nextFields := map[string]any{}
		if err := json.Unmarshal(encodedProvider, &nextFields); err != nil {
			return err
		}
		providers[cfg.ID] = nextFields
	}
	payload["providers"] = providers

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
