//go:build windows

package wsl

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestProbePiCommand(t *testing.T) {
	original := piDetectionCommand
	defer func() { piDetectionCommand = original }()

	var gotName string
	var gotArgs []string
	var gotCommand *exec.Cmd
	piDetectionCommand = func(name string, args ...string) *exec.Cmd {
		gotName = name
		gotArgs = append([]string(nil), args...)
		gotCommand = exec.Command("cmd", "/c", "exit", "0")
		return gotCommand
	}

	if _, err := probePi("Ubuntu"); err != nil {
		t.Fatal(err)
	}

	wantArgs := []string{"--distribution", "Ubuntu", "--exec", "sh", "-c", piDetectionScript}
	if gotName != "wsl.exe" || !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("probePi() command = %q %#v, want %q %#v", gotName, gotArgs, "wsl.exe", wantArgs)
	}
	if gotCommand.SysProcAttr == nil || gotCommand.SysProcAttr.CreationFlags != piDetectionCreateNoWindow {
		t.Errorf("probePi() CreationFlags = %#v, want %#v", gotCommand.SysProcAttr, piDetectionCreateNoWindow)
	}
}
