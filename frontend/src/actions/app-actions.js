import { currentProvider } from "./provider-actions.js";

export function createAppActions({ root, api, store, feedback }) {
  function readSettingsForm() {
    return {
      piCommand: root.querySelector('input[name="piCommand"]').value.trim(),
      piSettingsPath: root.querySelector('input[name="piSettingsPath"]').value.trim(),
      piModelsPath: root.querySelector('input[name="piModelsPath"]').value.trim(),
      piSwitchConfigPath: root.querySelector('input[name="piSwitchConfigPath"]').value.trim(),
      darkMode: root.querySelector('input[name="darkMode"]')?.checked ?? false,
      workingDir: root.querySelector('input[name="workingDir"]')?.value?.trim() || ""
    };
  }

  async function setDefault() {
    const provider = currentProvider(store.getState());
    if (!provider) return;

    try {
      await api.setDefaultModel(provider.id, provider.selectedModelId);
      store.setState((state) => ({
        ...state,
        defaultProviderId: provider.id,
        defaultModelId: provider.selectedModelId
      }));
    } catch (error) {
      feedback.showError("设置默认模型失败", error);
    }
  }

  async function directLaunch() {
    const provider = currentProvider(store.getState());
    if (!provider) return;

    try {
      await api.executeLaunchPi(provider.id, provider.selectedModelId);
    } catch (error) {
      feedback.showError("启动终端失败", error);
    }
  }

  async function saveSettings(options = {}) {
    const nextSettings = readSettingsForm();
    const { closeModal = true, immediateTheme = false } = options;
    const previousState = store.getState();
    const previousSettings = previousState.settings;

    try {
      if (immediateTheme) {
        store.setState((state) => ({
          ...state,
          settings: nextSettings
        }));
      }
      await api.updateSettings(nextSettings);
      store.setState((state) => ({
        ...state,
        settings: nextSettings,
        modal: closeModal ? null : state.modal
      }));
    } catch (error) {
      if (immediateTheme) {
        store.setState((state) => ({
          ...state,
          settings: previousSettings
        }));
      }
      feedback.showError("保存应用设置失败", error);
    }
  }

  return { setDefault, directLaunch, saveSettings };
}
