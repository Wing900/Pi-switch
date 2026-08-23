package wsl

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestPiConfigDocumentLimit(t *testing.T) {
	if piConfigDocumentLimit != 1<<23 {
		t.Fatalf("piConfigDocumentLimit = %d, want %d", piConfigDocumentLimit, 1<<23)
	}
}

func TestPiConfigDocumentsJSON(t *testing.T) {
	data, err := json.Marshal(PiConfigDocuments{Distro: "Ubuntu"})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"distro":"Ubuntu","settingsExists":false,"modelsExists":false,"settingsJson":"","modelsJson":""}`
	if string(data) != want {
		t.Fatalf("json.Marshal(PiConfigDocuments) = %s, want %s", data, want)
	}
}

func TestReadPiConfigDocuments(t *testing.T) {
	const distro = "Ubuntu"
	detection := PiDetection{
		Distro:         distro,
		PiHome:         "/home/alice/.pi",
		SettingsExists: true,
		ModelsExists:   true,
	}
	detectOK := func(string) (PiDetection, error) { return detection, nil }

	t.Run("discovery error skips reads", func(t *testing.T) {
		called := false
		got, err := readPiConfigDocuments(distro, func(string) (PiDetection, error) {
			return PiDetection{}, errors.New("list failed")
		}, func(string, string) (string, error) {
			called = true
			return "ignored", nil
		})
		if err == nil || !strings.Contains(err.Error(), "list failed") {
			t.Fatalf("error = %v, want wrapped discovery error", err)
		}
		if got != (PiConfigDocuments{}) {
			t.Fatalf("result = %#v, want zero value", got)
		}
		if called {
			t.Fatal("file read was attempted after a discovery error")
		}
	})

	t.Run("both missing", func(t *testing.T) {
		called := false
		got, err := readPiConfigDocuments(distro, func(string) (PiDetection, error) {
			return PiDetection{Distro: distro, PiHome: "/home/alice/.pi"}, nil
		}, func(string, string) (string, error) {
			called = true
			return "ignored", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		want := PiConfigDocuments{Distro: distro}
		if got != want {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
		if called {
			t.Fatal("file read was attempted for missing documents")
		}
	})

	t.Run("settings only", func(t *testing.T) {
		var paths []string
		got, err := readPiConfigDocuments(distro, func(string) (PiDetection, error) {
			return PiDetection{Distro: distro, PiHome: detection.PiHome, SettingsExists: true}, nil
		}, func(_, linuxPath string) (string, error) {
			paths = append(paths, linuxPath)
			return `{"theme":"dark"}`, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		want := PiConfigDocuments{
			Distro:         distro,
			SettingsExists: true,
			SettingsJSON:   `{"theme":"dark"}`,
		}
		if got != want {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
		if !reflect.DeepEqual(paths, []string{path.Join(detection.PiHome, "agent", "settings.json")}) {
			t.Fatalf("read paths = %#v, want settings.json only", paths)
		}
	})

	t.Run("models only", func(t *testing.T) {
		var paths []string
		got, err := readPiConfigDocuments(distro, func(string) (PiDetection, error) {
			return PiDetection{Distro: distro, PiHome: detection.PiHome, ModelsExists: true}, nil
		}, func(_, linuxPath string) (string, error) {
			paths = append(paths, linuxPath)
			return "[]", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		want := PiConfigDocuments{
			Distro:       distro,
			ModelsExists: true,
			ModelsJSON:   "[]",
		}
		if got != want {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
		if !reflect.DeepEqual(paths, []string{path.Join(detection.PiHome, "agent", "models.json")}) {
			t.Fatalf("read paths = %#v, want models.json only", paths)
		}
	})

	t.Run("both present including empty and non-JSON", func(t *testing.T) {
		var paths []string
		got, err := readPiConfigDocuments(distro, detectOK, func(_, linuxPath string) (string, error) {
			paths = append(paths, linuxPath)
			if strings.HasSuffix(linuxPath, "settings.json") {
				return "", nil
			}
			return "not json\n  ", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		want := PiConfigDocuments{
			Distro:         distro,
			SettingsExists: true,
			ModelsExists:   true,
			ModelsJSON:     "not json\n  ",
		}
		if got != want {
			t.Fatalf("result = %#v, want %#v", got, want)
		}
		wantPaths := []string{
			path.Join(detection.PiHome, "agent", "settings.json"),
			path.Join(detection.PiHome, "agent", "models.json"),
		}
		if !reflect.DeepEqual(paths, wantPaths) {
			t.Fatalf("read paths = %#v, want %#v", paths, wantPaths)
		}
	})

	t.Run("paths with spaces stay independent argv values", func(t *testing.T) {
		piHome := "/srv/users/first last/.pi"
		var gotDistro, gotPath string
		got, err := readPiConfigDocuments("Ubuntu 24.04 LTS", func(name string) (PiDetection, error) {
			return PiDetection{Distro: name, PiHome: piHome, SettingsExists: true}, nil
		}, func(name, linuxPath string) (string, error) {
			gotDistro = name
			gotPath = linuxPath
			return "{}", nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.Distro != "Ubuntu 24.04 LTS" || got.SettingsJSON != "{}" {
			t.Fatalf("result = %#v, want spaced distro and raw content", got)
		}
		if gotDistro != "Ubuntu 24.04 LTS" {
			t.Fatalf("reader distro = %q, want exact argv value", gotDistro)
		}
		if gotPath != path.Join(piHome, "agent", "settings.json") {
			t.Fatalf("reader path = %q, want path.Join result", gotPath)
		}
		if strings.Contains(gotPath, `"`) || strings.Contains(gotDistro, "wsl.exe") {
			t.Fatal("distro or path looks interpolated into a command string")
		}
	})

	t.Run("read failure returns zero result", func(t *testing.T) {
		got, err := readPiConfigDocuments(distro, detectOK, func(string, string) (string, error) {
			return "partial", errors.New("cat failed")
		})
		if err == nil || !strings.Contains(err.Error(), "cat failed") {
			t.Fatalf("error = %v, want contextual read error", err)
		}
		if got != (PiConfigDocuments{}) {
			t.Fatalf("result = %#v, want zero value", got)
		}
	})

	t.Run("first file failure prevents second read", func(t *testing.T) {
		var paths []string
		got, err := readPiConfigDocuments(distro, detectOK, func(_, linuxPath string) (string, error) {
			paths = append(paths, linuxPath)
			return "partial", errors.New("permission denied")
		})
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Fatalf("error = %v, want first-file read error", err)
		}
		if got != (PiConfigDocuments{}) {
			t.Fatalf("result = %#v, want zero value", got)
		}
		if len(paths) != 1 || paths[0] != path.Join(detection.PiHome, "agent", "settings.json") {
			t.Fatalf("read paths = %#v, want only the first present file", paths)
		}
	})

	t.Run("second file failure discards first content", func(t *testing.T) {
		got, err := readPiConfigDocuments(distro, detectOK, func(_, linuxPath string) (string, error) {
			if strings.HasSuffix(linuxPath, "settings.json") {
				return `{"ok":true}`, nil
			}
			return "partial", errors.New("models read failed")
		})
		if err == nil || !strings.Contains(err.Error(), "models read failed") {
			t.Fatalf("error = %v, want second-file read error", err)
		}
		if got != (PiConfigDocuments{}) {
			t.Fatalf("result = %#v, want zero value", got)
		}
	})
}

func TestReadBounded(t *testing.T) {
	t.Run("under limit", func(t *testing.T) {
		input := []byte("abcdef")
		got, err := readBounded(bytes.NewReader(input), 8)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, input) {
			t.Fatalf("readBounded() = %q, want %q", got, input)
		}
	})

	t.Run("exact limit", func(t *testing.T) {
		input := bytes.Repeat([]byte("a"), 8)
		got, err := readBounded(bytes.NewReader(input), 8)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, input) {
			t.Fatalf("readBounded() = %q, want %q", got, input)
		}
	})

	t.Run("overflow does not slurp the source", func(t *testing.T) {
		const limit int64 = 8
		source := bytes.NewReader(bytes.Repeat([]byte("a"), 1024))
		got, err := readBounded(source, limit)
		if err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("error = %v, want size error", err)
		}
		if got != nil {
			t.Fatalf("data = %q, want nil", got)
		}
		if source.Len() == 0 {
			t.Fatal("readBounded buffered the whole source")
		}
		if unread := source.Len(); unread != 1024-int(limit+1) {
			t.Fatalf("unread bytes = %d, want %d", unread, 1024-int(limit+1))
		}
	})
}

func TestDecodePiConfigDocument(t *testing.T) {
	got, err := decodePiConfigDocument([]byte("{\n  \"ok\": true\n}"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "{\n  \"ok\": true\n}" {
		t.Fatalf("decodePiConfigDocument() = %q, want raw text", got)
	}

	empty, err := decodePiConfigDocument(nil)
	if err != nil || empty != "" {
		t.Fatalf("decodePiConfigDocument(nil) = %q, %v, want empty success", empty, err)
	}

	if _, err := decodePiConfigDocument([]byte{0xff, 0xfe}); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("error = %v, want invalid UTF-8 error", err)
	}
}

func TestConsumePiConfigCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local cat-based document reads are exercised on POSIX hosts")
	}

	writeDoc := func(t *testing.T, name, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("raw contents", func(t *testing.T) {
		path := writeDoc(t, "settings.json", "{\n  \"ok\": true\n}")
		got, err := consumePiConfigCommand(exec.Command("cat", path), 1024)
		if err != nil {
			t.Fatal(err)
		}
		if got != "{\n  \"ok\": true\n}" {
			t.Fatalf("content = %q, want raw file text", got)
		}
	})

	t.Run("empty present file", func(t *testing.T) {
		path := writeDoc(t, "empty.json", "")
		got, err := consumePiConfigCommand(exec.Command("cat", path), 1024)
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Fatalf("content = %q, want empty string", got)
		}
	})

	t.Run("process failure", func(t *testing.T) {
		got, err := consumePiConfigCommand(exec.Command("cat", filepath.Join(t.TempDir(), "missing.json")), 1024)
		if err == nil {
			t.Fatal("error = nil, want process error")
		}
		if got != "" {
			t.Fatalf("content = %q, want empty string", got)
		}
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "binary.json")
		if err := os.WriteFile(path, []byte{0xff, 0xfe, 'a'}, 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := consumePiConfigCommand(exec.Command("cat", path), 1024)
		if err == nil || !strings.Contains(err.Error(), "UTF-8") {
			t.Fatalf("error = %v, want invalid UTF-8 error", err)
		}
		if got != "" {
			t.Fatalf("content = %q, want empty string", got)
		}
	})

	t.Run("overflow prefers size error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "big.json")
		if err := os.WriteFile(path, bytes.Repeat([]byte("a"), 256*1024), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := consumePiConfigCommand(exec.Command("cat", path), 32)
		if err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("error = %v, want size error", err)
		}
		if got != "" {
			t.Fatalf("content = %q, want empty string", got)
		}
	})
}

func TestReadPiConfigDocuments_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Windows implementation requires a WSL installation")
	}

	got, err := ReadPiConfigDocuments("Ubuntu")
	if err == nil || !strings.Contains(err.Error(), "supported only on Windows") {
		t.Fatalf("error = %v, want unsupported error", err)
	}
	if got != (PiConfigDocuments{}) {
		t.Fatalf("result = %#v, want zero value", got)
	}
}
