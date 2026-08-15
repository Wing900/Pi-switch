package wsl

import (
	"errors"
	"strings"
	"unicode/utf16"
)

type Detection struct {
	Detected bool     `json:"detected"`
	Distros  []string `json:"distros"`
}

func parseDistroList(data []byte) ([]string, error) {
	if len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe {
		data = data[2:]
	}
	if len(data)%2 != 0 {
		return nil, errors.New("WSL 发行版列表不是有效的 UTF-16LE 输出")
	}

	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
	}
	if err := validateUTF16(units); err != nil {
		return nil, err
	}

	distros := make([]string, 0)
	for _, line := range strings.Split(string(utf16.Decode(units)), "\n") {
		distro := strings.TrimSpace(strings.Trim(line, "\x00"))
		if distro != "" {
			distros = append(distros, distro)
		}
	}
	return distros, nil
}

func validateUTF16(units []uint16) error {
	for index, unit := range units {
		switch {
		case unit >= 0xd800 && unit <= 0xdbff:
			if index+1 >= len(units) || units[index+1] < 0xdc00 || units[index+1] > 0xdfff {
				return errors.New("WSL 发行版列表包含无效的 UTF-16 代理项")
			}
		case unit >= 0xdc00 && unit <= 0xdfff:
			if index == 0 || units[index-1] < 0xd800 || units[index-1] > 0xdbff {
				return errors.New("WSL 发行版列表包含无效的 UTF-16 代理项")
			}
		}
	}
	return nil
}
