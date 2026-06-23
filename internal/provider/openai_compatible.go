package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func FetchOpenAICompatibleModels(cfg Config, apiKey string) ([]ModelInfo, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("Base URL 为空")
	}
	url := strings.TrimRight(cfg.BaseURL, "/") + "/models"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
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

	var payload modelsResponse
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
			Name:      item.ID,
			Reasoning: inferReasoning(item.ID),
		})
	}
	if len(models) == 0 {
		return nil, errors.New("未从 /models 返回中解析到模型")
	}
	return models, nil
}

func inferReasoning(id string) bool {
	lower := strings.ToLower(id)
	for _, token := range []string{"reasoner", "reasoning", "r1", "thinking"} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}
