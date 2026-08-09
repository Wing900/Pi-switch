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
	transport, err := provider.NewModelTransport(model)
	if err != nil {
		t.Fatal(err)
	}
	if transport.ExtraFieldsJSON != `{
  "vendorOption": {
    "fast": true
  }
}` {
		t.Fatalf("extraFieldsJson = %s", transport.ExtraFieldsJSON)
	}
	if transport.API != "openai-completions" {
		t.Fatalf("api = %q", transport.API)
	}
	if transport.ThinkingLevelMap["high"] != "high" {
		t.Fatalf("thinkingLevelMap = %#v", transport.ThinkingLevelMap)
	}
}

func TestMergeModelsPreservesMissingKnownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{
        "id": "keep",
        "name": "Old name",
        "contextWindow": 200000,
        "maxTokens": 32000,
        "compat": {"supportsDeveloperRole": false}
      }]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := MergeModels(path, "demo", []provider.ModelInfo{{ID: "keep", Name: "New name"}}); err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	for _, field := range []string{"contextWindow", "maxTokens", "compat"} {
		if _, ok := model[field]; !ok {
			t.Errorf("incremental merge removed existing known field %q", field)
		}
	}
}

func TestMergeModelsPreservesMissingNestedCompatFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{
        "id":"keep",
        "name":"Keep",
        "compat":{"supportsDeveloperRole":false,"supportsTemperature":true,"vendorCompat":1}
      }]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := MergeModels(path, "demo", []provider.ModelInfo{{
		ID:     "keep",
		Name:   "Updated",
		Compat: map[string]any{"supportsDeveloperRole": true},
	}}); err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	var compat map[string]json.RawMessage
	if err := json.Unmarshal(model["compat"], &compat); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"supportsTemperature", "vendorCompat"} {
		if _, ok := compat[field]; !ok {
			t.Errorf("incremental merge removed compat field %q", field)
		}
	}
	if got := string(compat["supportsDeveloperRole"]); got != "true" {
		t.Fatalf("supportsDeveloperRole = %s, want true", got)
	}
}
func TestReplaceModelsPreservesUnknownNestedCompatFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{
        "id":"keep",
        "name":"Keep",
        "compat":{"supportsDeveloperRole":false,"vendorCompat":{"mode":"future"}}
      }]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ReplaceModels(path, "demo", []provider.ModelInfo{{
		ID:     "keep",
		Name:   "Updated",
		Compat: map[string]any{"supportsDeveloperRole": true},
	}}); err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	var compat map[string]json.RawMessage
	if err := json.Unmarshal(model["compat"], &compat); err != nil {
		t.Fatal(err)
	}
	if _, ok := compat["vendorCompat"]; !ok {
		t.Fatal("replace list removed unknown nested compat field")
	}
	if got := string(compat["supportsDeveloperRole"]); got != "true" {
		t.Fatalf("supportsDeveloperRole = %s, want true", got)
	}
}

func TestReplaceModelsPreservesUnknownTopLevelExtraFieldsName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{"id":"keep","name":"Keep","extraFields":{"vendorFlag":true}}]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	providers, err := ReadAllModels(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReplaceModels(path, "demo", providers[0].Models); err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	if _, ok := model["extraFields"]; !ok {
		t.Fatal("unknown top-level extraFields field was not preserved")
	}
	if _, ok := model["vendorFlag"]; ok {
		t.Fatal("nested extraFields value leaked into the model top level")
	}
}

func TestReplaceModelsPreservesUnknownLargeInteger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{"id":"keep","name":"Keep","vendorId":9007199254740993}]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	providers, err := ReadAllModels(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReplaceModels(path, "demo", providers[0].Models); err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	if got := string(model["vendorId"]); got != "9007199254740993" {
		t.Fatalf("vendorId = %s, want exact 9007199254740993", got)
	}
}

func TestReplaceModelsRejectsStaleListRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{"providers":{"demo":{"models":[{"id":"keep","name":"Keep"}]}}}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}
	providers, err := ReadAllModels(path)
	if err != nil {
		t.Fatal(err)
	}
	expectedRevision := provider.ModelListRevision(providers[0].Models)
	if expectedRevision == "" {
		t.Fatal("model list has no revision")
	}

	external := []byte(`{"providers":{"demo":{"models":[{"id":"keep","name":"Keep"},{"id":"external","name":"External"}]}}}`)
	if err := os.WriteFile(path, external, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReplaceModels(path, "demo", providers[0].Models, expectedRevision); err == nil {
		t.Fatal("stale list replacement unexpectedly removed an external model")
	}
	if model := readModelDocument(t, path, "demo", "external"); model == nil {
		t.Fatal("external model was removed")
	}
}
func TestReplaceModelsRejectsStaleRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{
  "providers": {
    "demo": {
      "models": [{"id":"keep","name":"Before","contextWindow":128000}]
    }
  }
}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}

	providers, err := ReadAllModels(path)
	if err != nil {
		t.Fatal(err)
	}
	stale := providers[0].Models[0]
	stale.Name = "Edited in Pi Switch"
	stale.ReplaceDocument = true

	external := []byte(`{
  "providers": {
    "demo": {
      "models": [{"id":"keep","name":"Externally changed","contextWindow":200000}]
    }
  }
}`)
	if err := os.WriteFile(path, external, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ReplaceModels(path, "demo", []provider.ModelInfo{stale}); err == nil {
		t.Fatal("stale model edit unexpectedly overwrote an external change")
	}
	model := readModelDocument(t, path, "demo", "keep")
	var name string
	if err := json.Unmarshal(model["name"], &name); err != nil {
		t.Fatal(err)
	}
	if name != "Externally changed" {
		t.Fatalf("name = %q, want external value", name)
	}
}

func TestReplaceModelsReturnsFreshRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	initial := []byte(`{"providers":{"demo":{"models":[{"id":"keep","name":"Before"}]}}}`)
	if err := os.WriteFile(path, initial, 0o644); err != nil {
		t.Fatal(err)
	}
	providers, err := ReadAllModels(path)
	if err != nil {
		t.Fatal(err)
	}
	model := providers[0].Models[0]
	oldRevision := model.Revision
	model.Name = "After"
	model.ReplaceDocument = true

	replaced, err := ReplaceModels(path, "demo", []provider.ModelInfo{model})
	if err != nil {
		t.Fatal(err)
	}
	if len(replaced) != 1 || replaced[0].Revision == "" || replaced[0].Revision == oldRevision {
		t.Fatalf("revisions = old %q, new %#v", oldRevision, replaced)
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
		ID:              "keep",
		Name:            "New name",
		ReplaceDocument: true,
	}})
	if err != nil {
		t.Fatal(err)
	}

	model := readModelDocument(t, path, "demo", "keep")
	if _, ok := model["vendorOption"]; ok {
		t.Fatal("explicit editor save must allow removal of unknown fields")
	}
	if _, ok := model["replaceDocument"]; ok {
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

func readModelDocument(t *testing.T, path string, providerID string, modelID string) map[string]json.RawMessage {
	t.Helper()
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
	for _, model := range document.Providers[providerID].Models {
		var id string
		if err := json.Unmarshal(model["id"], &id); err == nil && id == modelID {
			return model
		}
	}
	t.Fatalf("model %s/%s not found", providerID, modelID)
	return nil
}
