package configsync

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"piswitch/internal/config"
	"piswitch/internal/pi"
	"piswitch/internal/provider"
)

// Coordinator serializes every application mutation and routes it to the
// smallest possible file patch. Pi files remain the source of truth; the app
// config stores UI metadata and cached selections.
type Coordinator struct {
	mu         sync.Mutex
	service    *config.Service
	afterWrite func(string)
}

func New(service *config.Service, afterWrite func(string)) *Coordinator {
	return &Coordinator{service: service, afterWrite: afterWrite}
}

func (c *Coordinator) Load() (config.SwitchConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.service.Load()
}

func (c *Coordinator) UpsertProvider(oldID string, input provider.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	input = provider.Normalize(input)
	if err := config.ValidateProvider(input); err != nil {
		return err
	}
	cfg, err := c.service.Load()
	if err != nil {
		return err
	}
	if oldID != "" {
		if _, err := cfg.ProviderByID(oldID); err != nil {
			return err
		}
	}
	if oldID == "" || oldID != input.ID {
		if _, err := cfg.ProviderByID(input.ID); err == nil {
			return errors.New("Provider ID 已存在：" + input.ID)
		}
	}

	var defaults pi.DefaultSettings
	providerRenamed := oldID != "" && oldID != input.ID
	if providerRenamed {
		defaults, err = pi.ReadDefaults(cfg.Settings.PiSettingsPath)
		if err != nil {
			return err
		}
	}

	if err := pi.UpsertProvider(cfg.Settings.PiModelsPath, input, oldID); err != nil {
		return err
	}
	c.wrote(cfg.Settings.PiModelsPath)
	cfg.UpsertProvider(input, oldID)

	if providerRenamed && defaults.DefaultProvider == oldID {
		if err := pi.PatchDefaults(cfg.Settings.PiSettingsPath, pi.DefaultPatch{
			DefaultProvider: stringPointer(input.ID),
		}); err != nil {
			return err
		}
		c.wrote(cfg.Settings.PiSettingsPath)
	}
	if cfg.Settings.LastDefaultProviderID == oldID {
		cfg.Settings.LastDefaultProviderID = input.ID
	}
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) DeleteProvider(providerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, err := c.service.Load()
	if err != nil {
		return err
	}
	if _, err := cfg.ProviderByID(providerID); err != nil {
		return err
	}
	defaults, err := pi.ReadDefaults(cfg.Settings.PiSettingsPath)
	if err != nil {
		return err
	}

	if err := pi.DeleteProvider(cfg.Settings.PiModelsPath, providerID); err != nil {
		return err
	}
	c.wrote(cfg.Settings.PiModelsPath)
	cfg.DeleteProvider(providerID)

	if defaults.DefaultProvider == providerID {
		empty := ""
		if err := pi.PatchDefaults(cfg.Settings.PiSettingsPath, pi.DefaultPatch{
			DefaultProvider: &empty,
			DefaultModel:    &empty,
		}); err != nil {
			return err
		}
		c.wrote(cfg.Settings.PiSettingsPath)
	}
	if cfg.Settings.LastDefaultProviderID == providerID {
		cfg.Settings.LastDefaultProviderID = ""
		cfg.Settings.LastDefaultModelID = ""
	}
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) MergeModels(providerID string, models []provider.ModelInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, current, err := c.providerState(providerID)
	if err != nil {
		return err
	}
	merged, err := pi.MergeModels(cfg.Settings.PiModelsPath, providerID, models)
	if err != nil {
		return err
	}
	c.wrote(cfg.Settings.PiModelsPath)
	current.Models = merged
	cfg.UpsertProvider(provider.Normalize(current), providerID)
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) ReplaceModels(providerID string, models []provider.ModelInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	renames := modelRenames(models)

	cfg, current, err := c.providerState(providerID)
	if err != nil {
		return err
	}
	defaults, err := pi.ReadDefaults(cfg.Settings.PiSettingsPath)
	if err != nil {
		return err
	}

	replaced, err := pi.ReplaceModels(cfg.Settings.PiModelsPath, providerID, models)
	if err != nil {
		return err
	}
	c.wrote(cfg.Settings.PiModelsPath)
	current.Models = replaced
	if renamedID := renames[current.SelectedModelID]; renamedID != "" {
		current.SelectedModelID = renamedID
	}
	current = provider.Normalize(current)
	if !providerHasModel(current, current.SelectedModelID) {
		if len(replaced) > 0 {
			current.SelectedModelID = replaced[0].ID
		} else {
			current.SelectedModelID = ""
		}
	}
	cfg.UpsertProvider(current, providerID)

	defaultModelID := defaults.DefaultModel
	if defaults.DefaultProvider == providerID {
		if renamedID := renames[defaultModelID]; renamedID != "" {
			defaultModelID = renamedID
			if err := pi.PatchDefaults(cfg.Settings.PiSettingsPath, pi.DefaultPatch{DefaultModel: stringPointer(defaultModelID)}); err != nil {
				return err
			}
			c.wrote(cfg.Settings.PiSettingsPath)
		}
	}
	defaultRemoved := defaults.DefaultProvider == providerID && defaultModelID != "" && provider.EnsureModel(current, defaultModelID) != nil
	if defaultRemoved {
		empty := ""
		if err := pi.PatchDefaults(cfg.Settings.PiSettingsPath, pi.DefaultPatch{DefaultModel: &empty}); err != nil {
			return err
		}
		c.wrote(cfg.Settings.PiSettingsPath)
	}
	if cfg.Settings.LastDefaultProviderID == providerID {
		if renamedID := renames[cfg.Settings.LastDefaultModelID]; renamedID != "" {
			cfg.Settings.LastDefaultModelID = renamedID
		}
		if provider.EnsureModel(current, cfg.Settings.LastDefaultModelID) != nil {
			cfg.Settings.LastDefaultModelID = ""
		}
	}
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) SetDefault(providerID string, modelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, current, err := c.providerState(providerID)
	if err != nil {
		return err
	}
	if err := provider.EnsureModel(current, modelID); err != nil {
		return err
	}
	if err := pi.PatchDefaults(cfg.Settings.PiSettingsPath, pi.DefaultPatch{
		DefaultProvider: stringPointer(providerID),
		DefaultModel:    stringPointer(modelID),
	}); err != nil {
		return err
	}
	c.wrote(cfg.Settings.PiSettingsPath)

	current.SelectedModelID = modelID
	cfg.UpsertProvider(current, providerID)
	cfg.Settings.LastDefaultProviderID = providerID
	cfg.Settings.LastDefaultModelID = modelID
	return c.saveAppConfig(cfg)
}

func modelRenames(models []provider.ModelInfo) map[string]string {
	renames := map[string]string{}
	for index := range models {
		value, exists := models[index].ExtraFields[provider.ModelOriginalIDField]
		if !exists {
			if rawValue, rawExists := models[index].Extra[provider.ModelOriginalIDField]; rawExists {
				var oldID string
				if json.Unmarshal(rawValue, &oldID) == nil {
					value = oldID
					exists = true
				}
			}
		}
		if exists {
			oldID, ok := value.(string)
			if ok && oldID != "" && oldID != models[index].ID {
				renames[oldID] = models[index].ID
			}
		}
		delete(models[index].ExtraFields, provider.ModelOriginalIDField)
		delete(models[index].Extra, provider.ModelOriginalIDField)
	}
	return renames
}

func providerHasModel(cfg provider.Config, modelID string) bool {
	if modelID == "" {
		return false
	}
	for _, model := range cfg.Models {
		if model.ID == modelID {
			return true
		}
	}
	return false
}

func (c *Coordinator) UpdateSettings(input config.AppSettings) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, err := c.service.Load()
	if err != nil {
		return err
	}
	next := config.NormalizeSettings(input)
	next.PiSwitchConfigPath = c.service.ConfigPath()
	next.LastDefaultProviderID = cfg.Settings.LastDefaultProviderID
	next.LastDefaultModelID = cfg.Settings.LastDefaultModelID
	next.LastUpdateCheckAtUnix = cfg.Settings.LastUpdateCheckAtUnix
	cfg.Settings = next
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) RecordUpdateCheck() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg, err := c.service.Load()
	if err != nil {
		return err
	}
	cfg.Settings.LastUpdateCheckAtUnix = time.Now().Unix()
	return c.saveAppConfig(cfg)
}

func (c *Coordinator) providerState(providerID string) (config.SwitchConfig, provider.Config, error) {
	cfg, err := c.service.Load()
	if err != nil {
		return config.SwitchConfig{}, provider.Config{}, err
	}
	current, err := cfg.ProviderByID(providerID)
	return cfg, current, err
}

func (c *Coordinator) saveAppConfig(cfg config.SwitchConfig) error {
	if err := c.service.Save(cfg); err != nil {
		return err
	}
	c.wrote(c.service.ConfigPath())
	return nil
}

func (c *Coordinator) wrote(path string) {
	if c.afterWrite != nil {
		c.afterWrite(path)
	}
}

func stringPointer(value string) *string {
	return &value
}
