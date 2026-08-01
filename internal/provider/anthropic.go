package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const anthropicVersion = "2023-06-01"

type anthropicModelsResponse struct {
	Data []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"data"`
}

// FetchAnthropicModels 通过 Anthropic 官方 /v1/models 端点拉取模型。
// 鉴权使用 x-api-key + anthropic-version header（区别于 OpenAI 兼容的 Bearer）。
func FetchAnthropicModels(cfg Config, apiKey string) ([]ModelInfo, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("Base URL 为空")
	}
	url := strings.TrimRight(cfg.BaseURL, "/")
	if strings.HasSuffix(url, "/v1") {
		url += "/models"
	} else {
		url += "/v1/models"
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}
	req.Header.Set("anthropic-version", anthropicVersion)
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("服务返回状态码 " + resp.Status)
	}

	var payload anthropicModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		models = append(models, ModelInfo{
			ID:        item.ID,
			Name:      item.DisplayName,
			Reasoning: inferReasoning(item.ID),
		})
	}
	if len(models) == 0 {
		return nil, errors.New("未从 /v1/models 返回中解析到模型")
	}
	return models, nil
}

// FetchModelsByAPI 根据 provider 的 API 类型路由到对应的模型拉取实现。
// 支持 anthropic-messages（Anthropic 原生 API）与 openai-completions / openai-responses（OpenAI 兼容）。
func FetchModelsByAPI(cfg Config, apiKey string) ([]ModelInfo, error) {
	switch cfg.API {
	case "anthropic-messages", "anthropic":
		return FetchAnthropicModels(cfg, apiKey)
	default:
		return FetchOpenAICompatibleModels(cfg, apiKey)
	}
}
