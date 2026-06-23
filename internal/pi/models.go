package pi

import (
	"encoding/json"
	"os"
	"path/filepath"

	"piswitch/internal/provider"
	"piswitch/internal/system"
)

type modelsEnvelope struct {
	Providers map[string]providerPayload `json:"providers"`
}

type providerPayload struct {
	BaseURL string               `json:"baseUrl"`
	API     string               `json:"api"`
	APIKey  string               `json:"apiKey"`
	Models  []provider.ModelInfo `json:"models"`
}

func WriteModels(path string, cfg provider.Config) error {
	if err := system.BackupFile(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	key := cfg.APIKeyLiteral
	if cfg.APIKeyEnv != "" {
		key = "$" + cfg.APIKeyEnv
	}
	payload := modelsEnvelope{
		Providers: map[string]providerPayload{
			cfg.ID: {
				BaseURL: cfg.BaseURL,
				API:     cfg.API,
				APIKey:  key,
				Models:  cfg.Models,
			},
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
