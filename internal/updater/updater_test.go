package updater

import (
	"testing"
	"time"

	"piswitch/internal/config"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.0.0.11", "0.0.0.11", 0},
		{"0.0.0.11", "v0.0.0.11", 0},
		{"0.0.0.10", "0.0.0.11", -1},
		{"0.0.0.11", "0.0.0.10", 1},
		{"0.0.1.0", "0.0.0.11", 1},
		{"1.0.0", "0.0.0.11", 1},
		{"0.0.0.9", "0.0.0.11", -1},
	}
	for _, c := range cases {
		got, err := Compare(c.a, c.b)
		if err != nil {
			t.Fatalf("Compare(%q,%q) error: %v", c.a, c.b, err)
		}
		if got != c.want {
			t.Errorf("Compare(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCompareInvalid(t *testing.T) {
	if _, err := Compare("a.b.c", "0.0.1"); err == nil {
		t.Error("expected error for non-numeric version")
	}
}

func TestNeedsCheck(t *testing.T) {
	// 从未检查过 -> 需要
	if !NeedsCheck(config.AppSettings{}) {
		t.Error("zero value should need check")
	}
	// 刚刚检查过 -> 不需要
	settings := config.AppSettings{LastUpdateCheckAtUnix: time.Now().Unix()}
	if NeedsCheck(settings) {
		t.Error("just-checked should not need check")
	}
	// 8 天前检查过 -> 需要
	settings = config.AppSettings{LastUpdateCheckAtUnix: time.Now().Add(-8 * 24 * time.Hour).Unix()}
	if !NeedsCheck(settings) {
		t.Error("8-day-old check should need recheck")
	}
}
