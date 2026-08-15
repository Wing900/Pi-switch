package configsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"piswitch/internal/config"
	"piswitch/internal/paths"
	"piswitch/internal/pi"
	"piswitch/internal/provider"
)

func TestReplaceModelsRenamesSelectedAndDefaultModel(t *testing.T) {
	tempDir := t.TempDir()
	appPaths := paths.AppPaths{
		PiSwitchConfigPath: filepath.Join(tempDir, "piswitch.json"),
		PiSettingsPath:     filepath.Join(tempDir, "settings.json"),
		PiModelsPath:       filepath.Join(tempDir, "models.json"),
		BackupDir:          filepath.Join(tempDir, "backups"),
	}
	modelsDocument := []byte(`{
  "providers": {
    "p": {
      "baseUrl": "https://example.test/v1",
      "api": "openai-completions",
      "apiKey": "test",
      "models": [{"id":"old-id","name":"Old","vendorOption":true}]
    }
  }
}`)
	settingsDocument := []byte(`{"defaultProvider":"p","defaultModel":"old-id"}`)
	if err := os.WriteFile(appPaths.PiModelsPath, modelsDocument, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(appPaths.PiSettingsPath, settingsDocument, 0o644); err != nil {
		t.Fatal(err)
	}

	service := config.NewService(appPaths)
	initialConfig, err := service.Load()
	if err != nil {
		t.Fatal(err)
	}
	initialConfig.Settings.LastDefaultProviderID = "p"
	initialConfig.Settings.LastDefaultModelID = "old-id"
	oldModel, err := initialConfig.ProviderByID("p")
	if err != nil {
		t.Fatal(err)
	}
	oldRevision := oldModel.Models[0].Revision
	if oldRevision == "" {
		t.Fatal("loaded model has no revision")
	}
	if err := service.Save(initialConfig); err != nil {
		t.Fatal(err)
	}

	coordinator := New(service, nil)
	if _, err := coordinator.ReplaceModels("p", []provider.ModelInfo{{
		ID:              "new-id",
		Name:            "New",
		ReplaceDocument: true,
		OriginalID:      "old-id",
		Revision:        oldRevision,
	}}); err != nil {
		t.Fatal(err)
	}

	defaults, err := pi.ReadDefaults(appPaths.PiSettingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if defaults.DefaultModel != "new-id" {
		t.Fatalf("default model = %q, want new-id", defaults.DefaultModel)
	}

	cfg, err := coordinator.Load()
	if err != nil {
		t.Fatal(err)
	}
	current, err := cfg.ProviderByID("p")
	if err != nil {
		t.Fatal(err)
	}
	if current.SelectedModelID != "new-id" {
		t.Fatalf("selected model = %q, want new-id", current.SelectedModelID)
	}
	if cfg.Settings.LastDefaultModelID != "new-id" {
		t.Fatalf("last default model = %q, want new-id", cfg.Settings.LastDefaultModelID)
	}

	data, err := os.ReadFile(appPaths.PiModelsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("models document is empty")
	}
	encoded := string(data)
	if containsAny(encoded, "replaceDocument", "originalId") {
		t.Fatalf("internal transport metadata leaked into models.json: %s", encoded)
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func TestModelRenamesReadsMetadataWithoutMutatingModels(t *testing.T) {
	models := []provider.ModelInfo{{
		ID:         "new-id",
		OriginalID: "old-id",
	}}

	renames := modelRenames(models)
	if renames["old-id"] != "new-id" {
		t.Fatalf("renames = %#v, want old-id -> new-id", renames)
	}
	if models[0].OriginalID != "old-id" {
		t.Fatal("rename metadata was mutated before the persistence layer used it")
	}
}
