package pi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	return mutateProviderModels(path, providerID, func(existing []provider.ModelInfo) ([]provider.ModelInfo, error) {
		return provider.MergeModels(existing, incoming), nil
	}, modelDocumentMergeIncremental)
}

func ReplaceModels(path string, providerID string, models []provider.ModelInfo, expectedRevisions ...string) ([]provider.ModelInfo, error) {
	result := provider.NormalizeModels(models)
	expectedRevision := ""
	if len(expectedRevisions) > 0 {
		expectedRevision = expectedRevisions[0]
	}
	return mutateProviderModels(path, providerID, func(existing []provider.ModelInfo) ([]provider.ModelInfo, error) {
		if expectedRevision != "" && provider.ModelListRevision(existing) != expectedRevision {
			return nil, errors.New("模型列表已被外部修改，请重新打开编辑器")
		}
		return result, nil
	}, modelDocumentReplaceList)
}

type modelDocumentMergeMode int

const (
	modelDocumentMergeIncremental modelDocumentMergeMode = iota
	modelDocumentReplaceList
)

func mutateProviderModels(path string, providerID string, mutate func([]provider.ModelInfo) ([]provider.ModelInfo, error), mode modelDocumentMergeMode) ([]provider.ModelInfo, error) {
	var persisted []provider.ModelInfo
	err := mutateJSONDocument(path, func(payload map[string]json.RawMessage) error {
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
		next, err := mutate(existingModels)
		if err != nil {
			return err
		}
		nextModels := provider.NormalizeModels(next)
		mergedModels, err := mergeModelDocuments(fields["models"], nextModels, mode)
		if err != nil {
			return err
		}
		fields["models"] = mergedModels
		persisted, err = decodeModels(mergedModels)
		if err != nil {
			return err
		}
		providers[providerID], err = json.Marshal(fields)
		if err != nil {
			return err
		}
		return encodeProviders(payload, providers)
	})
	return persisted, err
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
		models, err := mergeModelDocuments(fields["models"], cfg.Models, modelDocumentReplaceList)
		if err != nil {
			return nil, err
		}
		fields["models"] = models
	}
	return json.Marshal(fields)
}

var knownModelDocumentFields = map[string]struct{}{
	"id": {}, "name": {}, "api": {}, "baseUrl": {}, "reasoning": {}, "thinkingLevelMap": {}, "input": {},
	"cost": {}, "contextWindow": {}, "maxTokens": {}, "samplingParams": {}, "headers": {}, "compat": {},
}

func mergeModelDocuments(existing json.RawMessage, models []provider.ModelInfo, mode modelDocumentMergeMode) (json.RawMessage, error) {
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
		delete(nextFields, "selected")
		replaceDocument := model.ReplaceDocument
		originalID := strings.TrimSpace(model.OriginalID)
		oldFields := existingByID[model.ID]
		if oldFields == nil && originalID != "" {
			oldFields = existingByID[originalID]
		}
		if model.Revision != "" {
			if oldFields == nil || modelDocumentRevision(oldFields) != model.Revision {
				return nil, fmt.Errorf("模型 %s 已被外部修改，请重新打开编辑器", model.ID)
			}
		}
		if !replaceDocument && oldFields != nil {
			if rawCompat, ok := oldFields["compat"]; ok {
				if err := mergeCompatFields(nextFields, rawCompat, mode == modelDocumentMergeIncremental); err != nil {
					return nil, err
				}
			}
			for key, value := range oldFields {
				if key == "selected" {
					continue
				}
				if mode == modelDocumentReplaceList {
					if _, known := knownModelDocumentFields[key]; known {
						continue
					}
				}
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

func mergeCompatFields(modelFields map[string]json.RawMessage, existingCompat json.RawMessage, preserveKnown bool) error {
	oldFields := map[string]json.RawMessage{}
	if err := json.Unmarshal(existingCompat, &oldFields); err != nil {
		return err
	}
	compatFields := map[string]json.RawMessage{}
	if rawCompat, ok := modelFields["compat"]; ok {
		if err := json.Unmarshal(rawCompat, &compatFields); err != nil {
			return err
		}
	}
	for field, value := range oldFields {
		if !preserveKnown && provider.IsTransportCompatField(field) {
			continue
		}
		if _, exists := compatFields[field]; !exists {
			compatFields[field] = value
		}
	}
	if len(compatFields) == 0 {
		delete(modelFields, "compat")
		return nil
	}
	encoded, err := json.Marshal(compatFields)
	if err != nil {
		return err
	}
	modelFields["compat"] = encoded
	return nil
}

func modelDocumentRevision(fields map[string]json.RawMessage) string {
	encoded, err := json.Marshal(fields)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
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
