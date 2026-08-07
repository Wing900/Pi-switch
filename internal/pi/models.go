package pi

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"

	"piswitch/internal/provider"
)

type providerPayload struct {
	BaseURL string               `json:"baseUrl"`
	API     string               `json:"api"`
	APIKey  string               `json:"apiKey"`
	Headers map[string]string    `json:"headers,omitempty"`
	Models  []provider.ModelInfo `json:"models"`
}

// ReadAllModels maps Pi's models.json into the provider shape used by the UI.
// Raw fields stay in models.json and are retained by every mutation below.
func ReadAllModels(path string) ([]provider.Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	payload := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	rawProviders, err := decodeProviders(payload)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rawProviders))
	for id := range rawProviders {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	providers := make([]provider.Config, 0, len(ids))
	for _, id := range ids {
		var raw providerPayload
		if err := json.Unmarshal(rawProviders[id], &raw); err != nil {
			return nil, err
		}

		cfg := provider.Config{
			ID:            id,
			Name:          id,
			Type:          "openai-compatible",
			BaseURL:       raw.BaseURL,
			API:           raw.API,
			Headers:       raw.Headers,
			Models:        provider.NormalizeModels(raw.Models),
			HeaderMode:    "none",
			CustomHeaders: map[string]string{},
		}
		if strings.Contains(strings.ToLower(raw.API), "anthropic") {
			cfg.Type = "anthropic"
		}
		if len(cfg.Headers) > 0 {
			cfg.HeaderMode = "custom"
			cfg.CustomHeaders = cloneHeaders(cfg.Headers)
		}
		if strings.HasPrefix(raw.APIKey, "$") && len(strings.TrimPrefix(raw.APIKey, "$")) > 0 {
			cfg.APIKeyEnv = strings.TrimPrefix(raw.APIKey, "$")
		} else {
			cfg.APIKeyLiteral = raw.APIKey
		}
		providers = append(providers, provider.Normalize(cfg))
	}

	return providers, nil
}

// UpsertProvider updates one Provider. Existing models and unknown fields are
// retained; a newly created Provider receives the models supplied by cfg.
func UpsertProvider(path string, cfg provider.Config, oldID string) error {
	cfg = provider.Normalize(cfg)
	return mutateJSONDocument(path, func(payload map[string]json.RawMessage) error {
		providers, err := decodeProviders(payload)
		if err != nil {
			return err
		}

		sourceID := cfg.ID
		if strings.TrimSpace(oldID) != "" {
			sourceID = oldID
		}
		existing, exists := providers[sourceID]
		merged, err := mergeProviderDocument(existing, cfg, !exists)
		if err != nil {
			return err
		}
		if sourceID != cfg.ID {
			delete(providers, sourceID)
		}
		providers[cfg.ID] = merged
		return encodeProviders(payload, providers)
	})
}

func DeleteProvider(path string, providerID string) error {
	return mutateJSONDocument(path, func(payload map[string]json.RawMessage) error {
		providers, err := decodeProviders(payload)
		if err != nil {
			return err
		}
		delete(providers, providerID)
		return encodeProviders(payload, providers)
	})
}

func MergeModels(path string, providerID string, incoming []provider.ModelInfo) ([]provider.ModelInfo, error) {
	var result []provider.ModelInfo
	err := mutateProviderModels(path, providerID, func(existing []provider.ModelInfo) []provider.ModelInfo {
		result = provider.MergeModels(existing, incoming)
		return result
	})
	return result, err
}

func ReplaceModels(path string, providerID string, models []provider.ModelInfo) ([]provider.ModelInfo, error) {
	result := provider.NormalizeModels(models)
	err := mutateProviderModels(path, providerID, func([]provider.ModelInfo) []provider.ModelInfo {
		return result
	})
	return result, err
}

func mutateProviderModels(path string, providerID string, mutate func([]provider.ModelInfo) []provider.ModelInfo) error {
	return mutateJSONDocument(path, func(payload map[string]json.RawMessage) error {
		providers, err := decodeProviders(payload)
		if err != nil {
			return err
		}
		rawProvider, exists := providers[providerID]
		if !exists {
			return errors.New("未找到 Provider：" + providerID)
		}

		fields := map[string]json.RawMessage{}
		if err := json.Unmarshal(rawProvider, &fields); err != nil {
			return err
		}
		existingModels, err := decodeModels(fields["models"])
		if err != nil {
			return err
		}
		nextModels := provider.NormalizeModels(mutate(existingModels))
		mergedModels, err := mergeModelDocuments(fields["models"], nextModels)
		if err != nil {
			return err
		}
		fields["models"] = mergedModels
		providers[providerID], err = json.Marshal(fields)
		if err != nil {
			return err
		}
		return encodeProviders(payload, providers)
	})
}

func decodeProviders(payload map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	providers := map[string]json.RawMessage{}
	if raw, ok := payload["providers"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &providers); err != nil {
			return nil, err
		}
	}
	return providers, nil
}

func encodeProviders(payload map[string]json.RawMessage, providers map[string]json.RawMessage) error {
	encoded, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	payload["providers"] = encoded
	return nil
}

func mergeProviderDocument(existing json.RawMessage, cfg provider.Config, includeModels bool) (json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &fields); err != nil {
			return nil, err
		}
	}

	key := cfg.APIKeyLiteral
	if cfg.APIKeyEnv != "" {
		key = "$" + cfg.APIKeyEnv
	}
	if err := setJSONField(fields, "baseUrl", cfg.BaseURL); err != nil {
		return nil, err
	}
	if err := setJSONField(fields, "api", cfg.API); err != nil {
		return nil, err
	}
	if err := setJSONField(fields, "apiKey", key); err != nil {
		return nil, err
	}
	if err := setJSONField(fields, "headers", cfg.Headers); err != nil {
		return nil, err
	}
	if includeModels {
		models, err := mergeModelDocuments(fields["models"], cfg.Models)
		if err != nil {
			return nil, err
		}
		fields["models"] = models
	}
	return json.Marshal(fields)
}

func mergeModelDocuments(existing json.RawMessage, models []provider.ModelInfo) (json.RawMessage, error) {
	existingByID := map[string]map[string]json.RawMessage{}
	if len(existing) > 0 {
		var rawModels []json.RawMessage
		if err := json.Unmarshal(existing, &rawModels); err != nil {
			return nil, err
		}
		for _, rawModel := range rawModels {
			fields := map[string]json.RawMessage{}
			if err := json.Unmarshal(rawModel, &fields); err != nil {
				return nil, err
			}
			var id string
			if err := json.Unmarshal(fields["id"], &id); err != nil || strings.TrimSpace(id) == "" {
				continue
			}
			existingByID[id] = fields
		}
	}

	merged := make([]json.RawMessage, 0, len(models))
	for _, model := range provider.NormalizeModels(models) {
		encodedModel, err := json.Marshal(model)
		if err != nil {
			return nil, err
		}
		nextFields := map[string]json.RawMessage{}
		if err := json.Unmarshal(encodedModel, &nextFields); err != nil {
			return nil, err
		}
		if oldFields := existingByID[model.ID]; oldFields != nil {
			for key, value := range oldFields {
				if _, exists := nextFields[key]; !exists {
					nextFields[key] = value
				}
			}
		}
		mergedModel, err := json.Marshal(nextFields)
		if err != nil {
			return nil, err
		}
		merged = append(merged, mergedModel)
	}
	return json.Marshal(merged)
}

func decodeModels(raw json.RawMessage) ([]provider.ModelInfo, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var models []provider.ModelInfo
	if err := json.Unmarshal(raw, &models); err != nil {
		return nil, err
	}
	return provider.NormalizeModels(models), nil
}

func cloneHeaders(input map[string]string) map[string]string {
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
