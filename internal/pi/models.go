package pi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"piswitch/internal/provider"
	"piswitch/internal/system"
)

type providerPayload struct {
	BaseURL string               `json:"baseUrl"`
	API     string               `json:"api"`
	APIKey  string               `json:"apiKey"`
	Headers map[string]string    `json:"headers,omitempty"`
	Models  []provider.ModelInfo `json:"models"`
}

// WriteAllModelsOMP 以 Oh My Pi (omp) 的 models.yml 格式写入 provider 与模型定义，
// 并做增量合并：保留文件中原有、不由 Pi Switch 管理的顶层键（如 equivalence）
// 与其他 provider（用户手写），仅 upsert 本工具管理的 provider，避免覆盖 OMP 现有配置。
// apiKey 值采用 omp 语义：直接写环境变量名（无 $ 前缀），omp 解析时先查 env 再回退字面量。
func WriteAllModelsOMP(path string, providersConfig []provider.Config) error {
	if err := system.BackupFile(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var root yaml.Node
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("解析现有 models.yml 失败：%w", err)
		}
	}
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) > 0 {
			root = *root.Content[0]
		}
	}
	if root.Kind == 0 {
		root.Kind = yaml.MappingNode
	}
	if root.Kind != yaml.MappingNode {
		return errors.New("models.yml 顶层必须是 mapping")
	}

	providersNode := yamlMappingValue(&root, "providers")
	if providersNode == nil {
		providersNode = &yaml.Node{Kind: yaml.MappingNode}
		yamlSetMapping(&root, "providers", providersNode)
	}
	if providersNode.Kind != yaml.MappingNode {
		return errors.New("models.yml 的 providers 必须是 mapping")
	}

	for _, cfg := range providersConfig {
		existing := yamlMappingValue(providersNode, cfg.ID)
		updated := buildProviderYAMLNode(cfg, existing)
		yamlSetMapping(providersNode, cfg.ID, updated)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&root); err != nil {
		return err
	}
	_ = enc.Close()
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// buildProviderYAMLNode 构造 provider 的 YAML mapping 节点。
// 保留现有节点中不由 Pi Switch 管理的键（headers/compat/discovery/auth/transport 等），
// 覆盖 baseUrl / api / apiKey / headers / models。
func buildProviderYAMLNode(cfg provider.Config, existing *yaml.Node) *yaml.Node {
	result := &yaml.Node{Kind: yaml.MappingNode}
	managed := map[string]bool{"baseUrl": true, "api": true, "apiKey": true, "headers": true, "models": true}
	if existing != nil && existing.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(existing.Content); i += 2 {
			key := existing.Content[i].Value
			if !managed[key] {
				result.Content = append(result.Content, existing.Content[i], existing.Content[i+1])
			}
		}
	}

	yamlAppendScalarPair(result, "baseUrl", cfg.BaseURL)
	yamlAppendScalarPair(result, "api", cfg.API)

	key := cfg.APIKeyLiteral
	if cfg.APIKeyEnv != "" {
		key = cfg.APIKeyEnv
	}
	if key != "" {
		yamlAppendScalarPair(result, "apiKey", key)
	}
	if len(cfg.Headers) > 0 {
		headers := &yaml.Node{Kind: yaml.MappingNode}
		for k, v := range cfg.Headers {
			yamlAppendScalarPair(headers, k, v)
		}
		result.Content = append(result.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "headers"}, headers)
	}
	if len(cfg.Models) > 0 {
		models := &yaml.Node{Kind: yaml.SequenceNode}
		for _, m := range cfg.Models {
			entry := &yaml.Node{Kind: yaml.MappingNode}
			yamlAppendScalarPair(entry, "id", m.ID)
			if m.Name != "" && m.Name != m.ID {
				yamlAppendScalarPair(entry, "name", m.Name)
			}
			if m.Reasoning {
				yamlAppendScalarPair(entry, "reasoning", true)
			}
			if m.ContextWindow > 0 {
				yamlAppendScalarPair(entry, "contextWindow", m.ContextWindow)
			}
			if m.MaxTokens > 0 {
				yamlAppendScalarPair(entry, "maxTokens", m.MaxTokens)
			}
			models.Content = append(models.Content, entry)
		}
		result.Content = append(result.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: "models"}, models)
	}
	return result
}

func yamlAppendScalarPair(m *yaml.Node, key string, value any) {
	var valueNode yaml.Node
	if err := valueNode.Encode(value); err != nil {
		valueNode = yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", value)}
	}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, &valueNode)
}

func yamlMappingValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

func yamlSetMapping(m *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return
		}
	}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key}, value)
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
			Headers: cfg.Headers,
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
