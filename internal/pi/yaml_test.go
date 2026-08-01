package pi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"piswitch/internal/provider"
)

func TestWriteAllModelsOMP(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.yml")

	cfg := provider.Config{
		ID:        "anthropic",
		BaseURL:   "https://api.anthropic.com",
		API:       "anthropic-messages",
		APIKeyEnv: "ANTHROPIC_API_KEY",
		Headers:   map[string]string{"x-custom": "abc", "user-agent": "pi-switch/1.0"},
		Models: []provider.ModelInfo{
			{ID: "claude-opus-4-6", Name: "Claude Opus 4.6", Reasoning: true, ContextWindow: 1000000, MaxTokens: 128000},
			{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"},
		},
	}

	if err := WriteAllModelsOMP(path, []provider.Config{cfg}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{
		"providers:",
		"  anthropic:",
		"    api: anthropic-messages",
		"    apiKey: ANTHROPIC_API_KEY",
		"    baseUrl: https://api.anthropic.com",
		"    headers:", // 自定义请求头必须写入
		"      user-agent: pi-switch/1.0",
		"      x-custom: abc",
		"      - id: claude-opus-4-6",
		"        contextWindow: 1000000",
		"        maxTokens: 128000",
		"        reasoning: true",
		"      - id: claude-haiku-4-5",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("内容缺少 %q，实际:\n%s", want, content)
		}
	}
	// omp 语义：apiKey 不带 $ 前缀
	if strings.Contains(content, "$ANTHROPIC_API_KEY") {
		t.Errorf("apiKey 不应含 $ 前缀:\n%s", content)
	}

	// 生成物必须能被 YAML 解析
	var parsed yaml.Node
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("生成的 models.yml 无法解析: %v", err)
	}
}

func TestWriteAllModelsOMPLiteralKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.yml")
	cfg := provider.Config{
		ID:            "local",
		BaseURL:       "http://127.0.0.1:1234/v1",
		API:           "openai-completions",
		APIKeyLiteral: "sk-literal",
		Models:        []provider.ModelInfo{{ID: "m1"}},
	}
	if err := WriteAllModelsOMP(path, []provider.Config{cfg}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "apiKey: sk-literal") {
		t.Errorf("字面量密钥未写入:\n%s", data)
	}
}

// TestWriteAllModelsOMPMergePreserves 验证增量合并：用户手写的 provider、
// equivalence 顶层键与注释必须保留，Pi Switch 管理的 provider 被更新。
func TestWriteAllModelsOMPMergePreserves(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.yml")

	existing := `# managed by omp user
providers:
  my-custom:
    baseUrl: https://custom.example.com/v1
    api: openai-completions
    apiKey: MY_CUSTOM_KEY
    models:
      - id: my-model-1
    compat:
      request: "{}"
  anthropic:
    baseUrl: https://old.example.com
    api: anthropic-messages
    apiKey: OLD_KEY
equivalence:
  overrides:
    my-custom/my-model-1: anthropic/claude-opus-4-6
`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := provider.Config{
		ID:        "anthropic",
		BaseURL:   "https://api.anthropic.com",
		API:       "anthropic-messages",
		APIKeyEnv: "ANTHROPIC_API_KEY",
		Models:    []provider.ModelInfo{{ID: "claude-opus-4-6", Name: "Claude Opus 4.6"}},
	}
	if err := WriteAllModelsOMP(path, []provider.Config{cfg}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// 用户手写的 provider 保留
	if !strings.Contains(content, "my-custom:") || !strings.Contains(content, "my-model-1") {
		t.Errorf("用户手写 provider 被覆盖:\n%s", content)
	}
	// 用户 provider 的非管理键（compat）保留
	if !strings.Contains(content, "compat:") {
		t.Errorf("用户 provider 的 compat 键丢失:\n%s", content)
	}
	// equivalence 顶层键保留
	if !strings.Contains(content, "equivalence:") || !strings.Contains(content, "my-custom/my-model-1: anthropic/claude-opus-4-6") {
		t.Errorf("equivalence 配置丢失:\n%s", content)
	}
	// 注释保留
	if !strings.Contains(content, "# managed by omp user") {
		t.Errorf("文件头注释丢失:\n%s", content)
	}
	// Pi Switch 管理的 provider 被更新（baseUrl/apiKey 更新、旧 key 移除）
	if !strings.Contains(content, "baseUrl: https://api.anthropic.com") {
		t.Errorf("anthropic baseUrl 未更新:\n%s", content)
	}
	if strings.Contains(content, "OLD_KEY") {
		t.Errorf("旧 apiKey 残留:\n%s", content)
	}
}

// TestWriteAllModelsOMPNewFile 验证无现有文件时创建完整结构。
func TestWriteAllModelsOMPNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.yml")
	cfg := provider.Config{
		ID:        "deepseek",
		BaseURL:   "https://api.deepseek.com/v1",
		API:       "openai-completions",
		APIKeyEnv: "DEEPSEEK_API_KEY",
		Models:    []provider.ModelInfo{{ID: "Deepseek-v4-pro"}},
	}
	if err := WriteAllModelsOMP(path, []provider.Config{cfg}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	for _, want := range []string{"providers:", "  deepseek:", "api: openai-completions", "apiKey: DEEPSEEK_API_KEY", "- id: Deepseek-v4-pro"} {
		if !strings.Contains(content, want) {
			t.Errorf("缺少 %q:\n%s", want, content)
		}
	}
}
