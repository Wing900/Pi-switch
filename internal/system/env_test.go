package system_test

import (
	"os"
	"testing"

	"piswitch/internal/system"
)

func TestResolveAPIKey_EnvAbsent_LiteralFallback(t *testing.T) {
	os.Unsetenv("TEST_API_KEY")
	key, result := system.ResolveAPIKey("TEST_API_KEY", "test-key-from-ui")
	if key != "test-key-from-ui" || !result.Found {
		t.Errorf("key=%q found=%v, want 'test-key-from-ui', true", key, result.Found)
	}
}

func TestResolveAPIKey_EnvPresent_UsesEnv(t *testing.T) {
	os.Setenv("TEST_API_KEY", "sk-from-env")
	defer os.Unsetenv("TEST_API_KEY")
	key, result := system.ResolveAPIKey("TEST_API_KEY", "test-key-from-ui")
	if key != "sk-from-env" || !result.Found {
		t.Errorf("key=%q found=%v, want 'sk-from-env', true", key, result.Found)
	}
}

func TestResolveAPIKey_BothAbsent_Error(t *testing.T) {
	os.Unsetenv("TEST_API_KEY")
	key, result := system.ResolveAPIKey("TEST_API_KEY", "")
	if key != "" || result.Found {
		t.Errorf("key=%q found=%v, want '', false", key, result.Found)
	}
}

func TestResolveAPIKey_LiteralOnly(t *testing.T) {
	key, result := system.ResolveAPIKey("", "test-key-from-ui")
	if key != "test-key-from-ui" || !result.Found {
		t.Errorf("key=%q found=%v, want 'test-key-from-ui', true", key, result.Found)
	}
}