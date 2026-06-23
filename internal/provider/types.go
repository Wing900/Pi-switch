package provider

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	BaseURL         string            `json:"baseUrl"`
	APIKeyEnv       string            `json:"apiKeyEnv"`
	APIKeyLiteral   string            `json:"apiKeyLiteral"`
	API             string            `json:"api"`
	Proxy           string            `json:"proxy"`
	Headers         map[string]string `json:"headers"`
	Models          []ModelInfo       `json:"models"`
	Host            string            `json:"host"`
	SelectedModelID string            `json:"selectedModelId"`
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
	if input.Type == "" {
		input.Type = "openai-compatible"
	}
	if input.Headers == nil {
		input.Headers = map[string]string{}
	}
	if input.Host == "" {
		input.Host = deriveHost(input.BaseURL)
	}
	if input.SelectedModelID == "" && len(input.Models) > 0 {
		input.SelectedModelID = input.Models[0].ID
	}
	return input
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
	for _, model := range config.Models {
		if model.ID == modelID {
			return nil
		}
	}
	return errors.New("模型不存在：" + modelID)
}

func FormatModelCount(count int) string {
	return fmt.Sprintf("%d 个", count)
}
