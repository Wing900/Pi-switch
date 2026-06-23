import { createProviderFromPreset } from "../config/presets.js";

function delay(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

export class PiSwitchApi {
  constructor() {
    this.state = {
      version: "0.0.0.1",
      settings: {
        piCommand: "pi",
        piSettingsPath: "~/.pi/agent/settings.json",
        piModelsPath: "~/.pi/agent/models.json",
        piSwitchConfigPath: "~/.piswitch/config.json"
      },
      providers: [
        createProviderFromPreset("deepseek"),
        createProviderFromPreset("ollama"),
        createProviderFromPreset("openrouter")
      ],
      defaultProviderId: "deepseek",
      defaultModelId: "deepseek-chat",
      logs: [
        "就绪。",
        "已载入 3 个 Provider。",
        "当前为原型接口层，后续可直接替换为 Wails 绑定。"
      ]
    };
  }

  async getAppState() {
    await delay(120);
    return clone(this.state);
  }

  async listProviders() {
    await delay(80);
    return clone(this.state.providers);
  }

  async createProvider(input) {
    await delay(120);
    this.state.providers.push(clone(input));
    this.state.logs.unshift(`已新增 Provider：${input.name}`);
  }

  async updateProvider(id, input) {
    await delay(120);
    this.state.providers = this.state.providers.map((provider) =>
      provider.id === id ? { ...clone(input) } : provider
    );
    this.state.logs.unshift(`已保存 Provider：${input.name}`);
  }

  async deleteProvider(id) {
    await delay(120);
    const target = this.state.providers.find((provider) => provider.id === id);
    this.state.providers = this.state.providers.filter((provider) => provider.id !== id);
    if (this.state.defaultProviderId === id) {
      const fallback = this.state.providers[0];
      this.state.defaultProviderId = fallback?.id ?? "";
      this.state.defaultModelId = fallback?.selectedModelId ?? "";
    }
    this.state.logs.unshift(`已删除 Provider：${target?.name ?? id}`);
  }

  async testConnection(id) {
    await delay(360);
    const provider = this.state.providers.find((item) => item.id === id);
    const hasEnv = Boolean(provider?.apiKeyEnv);
    const ok = !hasEnv || provider.apiKeyEnv.includes("API_KEY");
    this.state.logs.unshift(`已测试连接：${provider?.name ?? id}`);
    if (!provider) {
      throw new Error("未找到 Provider。");
    }
    if (!ok) {
      return {
        ok: false,
        title: "连接测试失败",
        lines: [
          "状态：失败",
          `问题：未检测到环境变量 ${provider.apiKeyEnv}`,
          "修复：请在系统环境变量中设置后重新打开 Pi Switch"
        ]
      };
    }
    return {
      ok: true,
      title: "连接测试",
      lines: [
        "状态：正常",
        "Base URL：可访问",
        hasEnv ? `环境变量：${provider.apiKeyEnv}` : "本地模式：使用字面量密钥",
        "/models：可用",
        `检测到模型：${provider.models.length || 2} 个`
      ]
    };
  }

  async fetchModels(id) {
    await delay(420);
    const provider = this.state.providers.find((item) => item.id === id);
    if (!provider) {
      throw new Error("未找到 Provider。");
    }
    const fetched = provider.models.length
      ? provider.models
      : [
          { id: `${provider.id}-chat`, name: `${provider.id}-chat`, reasoning: false },
          { id: `${provider.id}-reasoning`, name: `${provider.id}-reasoning`, reasoning: true }
        ];
    this.state.logs.unshift(`已获取模型：${provider.name}`);
    return clone(fetched);
  }

  async importModels(providerId, models) {
    await delay(180);
    this.state.providers = this.state.providers.map((provider) =>
      provider.id === providerId
        ? {
            ...provider,
            models: clone(models),
            selectedModelId: models[0]?.id ?? provider.selectedModelId
          }
        : provider
    );
    this.state.logs.unshift(`已导入 ${models.length} 个模型。`);
  }

  async setDefaultModel(providerId, modelId) {
    await delay(180);
    this.state.defaultProviderId = providerId;
    this.state.defaultModelId = modelId;
    this.state.logs.unshift(`已设置默认模型：${modelId}`);
  }

  async launchPi(providerId, modelId) {
    await delay(260);
    const command = `pi --provider ${providerId} --model ${modelId}`;
    this.state.logs.unshift(`准备启动：${command}`);
    return {
      command,
      checklist: [
        "Provider 配置存在",
        "models.json 中存在该模型",
        "环境变量检查已通过"
      ]
    };
  }

  async updateSettings(nextSettings) {
    await delay(150);
    this.state.settings = { ...clone(nextSettings) };
    this.state.logs.unshift("已保存 Pi 路径设置。");
  }
}
