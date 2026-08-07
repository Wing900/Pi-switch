package pi

import (
	"encoding/json"
	"errors"
	"os"
)

type DefaultSettings struct {
	DefaultProvider      string `json:"defaultProvider"`
	DefaultModel         string `json:"defaultModel"`
	DefaultThinkingLevel string `json:"defaultThinkingLevel"`
}

type DefaultPatch struct {
	DefaultProvider *string
	DefaultModel    *string
}

func ReadDefaults(path string) (DefaultSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultSettings{}, nil
		}
		return DefaultSettings{}, err
	}
	var payload DefaultSettings
	if err := json.Unmarshal(data, &payload); err != nil {
		return DefaultSettings{}, err
	}
	return payload, nil
}

// PatchDefaults changes only fields explicitly supplied by the caller.
// defaultThinkingLevel and every unknown setting remain untouched.
func PatchDefaults(path string, patch DefaultPatch) error {
	return mutateJSONDocument(path, func(payload map[string]json.RawMessage) error {
		if patch.DefaultProvider != nil {
			if err := setJSONField(payload, "defaultProvider", *patch.DefaultProvider); err != nil {
				return err
			}
		}
		if patch.DefaultModel != nil {
			if err := setJSONField(payload, "defaultModel", *patch.DefaultModel); err != nil {
				return err
			}
		}
		return nil
	})
}
