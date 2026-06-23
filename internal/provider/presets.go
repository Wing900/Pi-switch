package provider

func Presets() []Config {
	return []Config{
		{
			ID:              "deepseek",
			Name:            "DeepSeek",
			Type:            "openai-compatible",
			BaseURL:         "https://api.deepseek.com/v1",
			APIKeyEnv:       "DEEPSEEK_API_KEY",
			API:             "openai-completions",
			Host:            "api.deepseek.com",
			SelectedModelID: "Deepseek-v4-pro",
			Models: []ModelInfo{
				{ID: "Deepseek-v4-pro", Name: "Deepseek-v4-pro"},
			},
		},
		{
			ID:              "openai",
			Name:            "OpenAI",
			Type:            "openai-compatible",
			BaseURL:         "https://api.openai.com/v1",
			APIKeyEnv:       "OPENAI_API_KEY",
			API:             "openai-completions",
			Host:            "api.openai.com",
			SelectedModelID: "GPT-5.5",
			Models: []ModelInfo{
				{ID: "GPT-5.5", Name: "GPT-5.5"},
			},
		},
	}
}
