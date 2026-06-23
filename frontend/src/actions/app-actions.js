import { currentProvider } from "./provider-actions.js";

export function createAppActions({ root, api, store, feedback }) {
  async function setDefault() {
    const provider = currentProvider(store.getState());
    if (!provider) return;

    try {
      await api.setDefaultModel(provider.id, provider.selectedModelId);
      store.setState((state) => ({
        ...state,
        defaultProviderId: provider.id,
        defaultModelId: provider.selectedModelId,
        logs: [`默认模型已切换为：${provider.selectedModelId}`, ...state.logs]
      }));
    } catch (error) {
      feedback.showError("设置默认模型失败", error);
    }
  }

  async function prepareLaunch() {
    const provider = currentProvider(store.getState());
    if (!provider) return;

    try {
      const payload = await api.launchPi(provider.id, provider.selectedModelId);
      store.setState((state) => ({
        ...state,
        modal: { kind: "launch-preview", payload },
        logs: [`准备启动：${payload.command}`, ...state.logs]
      }));
    } catch (error) {
      feedback.showError("启动准备失败", error);
    }
  }

  async function confirmLaunch() {
    const provider = currentProvider(store.getState());
    if (!provider) return;

    try {
      await api.executeLaunchPi(provider.id, provider.selectedModelId);
      store.setState((state) => ({
        ...state,
        modal: null,
        logs: [`已启动：${provider.id}/${provider.selectedModelId}`, ...state.logs]
      }));
    } catch (error) {
      feedback.showError("启动终端失败", error);
    }
  }

  async function saveSettings() {
    const nextSettings = {
      piCommand: root.querySelector('input[name="piCommand"]').value.trim(),
      piSettingsPath: root.querySelector('input[name="piSettingsPath"]').value.trim(),
      piModelsPath: root.querySelector('input[name="piModelsPath"]').value.trim(),
      piSwitchConfigPath: root.querySelector('input[name="piSwitchConfigPath"]').value.trim()
    };

    try {
      await api.updateSettings(nextSettings);
      store.setState((state) => ({
        ...state,
        settings: nextSettings,
        modal: null,
        logs: ["已保存 Pi 路径设置。", ...state.logs]
      }));
    } catch (error) {
      feedback.showError("保存应用设置失败", error);
    }
  }

  return { setDefault, prepareLaunch, confirmLaunch, saveSettings };
}
