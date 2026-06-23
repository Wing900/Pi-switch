package paths

import (
	"os"
	"path/filepath"
)

type AppPaths struct {
	PiSwitchConfigPath string
	PiSettingsPath     string
	PiModelsPath       string
	BackupDir          string
}

func DefaultPaths() AppPaths {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return AppPaths{
		PiSwitchConfigPath: filepath.Join(home, ".piswitch", "config.json"),
		PiSettingsPath:     filepath.Join(home, ".pi", "agent", "settings.json"),
		PiModelsPath:       filepath.Join(home, ".pi", "agent", "models.json"),
		BackupDir:          filepath.Join(home, ".piswitch", "backups"),
	}
}
