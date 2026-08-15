package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"piswitch/internal/paths"
	"piswitch/internal/pi"
	"piswitch/internal/provider"
	"piswitch/internal/system"
)

type AppSettings struct {
	PiCommand             string `json:"piCommand"`
	PiSettingsPath        string `json:"piSettingsPath"`
	PiModelsPath          string `json:"piModelsPath"`
	PiSwitchConfigPath    string `json:"piSwitchConfigPath"`
	DarkMode              bool   `json:"darkMode"`
	LastDefaultProviderID string `json:"lastDefaultProviderId,omitempty"`
	LastDefaultModelID    string `json:"lastDefaultModelId,omitempty"`
	WorkingDir            string `json:"workingDir"`
	LastUpdateCheckAtUnix int64  `json:"lastUpdateCheckAt,omitempty"`
}

type SwitchConfig struct {
	Version   int               `json:"version"`
	Providers []provider.Config `json:"providers"`
	Settings  AppSettings       `json:"settings"`
}

type AppState struct {
	Version            string                     `json:"version"`
	Providers          []provider.ConfigTransport `json:"providers"`
	SelectedProviderID string                     `json:"selectedProviderId"`
	DefaultProviderID  string                     `json:"defaultProviderId"`
	DefaultModelID     string                     `json:"defaultModelId"`
	Settings           AppSettings                `json:"settings"`
	Logs               []string                   `json:"logs"`
}

type Service struct {
	paths paths.AppPaths
}

func NewService(appPaths paths.AppPaths) *Service {
	return &Service{paths: appPaths}
}

func (s *Service) Load() (SwitchConfig, error) {
	configPath := s.paths.PiSwitchConfigPath
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		cfg := defaultConfig(s.paths)
		modelsExist, imported, err := readPiProviders(cfg.Settings.PiModelsPath)
		if err != nil {
			return SwitchConfig{}, err
		}
		if modelsExist {
			cfg.Providers = syncPiProviders(cfg.Providers, imported)
		}
		return cfg, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return SwitchConfig{}, err
	}
	var cfg SwitchConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return SwitchConfig{}, err
	}
	cfg.Settings = NormalizeSettings(cfg.Settings)
	cfg.Settings.PiSwitchConfigPath = s.paths.PiSwitchConfigPath

	modelsExist, imported, err := readPiProviders(cfg.Settings.PiModelsPath)
	if err != nil {
		return SwitchConfig{}, err
	}
	if modelsExist {
		cfg.Providers = syncPiProviders(cfg.Providers, imported)
	}
	return cfg, nil
}

func (s *Service) Save(cfg SwitchConfig) error {
	cfg.Settings = NormalizeSettings(cfg.Settings)
	cfg.Settings.PiSwitchConfigPath = s.paths.PiSwitchConfigPath
	configPath := s.paths.PiSwitchConfigPath
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	if err := system.BackupFile(configPath); err != nil {
		return err
	}
	data, err := marshalCompatibleConfig(configPath, cfg)
	if err != nil {
		return err
	}
	return system.WriteFileAtomic(configPath, data, 0o644)
}

func (s *Service) ConfigPath() string {
	return s.paths.PiSwitchConfigPath
}

func marshalCompatibleConfig(path string, cfg SwitchConfig) ([]byte, error) {
	current := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &current); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	delete(current, "deletedProviderIds")

	encoded, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	next := map[string]any{}
	if err := json.Unmarshal(encoded, &next); err != nil {
		return nil, err
	}

	nextProviders := next["providers"]
	delete(next, "providers")
	mergeConfigMap(current, next)
	next["providers"] = nextProviders
	mergeProviderFields(current, next)
	return json.MarshalIndent(current, "", "  ")
}

func mergeConfigMap(target, source map[string]any) {
	for key, value := range source {
		sourceMap, sourceIsMap := value.(map[string]any)
		targetMap, targetIsMap := target[key].(map[string]any)
		if sourceIsMap && targetIsMap {
			mergeConfigMap(targetMap, sourceMap)
			continue
		}
		target[key] = value
	}
}

func mergeProviderFields(target, source map[string]any) {
	targetProviders, _ := target["providers"].([]any)
	sourceProviders, _ := source["providers"].([]any)
	existingByID := make(map[string]map[string]any, len(targetProviders))

	for _, item := range targetProviders {
		providerMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := providerMap["id"].(string); ok {
			existingByID[id] = providerMap
		}
	}

	merged := make([]any, 0, len(sourceProviders))
	for _, item := range sourceProviders {
		providerMap, ok := item.(map[string]any)
		if !ok {
			merged = append(merged, item)
			continue
		}
		id, _ := providerMap["id"].(string)
		if existing := existingByID[id]; existing != nil {
			mergeConfigMap(existing, providerMap)
			merged = append(merged, existing)
			continue
		}
		merged = append(merged, providerMap)
	}
	target["providers"] = merged
}

func defaultConfig(appPaths paths.AppPaths) SwitchConfig {
	return SwitchConfig{
		Version: 1,
		Settings: NormalizeSettings(AppSettings{
			PiCommand:          "pi",
			PiSettingsPath:     appPaths.PiSettingsPath,
			PiModelsPath:       appPaths.PiModelsPath,
			PiSwitchConfigPath: appPaths.PiSwitchConfigPath,
		}),
	}
}

func NormalizeSettings(input AppSettings) AppSettings {
	defaultPaths := paths.DefaultPaths()
	if input.PiCommand == "" {
		input.PiCommand = "pi"
	}
	if input.PiSettingsPath == "" {
		input.PiSettingsPath = defaultPaths.PiSettingsPath
	}
	if input.PiModelsPath == "" {
		input.PiModelsPath = defaultPaths.PiModelsPath
	}
	if input.PiSwitchConfigPath == "" {
		input.PiSwitchConfigPath = defaultPaths.PiSwitchConfigPath
	}
	if input.WorkingDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			input.WorkingDir = home
		}
	}
	return input
}

func (cfg *SwitchConfig) ProviderByID(id string) (provider.Config, error) {
	for _, item := range cfg.Providers {
		if item.ID == id {
			return item, nil
		}
	}
	return provider.Config{}, errors.New("未找到 Provider：" + id)
}

func (cfg *SwitchConfig) UpsertProvider(input provider.Config, oldID string) {
	for index, item := range cfg.Providers {
		if item.ID == oldID || item.ID == input.ID {
			cfg.Providers[index] = input
			return
		}
	}
	cfg.Providers = append(cfg.Providers, input)
}

func (cfg *SwitchConfig) DeleteProvider(id string) {
	next := make([]provider.Config, 0, len(cfg.Providers))
	for _, item := range cfg.Providers {
		if item.ID != id {
			next = append(next, item)
		}
	}
	cfg.Providers = next
}

func syncPiProviders(current []provider.Config, imported []provider.Config) []provider.Config {
	currentByID := make(map[string]provider.Config, len(current))
	for _, item := range current {
		currentByID[item.ID] = item
	}

	merged := make([]provider.Config, 0, len(imported))
	for _, incoming := range imported {
		if currentProvider, exists := currentByID[incoming.ID]; exists {
			currentProvider.BaseURL = incoming.BaseURL
			currentProvider.API = incoming.API
			currentProvider.APIKeyEnv = incoming.APIKeyEnv
			currentProvider.APIKeyLiteral = incoming.APIKeyLiteral
			currentProvider.Headers = incoming.Headers
			currentProvider.Models = incoming.Models
			currentProvider.Host = incoming.Host
			if currentProvider.Type == "" {
				currentProvider.Type = incoming.Type
			}
			if len(incoming.Headers) > 0 && (currentProvider.HeaderMode == "" || currentProvider.HeaderMode == "none") {
				currentProvider.HeaderMode = "custom"
			}
			merged = append(merged, provider.Normalize(currentProvider))
			continue
		}
		merged = append(merged, incoming)
	}

	return merged
}

func readPiProviders(path string) (bool, []provider.Config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}
	providers, err := pi.ReadAllModels(path)
	return true, providers, err
}
