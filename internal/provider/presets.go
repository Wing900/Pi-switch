package provider

func Presets() []Config {
	return []Config{
		{
			ID:              "deepseek",
			Name:            "DeepSeek",
			Type:            "openai-compatible",
			BaseURL:         "https://api.deepseek.com/v1",
			APIKeyEnv:       "",
			API:             "openai-completions",
			Host:            "api.deepseek.com",
			SelectedModelID: "Deepseek-v4-pro",
			Models: []ModelInfo{
				{ID: "Deepseek-v4-pro", Name: "Deepseek-v4-pro"},
			},
		},
		{
			ID:              "anthropic",
			Name:            "Anthropic",
			Type:            "anthropic",
			BaseURL:         "https://api.anthropic.com",
			APIKeyEnv:       "",
			API:             "anthropic-messages",
			Host:            "api.anthropic.com",
			SelectedModelID: "claude-opus-4-6",
			Models: []ModelInfo{
				{ID: "claude-opus-4-6", Name: "Claude Opus 4.6", Reasoning: true, ContextWindow: 1000000, MaxTokens: 128000},
				{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Reasoning: true, ContextWindow: 1000000, MaxTokens: 128000},
				{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", Reasoning: false, ContextWindow: 200000, MaxTokens: 64000},
			},
		},
		{
			ID:              "openai",
			Name:            "OpenAI",
			Type:            "openai-compatible",
			BaseURL:         "https://api.openai.com/v1",
			APIKeyEnv:       "",
			API:             "openai-completions",
			Host:            "api.openai.com",
			SelectedModelID: "GPT-5.5",
			Models: []ModelInfo{
				{ID: "GPT-5.5", Name: "GPT-5.5"},
			},
		},
	}
}
