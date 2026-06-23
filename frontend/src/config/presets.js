const PRESET_LIST = [
  {
    id: "deepseek",
    label: "DeepSeek",
    name: "DeepSeek",
    baseUrl: "https://api.deepseek.com/v1",
    apiKeyEnv: "DEEPSEEK_API_KEY",
    apiKeyLiteral: "",
    api: "openai-completions",
    type: "openai-compatible",
    host: "api.deepseek.com",
    models: [
      { id: "Deepseek-v4-pro", name: "Deepseek-v4-pro", reasoning: false }
    ]
  },
  {
    id: "openai",
    label: "OpenAI",
    name: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    apiKeyEnv: "OPENAI_API_KEY",
    apiKeyLiteral: "",
    api: "openai-completions",
    type: "openai-compatible",
    host: "api.openai.com",
    models: [
      { id: "GPT-5.5", name: "GPT-5.5", reasoning: false }
    ]
  }
];

export const PRESETS = PRESET_LIST;

export function createProviderFromPreset(presetId) {
  const preset = PRESET_LIST.find((item) => item.id === presetId) ?? PRESET_LIST[0];
  return {
    id: preset.id,
    name: preset.name,
    type: preset.type,
    baseUrl: preset.baseUrl,
    apiKeyEnv: preset.apiKeyEnv,
    apiKeyLiteral: preset.apiKeyLiteral,
    api: preset.api,
    proxy: "",
    headers: {},
    models: preset.models.map((model) => ({ ...model })),
    host: preset.host,
    selectedModelId: preset.models[0]?.id ?? "",
    envStatus: preset.apiKeyEnv ? "unknown" : "local"
  };
}
