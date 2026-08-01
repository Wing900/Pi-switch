package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeSettingsDefaultPi(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	got := NormalizeSettings(AppSettings{})
	if got.AgentDistribution != "pi" {
		t.Errorf("AgentDistribution = %q, want pi", got.AgentDistribution)
	}
	if got.PiCommand != "pi" {
		t.Errorf("PiCommand = %q, want pi", got.PiCommand)
	}
	if got.PiSettingsPath != filepath.Join(home, ".pi", "agent", "settings.json") {
		t.Errorf("PiSettingsPath = %q", got.PiSettingsPath)
	}
	if got.PiModelsPath != filepath.Join(home, ".pi", "agent", "models.json") {
		t.Errorf("PiModelsPath = %q", got.PiModelsPath)
	}
}

func TestNormalizeSettingsOhMyPi(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	got := NormalizeSettings(AppSettings{AgentDistribution: "ohmypi"})
	if got.PiCommand != "omp" {
		t.Errorf("PiCommand = %q, want omp", got.PiCommand)
	}
	if got.PiSettingsPath != filepath.Join(home, ".omp", "agent", "config.yml") {
		t.Errorf("PiSettingsPath = %q, want %q", got.PiSettingsPath, filepath.Join(home, ".omp", "agent", "config.yml"))
	}
	if got.PiModelsPath != filepath.Join(home, ".omp", "agent", "models.yml") {
		t.Errorf("PiModelsPath = %q, want %q", got.PiModelsPath, filepath.Join(home, ".omp", "agent", "models.yml"))
	}
}

func TestNormalizeSettingsOhMyPiKeepsCustomPaths(t *testing.T) {
	got := NormalizeSettings(AppSettings{
		AgentDistribution: "ohmypi",
		PiCommand:         "omp",
		PiSettingsPath:    "/custom/settings.json",
		PiModelsPath:      "/custom/models.json",
	})
	if got.PiCommand != "omp" {
		t.Errorf("PiCommand = %q, want omp", got.PiCommand)
	}
	if got.PiSettingsPath != "/custom/settings.json" {
		t.Errorf("PiSettingsPath = %q, want custom", got.PiSettingsPath)
	}
}
