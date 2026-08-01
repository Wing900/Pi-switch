package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAnthropicModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %q, want /v1/models", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "sk-test" {
			t.Errorf("x-api-key = %q, want sk-test", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want 2023-06-01", r.Header.Get("anthropic-version"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"claude-opus-4-6","display_name":"Claude Opus 4.6"}]}`))
	}))
	defer srv.Close()

	cfg := Config{BaseURL: srv.URL, API: "anthropic-messages"}
	models, err := FetchAnthropicModels(cfg, "sk-test")
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 1 || models[0].ID != "claude-opus-4-6" {
		t.Fatalf("models = %+v, want [claude-opus-4-6]", models)
	}
	if models[0].Name != "Claude Opus 4.6" {
		t.Errorf("Name = %q, want Claude Opus 4.6", models[0].Name)
	}
}

func TestFetchAnthropicModelsBaseURLWithV1(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path = %q, want /v1/models", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	// 用户填了带 /v1 的 BaseURL 时不应重复拼接
	cfg := Config{BaseURL: srv.URL + "/v1", API: "anthropic-messages"}
	if _, err := FetchAnthropicModels(cfg, "k"); err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestFetchModelsByAPIRouting(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"m1"}]}`))
	}))
	defer srv.Close()

	// anthropic-messages 路由到 /v1/models
	if _, err := FetchModelsByAPI(Config{BaseURL: srv.URL, API: "anthropic-messages"}, "k"); err != nil {
		t.Fatal(err)
	}
	// anthropic 别名同样生效
	if _, err := FetchModelsByAPI(Config{BaseURL: srv.URL, API: "anthropic"}, "k"); err != nil {
		t.Fatal(err)
	}
	// openai-completions 路由到 /models
	if _, err := FetchModelsByAPI(Config{BaseURL: srv.URL, API: "openai-completions"}, "k"); err != nil {
		t.Fatal(err)
	}

	want := []string{"/v1/models", "/v1/models", "/models"}
	if len(paths) != len(want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}
