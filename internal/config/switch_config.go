package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"piswitch/internal/paths"
	"piswitch/internal/provider"
)

type AppSettings struct {
	PiCommand             string `json:"piCommand"`
	PiSettingsPath        string `json:"piSettingsPath"`
	PiModelsPath          string `json:"piModelsPath"`
	PiSwitchConfigPath    string `json:"piSwitchConfigPath"`
	LastDefaultProviderID string `json:"lastDefaultProviderId,omitempty"`
	LastDefaultModelID    string `json:"lastDefaultModelId,omitempty"`
}

type SwitchConfig struct {
	Version   int               `json:"version"`
	Providers []provider.Config `json:"providers"`
	Settings  AppSettings       `json:"settings"`
}

type AppState struct {
	Version            string            `json:"version"`
	Providers          []provider.Config `json:"providers"`
	SelectedProviderID string            `json:"selectedProviderId"`
	DefaultProviderID  string            `json:"defaultProviderId"`
	DefaultModelID     string            `json:"defaultModelId"`
	Settings           AppSettings       `json:"settings"`
	Logs               []string          `json:"logs"`
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
		return defaultConfig(s.paths), nil
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
	if len(cfg.Providers) == 0 {
		cfg.Providers = provider.Presets()
	}
	return cfg, nil
}

func (s *Service) Save(cfg SwitchConfig) error {
	cfg.Settings = NormalizeSettings(cfg.Settings)
	if err := os.MkdirAll(filepath.Dir(cfg.Settings.PiSwitchConfigPath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.Settings.PiSwitchConfigPath, data, 0o644)
}

func defaultConfig(appPaths paths.AppPaths) SwitchConfig {
	return SwitchConfig{
		Version:   1,
		Providers: provider.Presets(),
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
