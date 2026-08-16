package provider

import (
	"crypto/sha256"
	"encoding/hex"
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
	HeaderMode      string                     `json:"headerMode"`
	Headers         map[string]string          `json:"headers"`
	CustomHeaders   map[string]string          `json:"customHeaders,omitempty"`
	Models          []ModelInfo                `json:"models"`
	Host            string                     `json:"host"`
	SelectedModelID string                     `json:"selectedModelId"`
	Extra           map[string]json.RawMessage `json:"-"`
}

var configFields = map[string]struct{}{
	"id": {}, "name": {}, "type": {}, "baseUrl": {}, "apiKeyEnv": {},
	"apiKeyLiteral": {}, "api": {}, "proxy": {}, "headerMode": {}, "headers": {}, "customHeaders": {},
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
	ID               string                     `json:"id"`
	Name             string                     `json:"name"`
	API              string                     `json:"api,omitempty"`
	BaseURL          string                     `json:"baseUrl,omitempty"`
	Reasoning        bool                       `json:"reasoning"`
	ThinkingLevelMap map[string]any             `json:"thinkingLevelMap,omitempty"`
	Input            []string                   `json:"input,omitempty"`
	Cost             map[string]any             `json:"cost,omitempty"`
	ContextWindow    int                        `json:"contextWindow,omitempty"`
	MaxTokens        int                        `json:"maxTokens,omitempty"`
	SamplingParams   map[string]any             `json:"samplingParams,omitempty"`
	Headers          map[string]string          `json:"headers,omitempty"`
	Compat           map[string]any             `json:"compat,omitempty"`
	Extra            map[string]json.RawMessage `json:"-"`
	CompatRaw        map[string]json.RawMessage `json:"-"`
	Revision         string                     `json:"-"`
	ReplaceDocument  bool                       `json:"-"`
	OriginalID       string                     `json:"-"`
}

var modelFields = map[string]struct{}{
	"id": {}, "name": {}, "api": {}, "baseUrl": {}, "reasoning": {}, "thinkingLevelMap": {}, "input": {},
	"cost": {}, "contextWindow": {}, "maxTokens": {}, "samplingParams": {}, "headers": {}, "compat": {},
}

var transportCompatFields = map[string]struct{}{
	"supportsStore": {}, "supportsDeveloperRole": {}, "supportsReasoningEffort": {}, "supportsUsageInStreaming": {},
	"supportsFinishReason": {}, "requiresToolResultName": {}, "requiresAssistantAfterToolResult": {},
	"requiresThinkingAsText": {}, "requiresReasoningContentOnAssistantMessages": {}, "supportsOpenAIGrammarTools": {},
	"supportsStrictMode": {}, "sendSessionAffinityHeaders": {}, "supportsLongCacheRetention": {},
	"supportsEagerToolInputStreaming": {}, "supportsCacheControlOnTools": {}, "supportsTemperature": {},
	"forceAdaptiveThinking": {}, "allowEmptySignature": {}, "supportsStrictTools": {}, "supportsToolReferences": {},
	"supportsToolSearch": {}, "zaiToolStream": {}, "supportsThinkingTokenBudget": {}, "supportsAdditionalTools": {},
	"supportsExplicitPromptCacheMode": {}, "maxTokensField": {}, "thinkingFormat": {}, "cacheControlFormat": {},
	"deferredToolsMode": {}, "sessionAffinityFormat": {}, "chatTemplateKwargs": {}, "chatTemplateArgs": {},
	"openRouterRouting": {}, "vercelGatewayRouting": {},
}

// UnmarshalJSON keeps model-specific fields in their original JSON form. The
// raw values are the persistence source of truth; transport conversion is kept
// separate so a real model field can never collide with Pi Switch metadata.
func (model *ModelInfo) UnmarshalJSON(data []byte) error {
	type modelAlias ModelInfo
	var decoded modelAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	canonical, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	compatRaw := map[string]json.RawMessage{}
	structuredCompat := map[string]any{}
	if value, ok := raw["compat"]; ok && len(value) > 0 {
		if err := json.Unmarshal(value, &compatRaw); err != nil {
			return err
		}
		for field, rawValue := range compatRaw {
			if _, known := transportCompatFields[field]; !known {
				continue
			}
			var decodedValue any
			if err := json.Unmarshal(rawValue, &decodedValue); err != nil {
				return err
			}
			structuredCompat[field] = decodedValue
		}
	}
	for field := range modelFields {
		delete(raw, field)
	}
	*model = ModelInfo(decoded)
	if len(structuredCompat) == 0 {
		model.Compat = nil
	} else {
		model.Compat = structuredCompat
	}
	model.Extra = raw
	model.CompatRaw = compatRaw
	hash := sha256.Sum256(canonical)
	model.Revision = hex.EncodeToString(hash[:])
	return nil
}

func (model ModelInfo) MarshalJSON() ([]byte, error) {
	type modelAlias ModelInfo
	data, err := json.Marshal(modelAlias(model))
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	compatFields := cloneRawFields(model.CompatRaw)
	if len(model.Compat) > 0 {
		encodedCompat, err := json.Marshal(model.Compat)
		if err != nil {
			return nil, err
		}
		structuredCompat := map[string]json.RawMessage{}
		if err := json.Unmarshal(encodedCompat, &structuredCompat); err != nil {
			return nil, err
		}
		for field, value := range structuredCompat {
			compatFields[field] = value
		}
	}
	if len(compatFields) == 0 {
		delete(fields, "compat")
	} else {
		fields["compat"], err = json.Marshal(compatFields)
		if err != nil {
			return nil, err
		}
	}

	for field, value := range model.Extra {
		if _, known := modelFields[field]; !known {
			fields[field] = value
		}
	}
	return json.Marshal(fields)
}

// ModelTransport is the Wails-facing model DTO. Unknown model fields are sent
// as raw JSON text so JavaScript never coerces large integers, and transport
// metadata stays outside the models.json namespace.
type ModelTransport struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	API                   string            `json:"api,omitempty"`
	BaseURL               string            `json:"baseUrl,omitempty"`
	Reasoning             bool              `json:"reasoning"`
	ThinkingLevelMap      map[string]any    `json:"thinkingLevelMap,omitempty"`
	Input                 []string          `json:"input,omitempty"`
	Cost                  map[string]any    `json:"cost,omitempty"`
	ContextWindow         int               `json:"contextWindow,omitempty"`
	MaxTokens             int               `json:"maxTokens,omitempty"`
	SamplingParams        map[string]any    `json:"samplingParams,omitempty"`
	Headers               map[string]string `json:"headers,omitempty"`
	Compat                map[string]any    `json:"compat,omitempty"`
	CompatExtraFieldsJSON string            `json:"compatExtraFieldsJson,omitempty"`
	ExtraFieldsJSON       string            `json:"extraFieldsJson,omitempty"`
	Revision              string            `json:"revision,omitempty"`
	ReplaceDocument       bool              `json:"replaceDocument,omitempty"`
	OriginalID            string            `json:"originalId,omitempty"`
}

func IsTransportCompatField(field string) bool {
	_, known := transportCompatFields[field]
	return known
}

func NewModelTransport(model ModelInfo) (ModelTransport, error) {
	compat, compatExtra, err := splitCompatForTransport(model)
	if err != nil {
		return ModelTransport{}, err
	}
	extraFields := cloneRawFields(model.Extra)
	delete(extraFields, "selected")
	extraJSON, err := marshalRawObject(extraFields)
	if err != nil {
		return ModelTransport{}, err
	}
	return ModelTransport{
		ID: model.ID, Name: model.Name, API: model.API, BaseURL: model.BaseURL, Reasoning: model.Reasoning,
		ThinkingLevelMap: model.ThinkingLevelMap, Input: model.Input, Cost: model.Cost,
		ContextWindow: model.ContextWindow, MaxTokens: model.MaxTokens, SamplingParams: model.SamplingParams,
		Headers: model.Headers, Compat: compat, CompatExtraFieldsJSON: compatExtra,
		ExtraFieldsJSON: extraJSON, Revision: model.Revision,
	}, nil
}

func (transport ModelTransport) ModelInfo() (ModelInfo, error) {
	extra, err := parseRawObject(transport.ExtraFieldsJSON)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("其他模型字段无效：%w", err)
	}
	for field := range modelFields {
		delete(extra, field)
	}
	delete(extra, "selected")

	compatRaw, err := parseRawObject(transport.CompatExtraFieldsJSON)
	if err != nil {
		return ModelInfo{}, fmt.Errorf("其他 compat 字段无效：%w", err)
	}
	for field := range transportCompatFields {
		delete(compatRaw, field)
	}
	return ModelInfo{
		ID: transport.ID, Name: transport.Name, API: transport.API, BaseURL: transport.BaseURL,
		Reasoning: transport.Reasoning, ThinkingLevelMap: transport.ThinkingLevelMap, Input: transport.Input,
		Cost: transport.Cost, ContextWindow: transport.ContextWindow, MaxTokens: transport.MaxTokens,
		SamplingParams: transport.SamplingParams, Headers: transport.Headers, Compat: transport.Compat,
		Extra: extra, CompatRaw: compatRaw, Revision: transport.Revision,
		ReplaceDocument: transport.ReplaceDocument, OriginalID: transport.OriginalID,
	}, nil
}

func ModelTransports(models []ModelInfo) ([]ModelTransport, error) {
	result := make([]ModelTransport, 0, len(models))
	for _, model := range models {
		converted, err := NewModelTransport(model)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func ModelsFromTransport(models []ModelTransport) ([]ModelInfo, error) {
	result := make([]ModelInfo, 0, len(models))
	for _, model := range models {
		converted, err := model.ModelInfo()
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

type ModelListTransport struct {
	Models   []ModelTransport `json:"models"`
	Revision string           `json:"revision"`
}

func NewModelListTransport(models []ModelInfo) (ModelListTransport, error) {
	converted, err := ModelTransports(models)
	if err != nil {
		return ModelListTransport{}, err
	}
	return ModelListTransport{Models: converted, Revision: ModelListRevision(models)}, nil
}

func ModelListRevision(models []ModelInfo) string {
	entries := make([][2]string, 0, len(models))
	for _, model := range models {
		if model.Revision == "" {
			return ""
		}
		entries = append(entries, [2]string{model.ID, model.Revision})
	}
	encoded, err := json.Marshal(entries)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:])
}

func splitCompatForTransport(model ModelInfo) (map[string]any, string, error) {
	structured := map[string]any{}
	extra := cloneRawFields(model.CompatRaw)
	for field, value := range model.CompatRaw {
		if _, known := transportCompatFields[field]; !known {
			continue
		}
		var decoded any
		if err := json.Unmarshal(value, &decoded); err != nil {
			return nil, "", err
		}
		structured[field] = decoded
		delete(extra, field)
	}
	for field, value := range model.Compat {
		if _, known := transportCompatFields[field]; known {
			structured[field] = value
			delete(extra, field)
			continue
		}
		if _, exists := extra[field]; !exists {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, "", err
			}
			extra[field] = encoded
		}
	}
	extraJSON, err := marshalRawObject(extra)
	if len(structured) == 0 {
		structured = nil
	}
	return structured, extraJSON, err
}

func parseRawObject(value string) (map[string]json.RawMessage, error) {
	if strings.TrimSpace(value) == "" {
		return map[string]json.RawMessage{}, nil
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(value), &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func marshalRawObject(fields map[string]json.RawMessage) (string, error) {
	if len(fields) == 0 {
		return "", nil
	}
	encoded, err := json.MarshalIndent(fields, "", "  ")
	return string(encoded), err
}

func cloneRawFields(input map[string]json.RawMessage) map[string]json.RawMessage {
	cloned := make(map[string]json.RawMessage, len(input))
	for field, value := range input {
		cloned[field] = value
	}
	return cloned
}

type ConfigTransport struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	BaseURL         string            `json:"baseUrl"`
	APIKeyEnv       string            `json:"apiKeyEnv"`
	APIKeyLiteral   string            `json:"apiKeyLiteral"`
	API             string            `json:"api"`
	Proxy           string            `json:"proxy"`
	HeaderMode      string            `json:"headerMode"`
	Headers         map[string]string `json:"headers"`
	CustomHeaders   map[string]string `json:"customHeaders,omitempty"`
	Models          []ModelTransport  `json:"models"`
	Host            string            `json:"host"`
	SelectedModelID string            `json:"selectedModelId"`
	ExtraFieldsJSON string            `json:"extraFieldsJson,omitempty"`
	ModelsRevision  string            `json:"modelsRevision,omitempty"`
}

func NewConfigTransport(cfg Config) (ConfigTransport, error) {
	models, err := ModelTransports(cfg.Models)
	if err != nil {
		return ConfigTransport{}, err
	}
	extraJSON, err := marshalRawObject(cfg.Extra)
	if err != nil {
		return ConfigTransport{}, err
	}
	return ConfigTransport{
		ID: cfg.ID, Name: cfg.Name, Type: cfg.Type, BaseURL: cfg.BaseURL,
		APIKeyEnv: cfg.APIKeyEnv, APIKeyLiteral: cfg.APIKeyLiteral, API: cfg.API, Proxy: cfg.Proxy,
		HeaderMode: cfg.HeaderMode, Headers: cfg.Headers, CustomHeaders: cfg.CustomHeaders,
		Models: models, Host: cfg.Host, SelectedModelID: cfg.SelectedModelID, ExtraFieldsJSON: extraJSON,
		ModelsRevision: ModelListRevision(cfg.Models),
	}, nil
}

func (transport ConfigTransport) Config() (Config, error) {
	models, err := ModelsFromTransport(transport.Models)
	if err != nil {
		return Config{}, err
	}
	extra, err := parseRawObject(transport.ExtraFieldsJSON)
	if err != nil {
		return Config{}, fmt.Errorf("其他 Provider 字段无效：%w", err)
	}
	for field := range configFields {
		delete(extra, field)
	}
	return Config{
		ID: transport.ID, Name: transport.Name, Type: transport.Type, BaseURL: transport.BaseURL,
		APIKeyEnv: transport.APIKeyEnv, APIKeyLiteral: transport.APIKeyLiteral, API: transport.API,
		Proxy: transport.Proxy, HeaderMode: transport.HeaderMode, Headers: transport.Headers,
		CustomHeaders: transport.CustomHeaders, Models: models, Host: transport.Host,
		SelectedModelID: transport.SelectedModelID, Extra: extra,
	}, nil
}

func ConfigTransports(configs []Config) ([]ConfigTransport, error) {
	result := make([]ConfigTransport, 0, len(configs))
	for _, cfg := range configs {
		converted, err := NewConfigTransport(cfg)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
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
	if input.APIKeyLiteral != "" {
		input.APIKeyEnv = ""
	}
	input.API = strings.TrimSpace(input.API)
	input.HeaderMode = strings.TrimSpace(input.HeaderMode)
	input.Models = NormalizeModels(input.Models)
	if input.Type == "" {
		input.Type = "openai-compatible"
	}
	if input.Headers == nil {
		input.Headers = map[string]string{}
	}
	if input.CustomHeaders == nil {
		input.CustomHeaders = map[string]string{}
	}
	if input.HeaderMode == "auto" {
		input.HeaderMode = "none"
		input.Headers = map[string]string{}
	}
	if input.HeaderMode == "" {
		input.HeaderMode = "none"
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
