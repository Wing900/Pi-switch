package pi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"piswitch/internal/provider"
)

func TestModelInfoExposesUnknownFieldsThroughTransportEnvelope(t *testing.T) {
	var model provider.ModelInfo
	if err := json.Unmarshal([]byte(`{
  "id":"demo",
  "name":"Demo",
  "api":"openai-completions",
  "thinkingLevelMap":{"high":"high","max":null},
  "vendorOption":{"fast":true}
}`), &model); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["vendorOption"]; ok {
		t.Fatal("unknown fields must use the Wails transport envelope")
	}
	var extra map[string]json.RawMessage
	if err := json.Unmarshal(fields["extraFields"], &extra); err != nil {
		t.Fatal(err)
	}
	if _, ok := extra["vendorOption"]; !ok {
		t.Fatal("transport envelope lost vendorOption")
	}
	if _, ok := fields["api"]; !ok {
		t.Fatal("typed model api field missing")
	}
	if _, ok := fields["thinkingLevelMap"]; !ok {
		t.Fatal("typed thinkingLevelMap field missing")
	}
}

func TestReplaceModelsUsesSubmittedDocumentForExplicitEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [
        {
          "id": "keep",
          "name": "Old name",
          "vendorOption": {"fast": true}
        }
      ]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReplaceModels(path, "demo", []provider.ModelInfo{{
		ID:   "keep",
		Name: "New name",
		Extra: map[string]json.RawMessage{
			"__piSwitchReplaceDocument": json.RawMessage(`true`),
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Providers map[string]struct {
			Models []map[string]json.RawMessage `json:"models"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	model := document.Providers["demo"].Models[0]
	if _, ok := model["vendorOption"]; ok {
		t.Fatal("explicit editor save must allow removal of unknown fields")
	}
	if _, ok := model["__piSwitchReplaceDocument"]; ok {
		t.Fatal("internal replacement marker must not be persisted")
	}
}

func TestReplaceModelsPreservesUnknownFieldsForRetainedModels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [
        {
          "id": "keep",
          "name": "Old name",
          "vendorOption": {"fast": true},
          "selected": true
        },
        {"id": "remove", "name": "Remove me", "vendorOption": {"keep": false}}
      ]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReplaceModels(path, "demo", []provider.ModelInfo{{
		ID:   "keep",
		Name: "New name",
	}})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Providers map[string]struct {
			Models []map[string]json.RawMessage `json:"models"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	models := document.Providers["demo"].Models
	if len(models) != 1 {
		t.Fatalf("models = %d, want 1", len(models))
	}
	if _, ok := models[0]["vendorOption"]; !ok {
		t.Fatal("retained model lost unknown vendorOption field")
	}
	if _, ok := models[0]["selected"]; ok {
		t.Fatal("UI-only selected field must not be persisted")
	}
	var name string
	if err := json.Unmarshal(models[0]["name"], &name); err != nil {
		t.Fatal(err)
	}
	if name != "New name" {
		t.Fatalf("name = %q, want New name", name)
	}
}
