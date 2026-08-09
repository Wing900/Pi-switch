package provider

import (
	"encoding/json"
	"testing"
)

func TestModelTransportPreservesUnknownRawJSON(t *testing.T) {
	var model ModelInfo
	if err := json.Unmarshal([]byte(`{
  "id":"demo",
  "name":"Demo",
  "reasoning":false,
  "extraFields":{"vendorFlag":true},
  "vendorId":9007199254740993,
  "compat":{
    "supportsDeveloperRole":false,
    "futureCounter":9007199254740993
  }
}`), &model); err != nil {
		t.Fatal(err)
	}

	transport, err := NewModelTransport(model)
	if err != nil {
		t.Fatal(err)
	}
	if transport.ExtraFieldsJSON == "" || transport.CompatExtraFieldsJSON == "" {
		t.Fatalf("transport lost raw fields: %#v", transport)
	}

	roundTripped, err := transport.ModelInfo()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(roundTripped)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if got := string(fields["vendorId"]); got != "9007199254740993" {
		t.Fatalf("vendorId = %s", got)
	}
	if _, ok := fields["extraFields"]; !ok {
		t.Fatal("real top-level extraFields field was not preserved")
	}
	var compat map[string]json.RawMessage
	if err := json.Unmarshal(fields["compat"], &compat); err != nil {
		t.Fatal(err)
	}
	if got := string(compat["futureCounter"]); got != "9007199254740993" {
		t.Fatalf("futureCounter = %s", got)
	}
	if got := string(compat["supportsDeveloperRole"]); got != "false" {
		t.Fatalf("supportsDeveloperRole = %s", got)
	}
}

func TestModelTransportDoesNotInventPersistenceRevision(t *testing.T) {
	transport, err := NewModelTransport(ModelInfo{ID: "fetched", Name: "Fetched"})
	if err != nil {
		t.Fatal(err)
	}
	if transport.Revision != "" {
		t.Fatalf("transient fetched model received persistence revision %q", transport.Revision)
	}
}
func TestModelTransportFiltersSelectedFromUnknownFields(t *testing.T) {
	model := ModelInfo{
		ID:   "demo",
		Name: "Demo",
		Extra: map[string]json.RawMessage{
			"selected":   json.RawMessage(`true`),
			"vendorFlag": json.RawMessage(`1`),
		},
	}
	transport, err := NewModelTransport(model)
	if err != nil {
		t.Fatal(err)
	}
	fields, err := parseRawObject(transport.ExtraFieldsJSON)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["selected"]; ok {
		t.Fatal("UI-only selected field leaked into the transport envelope")
	}
	if _, ok := fields["vendorFlag"]; !ok {
		t.Fatal("real unknown field was removed with selected")
	}
}

func TestModelTransportMetadataDoesNotConsumeSameNamedUnknownFields(t *testing.T) {
	transport := ModelTransport{
		ID:              "demo",
		Name:            "Demo",
		Revision:        "transport-revision",
		ReplaceDocument: true,
		OriginalID:      "old-demo",
		ExtraFieldsJSON: `{
  "revision":"provider-revision",
  "replaceDocument":"provider-value",
  "originalId":"provider-original"
}`,
	}
	model, err := transport.ModelInfo()
	if err != nil {
		t.Fatal(err)
	}
	if model.Revision != "transport-revision" || !model.ReplaceDocument || model.OriginalID != "old-demo" {
		t.Fatalf("transport metadata = %#v", model)
	}
	encoded, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["revision"]) != `"provider-revision"` {
		t.Fatalf("real revision field = %s", fields["revision"])
	}
	if string(fields["replaceDocument"]) != `"provider-value"` {
		t.Fatalf("real replaceDocument field = %s", fields["replaceDocument"])
	}
	if string(fields["originalId"]) != `"provider-original"` {
		t.Fatalf("real originalId field = %s", fields["originalId"])
	}
}

func TestConfigTransportIncludesStableModelListRevision(t *testing.T) {
	var persisted ModelInfo
	if err := json.Unmarshal([]byte(`{"id":"model","name":"Model"}`), &persisted); err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		ID:     "demo",
		Name:   "Demo",
		Models: []ModelInfo{persisted},
	}
	first, err := NewConfigTransport(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewConfigTransport(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if first.ModelsRevision == "" || first.ModelsRevision != second.ModelsRevision {
		t.Fatalf("model revisions = %q and %q", first.ModelsRevision, second.ModelsRevision)
	}
	if len(first.Models) != 1 || first.Models[0].Revision == "" {
		t.Fatalf("model transport revision missing: %#v", first.Models)
	}
}
func TestConfigTransportPreservesProviderUnknownFields(t *testing.T) {
	cfg := Config{
		ID:     "demo",
		Name:   "Demo",
		Models: []ModelInfo{{ID: "model", Name: "Model"}},
		Extra: map[string]json.RawMessage{
			"providerVendorId": json.RawMessage(`9007199254740993`),
		},
	}
	transport, err := NewConfigTransport(cfg)
	if err != nil {
		t.Fatal(err)
	}
	roundTripped, err := transport.Config()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(roundTripped.Extra["providerVendorId"]); got != "9007199254740993" {
		t.Fatalf("providerVendorId = %s", got)
	}
}
