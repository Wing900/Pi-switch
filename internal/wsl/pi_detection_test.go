package wsl

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestParsePiDetection(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   PiDetection
	}{
		{
			name:   "complete result",
			fields: []string{"/home/dotmint", "/usr/local/bin/pi", "/home/dotmint/.pi", "1", "1", "1"},
			want: PiDetection{
				Home:           "/home/dotmint",
				PiAvailable:    true,
				PiPath:         "/usr/local/bin/pi",
				PiHome:         "/home/dotmint/.pi",
				PiHomeExists:   true,
				SettingsExists: true,
				ModelsExists:   true,
			},
		},
		{
			name:   "Pi unavailable with configuration directory",
			fields: []string{"/home/dotmint", "", "/home/dotmint/.pi", "1", "0", "1"},
			want: PiDetection{
				Home:         "/home/dotmint",
				PiHome:       "/home/dotmint/.pi",
				PiHomeExists: true,
				ModelsExists: true,
			},
		},
		{
			name:   "root home with absent Pi paths",
			fields: []string{"/root", "", "/root/.pi", "0", "0", "0"},
			want: PiDetection{
				Home:   "/root",
				PiHome: "/root/.pi",
			},
		},
		{
			name:   "paths containing whitespace and newline",
			fields: []string{"/srv/users/first last\nuser", "/opt/pi tools\ncurrent/pi\n", "/srv/users/first last\nuser/.pi", "1", "1", "0"},
			want: PiDetection{
				Home:           "/srv/users/first last\nuser",
				PiAvailable:    true,
				PiPath:         "/opt/pi tools\ncurrent/pi\n",
				PiHome:         "/srv/users/first last\nuser/.pi",
				PiHomeExists:   true,
				SettingsExists: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsePiDetection(probeRecord(test.fields...))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("parsePiDetection() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParsePiDetection_InvalidOutput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid UTF-8",
			data: []byte{0xff, 0},
		},
		{
			name: "missing terminal NUL",
			data: []byte("/home/dotmint\x00\x00/home/dotmint/.pi\x000\x000\x000"),
		},
		{
			name: "wrong field count",
			data: probeRecord("/home/dotmint", "", "/home/dotmint/.pi", "0", "0"),
		},
		{
			name: "invalid flag",
			data: probeRecord("/home/dotmint", "", "/home/dotmint/.pi", "2", "0", "0"),
		},
		{
			name: "empty home",
			data: probeRecord("", "", "/home/dotmint/.pi", "0", "0", "0"),
		},
		{
			name: "relative home",
			data: probeRecord("home/dotmint", "", "/home/dotmint/.pi", "0", "0", "0"),
		},
		{
			name: "relative Pi path",
			data: probeRecord("/home/dotmint", "pi", "/home/dotmint/.pi", "0", "0", "0"),
		},
		{
			name: "relative Pi home",
			data: probeRecord("/home/dotmint", "", ".pi", "0", "0", "0"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parsePiDetection(test.data); err == nil {
				t.Fatal("parsePiDetection() error = nil, want error")
			}
		})
	}
}

func TestDetectPi(t *testing.T) {
	validProbe := func(string) ([]byte, error) {
		return probeRecord("/home/dotmint", "/usr/local/bin/pi", "/home/dotmint/.pi", "1", "1", "1"), nil
	}
	validDetect := func() (Detection, error) {
		return Detection{Detected: true, Distros: []string{"Ubuntu"}}, nil
	}

	t.Run("complete result", func(t *testing.T) {
		got, err := detectPi("Ubuntu", validDetect, validProbe)
		if err != nil {
			t.Fatal(err)
		}
		if got.Distro != "Ubuntu" || !got.PiAvailable || got.PiPath != "/usr/local/bin/pi" {
			t.Errorf("detectPi() = %#v, want complete Ubuntu Pi result", got)
		}
	})

	t.Run("empty distro skips detection", func(t *testing.T) {
		called := false
		_, err := detectPi("  ", func() (Detection, error) {
			called = true
			return Detection{}, nil
		}, validProbe)
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Errorf("detectPi() error = %v, want required distro error", err)
		}
		if called {
			t.Error("Detect() was called for an empty distro")
		}
	})

	t.Run("detection error", func(t *testing.T) {
		_, err := detectPi("Ubuntu", func() (Detection, error) {
			return Detection{}, errors.New("list failed")
		}, validProbe)
		if err == nil || !strings.Contains(err.Error(), "detect WSL distributions") {
			t.Errorf("detectPi() error = %v, want contextual detection error", err)
		}
	})

	t.Run("WSL unavailable skips probe", func(t *testing.T) {
		called := false
		_, err := detectPi("Ubuntu", func() (Detection, error) {
			return Detection{Distros: []string{}}, nil
		}, func(string) ([]byte, error) {
			called = true
			return nil, nil
		})
		if err == nil || !strings.Contains(err.Error(), "unavailable") {
			t.Errorf("detectPi() error = %v, want unavailable WSL error", err)
		}
		if called {
			t.Error("probe was called while WSL was unavailable")
		}
	})

	t.Run("unknown distro skips probe", func(t *testing.T) {
		called := false
		_, err := detectPi("Debian", validDetect, func(string) ([]byte, error) {
			called = true
			return nil, nil
		})
		if err == nil || !strings.Contains(err.Error(), `unknown WSL distribution "Debian"`) {
			t.Errorf("detectPi() error = %v, want unknown distro error", err)
		}
		if called {
			t.Error("probe was called for an unknown distro")
		}
	})

	t.Run("probe error", func(t *testing.T) {
		_, err := detectPi("Ubuntu", validDetect, func(string) ([]byte, error) {
			return nil, errors.New("probe failed")
		})
		if err == nil || !strings.Contains(err.Error(), `probe Pi in WSL distribution "Ubuntu"`) {
			t.Errorf("detectPi() error = %v, want contextual probe error", err)
		}
	})

	t.Run("malformed probe output", func(t *testing.T) {
		_, err := detectPi("Ubuntu", validDetect, func(string) ([]byte, error) {
			return []byte("not a probe record"), nil
		})
		if err == nil || !strings.Contains(err.Error(), "parse WSL Pi detection result") {
			t.Errorf("detectPi() error = %v, want contextual parse error", err)
		}
	})
}

func TestPiDetectionScriptProtocol(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the probe script is exercised with POSIX sh on non-Windows hosts")
	}

	home := t.TempDir()
	binDir := t.TempDir()
	piPath := filepath.Join(binDir, "pi")
	if err := os.WriteFile(piPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	piHome := filepath.Join(home, ".pi")
	if err := os.MkdirAll(filepath.Join(piHome, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(piHome, "agent", "settings.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(piHome, "agent", "models.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("sh", "-c", piDetectionScript)
	cmd.Env = append(os.Environ(), "HOME="+home, "PATH="+binDir+":"+os.Getenv("PATH"))
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	result, err := parsePiDetection(output)
	if err != nil {
		t.Fatal(err)
	}
	if !result.PiAvailable || result.PiPath != piPath || !result.PiHomeExists || !result.SettingsExists || result.ModelsExists {
		t.Errorf("probe result = %#v, want available Pi, directory home, settings file, and models directory excluded", result)
	}
}

func TestDetectPi_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Windows implementation requires a WSL installation")
	}

	if _, err := DetectPi("Ubuntu"); err == nil {
		t.Error("DetectPi() error = nil, want unsupported error")
	}
}

func probeRecord(fields ...string) []byte {
	return []byte(strings.Join(fields, "\x00") + "\x00")
}
