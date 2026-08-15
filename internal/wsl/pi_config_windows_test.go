//go:build windows

package wsl

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestReadPiConfigFileCommand(t *testing.T) {
	original := piConfigCommand
	defer func() { piConfigCommand = original }()

	tests := []struct {
		name      string
		distro    string
		linuxPath string
	}{
		{
			name:      "simple path",
			distro:    "Ubuntu",
			linuxPath: "/home/alice/.pi/agent/settings.json",
		},
		{
			name:      "spaces stay independent argv values",
			distro:    "Ubuntu 24.04 LTS",
			linuxPath: "/home/first last/.pi/agent/models.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gotName string
			var gotArgs []string
			var gotCommand *exec.Cmd
			piConfigCommand = func(name string, args ...string) *exec.Cmd {
				gotName = name
				gotArgs = append([]string(nil), args...)
				gotCommand = exec.Command("cmd", "/c", "exit", "0")
				return gotCommand
			}

			if _, err := readPiConfigFile(test.distro, test.linuxPath); err != nil {
				t.Fatal(err)
			}

			wantArgs := []string{"--distribution", test.distro, "--exec", "cat", test.linuxPath}
			if gotName != "wsl.exe" || !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Errorf("readPiConfigFile() command = %q %#v, want %q %#v", gotName, gotArgs, "wsl.exe", wantArgs)
			}
			if gotCommand.SysProcAttr == nil || gotCommand.SysProcAttr.CreationFlags != piConfigCreateNoWindow {
				t.Errorf("readPiConfigFile() CreationFlags = %#v, want %#v", gotCommand.SysProcAttr, piConfigCreateNoWindow)
			}
		})
	}
}
