package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Type            string                     `json:"type"`
	BaseURL         string                     `json:"baseUrl"`
	APIKeyEnv       string                     `json:"apiKeyEnv"`
	APIKeyLiteral   string                     `json:"apiKeyLiteral"`
	API             string                     `json:"api"`
	Proxy           string                     `json:"proxy"`
	Headers         map[string]string          `json:"headers"`
	Models          []ModelInfo                `json:"models"`
	Host            string                     `json:"host"`
	SelectedModelID string                     `json:"selectedModelId"`
	Extra           map[string]json.RawMessage `json:"-"`
}

var configFields = map[string]struct{}{
	"id": {}, "name": {}, "type": {}, "baseUrl": {}, "apiKeyEnv": {},
	"apiKeyLiteral": {}, "api": {}, "proxy": {}, "headers": {},
	"models": {}, "host": {}, "selectedModelId": {},
}

func (cfg *Config) UnmarshalJSON(data []byte) error {
	type configAlias Config
	var decoded configAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for field := range configFields {
		delete(raw, field)
	}
	*cfg = Config(decoded)
	cfg.Extra = raw
	return nil
}

func (cfg Config) MarshalJSON() ([]byte, error) {
	type configAlias Config
	data, err := json.Marshal(configAlias(cfg))
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for field, value := range cfg.Extra {
		if _, known := configFields[field]; !known {
			fields[field] = value
		}
	}
	return json.Marshal(fields)
}

type ModelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Reasoning     bool   `json:"reasoning"`
	ContextWindow int    `json:"contextWindow,omitempty"`
	MaxTokens     int    `json:"maxTokens,omitempty"`
}

type ConnectionTestResult struct {
	OK    bool     `json:"ok"`
	Title string   `json:"title"`
	Lines []string `json:"lines"`
}

func Normalize(input Config) Config {
	input.ID = strings.TrimSpace(input.ID)
	input.Name = strings.TrimSpace(input.Name)
	input.BaseURL = strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
	input.APIKeyEnv = strings.TrimSpace(input.APIKeyEnv)
	input.APIKeyLiteral = strings.TrimSpace(input.APIKeyLiteral)
	input.API = strings.TrimSpace(input.API)
	input.Models = NormalizeModels(input.Models)
	if input.Type == "" {
		input.Type = "openai-compatible"
	}
	if input.Headers == nil {
		input.Headers = map[string]string{}
	}
	if input.Host == "" {
		input.Host = deriveHost(input.BaseURL)
	}
	if !hasModel(input.Models, input.SelectedModelID) && len(input.Models) > 0 {
		input.SelectedModelID = input.Models[0].ID
	}
	return input
}

func NormalizeModels(models []ModelInfo) []ModelInfo {
	if len(models) == 0 {
		return nil
	}

	normalized := make([]ModelInfo, 0, len(models))
	seen := make(map[string]int, len(models))
	for _, model := range models {
		model.ID = strings.TrimSpace(model.ID)
		model.Name = strings.TrimSpace(model.Name)
		if model.ID == "" {
			continue
		}
		if model.Name == "" {
			model.Name = model.ID
		}
		if existingIndex, ok := seen[model.ID]; ok {
			normalized[existingIndex] = model
			continue
		}
		seen[model.ID] = len(normalized)
		normalized = append(normalized, model)
	}
	return normalized
}

func MergeModels(existing []ModelInfo, incoming []ModelInfo) []ModelInfo {
	merged := NormalizeModels(existing)
	next := NormalizeModels(incoming)
	if len(next) == 0 {
		return merged
	}

	indexByID := make(map[string]int, len(merged))
	for index, model := range merged {
		indexByID[model.ID] = index
	}

	for _, model := range next {
		if existingIndex, ok := indexByID[model.ID]; ok {
			merged[existingIndex] = model
			continue
		}
		indexByID[model.ID] = len(merged)
		merged = append(merged, model)
	}
	return merged
}

func deriveHost(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx]
	}
	if trimmed == "" {
		return "未配置"
	}
	return trimmed
}

func EnsureModel(config Config, modelID string) error {
	if hasModel(config.Models, modelID) {
		return nil
	}
	return errors.New("模型不存在：" + modelID)
}

func hasModel(models []ModelInfo, modelID string) bool {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return false
	}
	for _, model := range models {
		if model.ID == modelID {
			return true
		}
	}
	return false
}

func FormatModelCount(count int) string {
	return fmt.Sprintf("%d 个", count)
}
