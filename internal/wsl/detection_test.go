package wsl

import (
	"encoding/json"
	"reflect"
	"runtime"
	"testing"
	"unicode/utf16"
)

func TestParseDistroList(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []string
	}{
		{
			name: "one distro",
			data: utf16LE("Ubuntu\r\n", false),
			want: []string{"Ubuntu"},
		},
		{
			name: "multiple distros",
			data: utf16LE("Ubuntu\r\nDebian\r\n", false),
			want: []string{"Ubuntu", "Debian"},
		},
		{
			name: "BOM whitespace and spaces in name",
			data: utf16LE("\r\n Ubuntu  \r\nUbuntu 24.04 LTS\r\n", true),
			want: []string{"Ubuntu", "Ubuntu 24.04 LTS"},
		},
		{
			name: "supplementary plane Unicode",
			data: utf16LE("Ubuntu \U0001F680\r\n", false),
			want: []string{"Ubuntu \U0001F680"},
		},
		{
			name: "terminal NUL",
			data: utf16LE("Debian\r\n\x00", false),
			want: []string{"Debian"},
		},
		{
			name: "empty output",
			data: []byte{},
			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseDistroList(test.data)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("parseDistroList() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseDistroList_InvalidUTF16(t *testing.T) {
	tests := [][]byte{
		{0x55},
		{0x00, 0xd8},
		{0x00, 0xdc},
	}

	for _, data := range tests {
		if _, err := parseDistroList(data); err == nil {
			t.Errorf("parseDistroList(% x) error = nil, want error", data)
		}
	}
}

func TestDetectionJSONEmptyDistros(t *testing.T) {
	data, err := json.Marshal(Detection{Distros: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"detected":false,"distros":[]}` {
		t.Errorf("json.Marshal(Detection{}) = %s, want empty distros array", data)
	}
}

func TestDetect_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Windows implementation requires a WSL installation")
	}

	detection, err := Detect()
	if err != nil {
		t.Fatal(err)
	}
	if detection.Detected || !reflect.DeepEqual(detection.Distros, []string{}) {
		t.Errorf("Detect() = %#v, want unavailable empty result", detection)
	}
}

func utf16LE(text string, bom bool) []byte {
	data := make([]byte, 0, len(text)*2+2)
	if bom {
		data = append(data, 0xff, 0xfe)
	}
	for _, unit := range utf16.Encode([]rune(text)) {
		data = append(data, byte(unit), byte(unit>>8))
	}
	return data
}
