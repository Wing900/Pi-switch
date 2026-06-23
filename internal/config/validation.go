package config

import (
	"errors"
	"strings"

	"piswitch/internal/provider"
)

func ValidateProvider(input provider.Config) error {
	if strings.TrimSpace(input.ID) == "" {
		return errors.New("Provider ID 不能为空")
	}
	if strings.TrimSpace(input.Name) == "" {
		return errors.New("Provider 名称不能为空")
	}
	if strings.TrimSpace(input.BaseURL) == "" {
		return errors.New("Base URL 不能为空")
	}
	if strings.TrimSpace(input.API) == "" {
		return errors.New("API 模式不能为空")
	}
	return nil
}
