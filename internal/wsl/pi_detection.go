package wsl

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

const piDetectionFieldCount = 6

const piDetectionScript = `home=${HOME-}
pi_found=0
if command -v pi >/dev/null 2>&1; then
pi_found=1
fi
pi_home="$home/.pi"
pi_home_exists=0
settings_exists=0
models_exists=0
[ -d "$pi_home" ] && pi_home_exists=1
[ -f "$pi_home/agent/settings.json" ] && settings_exists=1
[ -f "$pi_home/agent/models.json" ] && models_exists=1
printf '%s\000' "$home"
if [ "$pi_found" = 1 ]; then
command -v pi | head -c -1 || exit 65
fi
printf '\000%s\000%s\000%s\000%s\000' "$pi_home" "$pi_home_exists" "$settings_exists" "$models_exists"
`

type PiDetection struct {
	Distro         string `json:"distro"`
	Home           string `json:"home"`
	PiAvailable    bool   `json:"piAvailable"`
	PiPath         string `json:"piPath"`
	PiHome         string `json:"piHome"`
	PiHomeExists   bool   `json:"piHomeExists"`
	SettingsExists bool   `json:"settingsExists"`
	ModelsExists   bool   `json:"modelsExists"`
}

type piProbe func(string) ([]byte, error)

func detectPi(distro string, detect func() (Detection, error), probe piProbe) (PiDetection, error) {
	if strings.TrimSpace(distro) == "" {
		return PiDetection{}, fmt.Errorf("WSL distribution name is required")
	}

	detection, err := detect()
	if err != nil {
		return PiDetection{}, fmt.Errorf("detect WSL distributions: %w", err)
	}
	if !detection.Detected {
		return PiDetection{}, fmt.Errorf("WSL is unavailable")
	}
	if !containsDistro(detection.Distros, distro) {
		return PiDetection{}, fmt.Errorf("unknown WSL distribution %q", distro)
	}

	data, err := probe(distro)
	if err != nil {
		return PiDetection{}, fmt.Errorf("probe Pi in WSL distribution %q: %w", distro, err)
	}
	detectionResult, err := parsePiDetection(data)
	if err != nil {
		return PiDetection{}, fmt.Errorf("parse WSL Pi detection result: %w", err)
	}
	detectionResult.Distro = distro
	return detectionResult, nil
}

func containsDistro(distros []string, target string) bool {
	for _, distro := range distros {
		if distro == target {
			return true
		}
	}
	return false
}

func parsePiDetection(data []byte) (PiDetection, error) {
	if !utf8.Valid(data) {
		return PiDetection{}, fmt.Errorf("probe output is not valid UTF-8")
	}
	if len(data) == 0 || data[len(data)-1] != 0 {
		return PiDetection{}, fmt.Errorf("probe output must end with NUL")
	}

	fields := bytes.Split(data[:len(data)-1], []byte{0})
	if len(fields) != piDetectionFieldCount {
		return PiDetection{}, fmt.Errorf("probe output has %d fields, want %d", len(fields), piDetectionFieldCount)
	}

	home := string(fields[0])
	piPath := string(fields[1])
	piHome := string(fields[2])
	if !path.IsAbs(home) {
		return PiDetection{}, fmt.Errorf("home path is not absolute")
	}
	if piPath != "" && !path.IsAbs(piPath) {
		return PiDetection{}, fmt.Errorf("Pi path is not absolute")
	}
	if !path.IsAbs(piHome) {
		return PiDetection{}, fmt.Errorf("Pi home path is not absolute")
	}

	piHomeExists, err := parseProbeFlag(fields[3])
	if err != nil {
		return PiDetection{}, fmt.Errorf("parse Pi home state: %w", err)
	}
	settingsExists, err := parseProbeFlag(fields[4])
	if err != nil {
		return PiDetection{}, fmt.Errorf("parse settings state: %w", err)
	}
	modelsExists, err := parseProbeFlag(fields[5])
	if err != nil {
		return PiDetection{}, fmt.Errorf("parse models state: %w", err)
	}

	return PiDetection{
		Home:           home,
		PiAvailable:    piPath != "",
		PiPath:         piPath,
		PiHome:         piHome,
		PiHomeExists:   piHomeExists,
		SettingsExists: settingsExists,
		ModelsExists:   modelsExists,
	}, nil
}

func parseProbeFlag(value []byte) (bool, error) {
	switch string(value) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("invalid flag %q", value)
	}
}
