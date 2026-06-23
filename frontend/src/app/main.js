import { PRESETS, createProviderFromPreset } from "../config/presets.js";
import { renderApp } from "../components/app-shell.js";
import { WailsApi } from "../services/wails-api.js";
import { createStore } from "../state/store.js";

const root = document.querySelector("#app");
const api = new WailsApi();
let autoSaveTimer = null;

const store = createStore({
  version: "0.0.0.1",
  providers: [],
  selectedProviderId: "",
  defaultProviderId: "",
  defaultModelId: "",
  settings: {
    piCommand: "pi",
    piSettingsPath: "",
    piModelsPath: "",
    piSwitchConfigPath: ""
  },
  presets: PRESETS,
  logs: [],
  modal: null,
  drawer: null,
  drawerSaveStatus: "idle"
});

function currentProvider(state) {
  return state.providers.find((provider) => provider.id === state.selectedProviderId) ?? state.providers[0];
}

function drawerField(name) {
  return root.querySelector(`.drawer [name="${name}"]`);
}

function syncSelectedModel(modelId) {
  store.setState((state) => ({
    ...state,
    providers: state.providers.map((provider) =>
      provider.id === state.selectedProviderId ? { ...provider, selectedModelId: modelId } : provider
    )
  }));
}

async function bootstrap() {
  const data = await api.getAppState();
  store.setState((state) => ({
    ...state,
        version: data.version,
        providers: data.providers,
        selectedProviderId: data.providers[0]?.id ?? "",
    defaultProviderId: data.defaultProviderId,
    defaultModelId: data.defaultModelId,
    settings: data.settings,
        logs: data.logs,
        drawer: null
      }));
}

function saveProviderForm({ log = true } = {}) {
  const state = store.getState();
  const provider = currentProvider(state);
  if (!provider) {
    return;
  }
  const nextProvider = {
    ...provider,
    id: drawerField("id")?.value.trim() ?? provider.id,
    name: drawerField("name")?.value.trim() ?? provider.name,
    baseUrl: drawerField("baseUrl")?.value.trim() ?? provider.baseUrl,
    apiKeyEnv: "",
    apiKeyLiteral: drawerField("apiKey")?.value.trim() ?? provider.apiKeyLiteral,
    api: drawerField("api")?.value.trim() ?? provider.api
  };
  api.updateProvider(provider.id, nextProvider).then(() => {
    store.setState((prev) => ({
      ...prev,
      providers: prev.providers.map((item) => (item.id === provider.id ? nextProvider : item)),
      selectedProviderId: nextProvider.id,
      defaultProviderId: prev.defaultProviderId === provider.id ? nextProvider.id : prev.defaultProviderId,
      logs: log ? [`已保存 Provider：${nextProvider.name}`, ...prev.logs] : prev.logs,
      drawer: { kind: "provider", providerId: nextProvider.id },
      drawerSaveStatus: "saved"
    }));
  });
}

function createProviderFromPresetAction(presetId) {
  const nextProvider = createProviderFromPreset(presetId);
  const exists = store.getState().providers.some((provider) => provider.id === nextProvider.id);
  const providerToUse = exists
    ? { ...nextProvider, id: `${nextProvider.id}-${Date.now().toString().slice(-4)}` }
    : nextProvider;
  api.createProvider(providerToUse).then(() => {
    store.setState((state) => ({
      ...state,
      providers: [...state.providers, providerToUse],
      selectedProviderId: providerToUse.id,
      logs: [`已新增 Provider：${providerToUse.name}`, ...state.logs],
      modal: null,
      drawer: { kind: "provider", providerId: providerToUse.id }
    }));
  });
}

function showErrorModal(state, title, error) {
  const message = error instanceof Error ? error.message : String(error ?? "未知错误");
  store.setState({
    ...state,
    modal: {
      kind: "test-result",
      payload: {
        title,
        lines: ["状态：失败", `问题：${message}`]
      }
    },
    logs: [`${title}：${message}`, ...state.logs]
  });
}

async function removeProviderById(state, providerId) {
  const provider = state.providers.find((item) => item.id === providerId);
  if (!provider) {
    return;
  }

  await api.deleteProvider(provider.id);
  const rest = state.providers.filter((item) => item.id !== provider.id);
  const fallbackProvider = rest[0];

  store.setState({
    ...state,
    providers: rest,
    selectedProviderId: fallbackProvider?.id ?? "",
    defaultProviderId: state.defaultProviderId === provider.id ? fallbackProvider?.id ?? "" : state.defaultProviderId,
    defaultModelId:
      state.defaultProviderId === provider.id
        ? fallbackProvider?.selectedModelId ?? ""
        : state.defaultModelId,
    logs: [`已删除 Provider：${provider.name}`, ...state.logs],
    drawer: state.drawer?.providerId === provider.id ? null : state.drawer
  });
}

function scheduleDrawerAutoSave() {
  const state = store.getState();
  if (!state.drawer || state.drawer.kind !== "provider") {
    return;
  }
  if (autoSaveTimer) {
    window.clearTimeout(autoSaveTimer);
  }
  store.setState({ ...state, drawerSaveStatus: "saving" });
  autoSaveTimer = window.setTimeout(() => {
    saveProviderForm({ log: false });
    autoSaveTimer = null;
  }, 320);
}

function bindEvents() {
  root.addEventListener("click", async (event) => {
    const target = event.target.closest("[data-confirm-delete-provider],[data-provider-id],[data-preset-id],[data-open-settings],[data-open-provider-settings],[data-open-add-provider],[data-test-connection],[data-fetch-models],[data-delete-provider],[data-set-default],[data-launch-pi],[data-close-modal],[data-import-models],[data-save-settings],[data-confirm-launch],[data-close-drawer],[data-window-minimise],[data-window-toggle-maximise],[data-window-close]");
    if (!target) {
      if (event.target.matches("[data-close-modal]")) {
        store.setState((state) => ({ ...state, modal: null }));
      }
      return;
    }

    const state = store.getState();

    if (target.dataset.providerId) {
      store.setState({ ...state, selectedProviderId: target.dataset.providerId });
      return;
    }

    if (target.dataset.presetId) {
      createProviderFromPresetAction(target.dataset.presetId);
      return;
    }

    if (target.dataset.openProviderSettings) {
      store.setState({
        ...state,
        selectedProviderId: target.dataset.openProviderSettings,
        drawer: { kind: "provider", providerId: target.dataset.openProviderSettings }
      });
      return;
    }

    if (target.hasAttribute("data-open-add-provider")) {
      store.setState({ ...state, modal: { kind: "add-provider" } });
      return;
    }

    if (target.hasAttribute("data-close-drawer")) {
      store.setState({ ...state, drawer: null });
      return;
    }

    if (target.hasAttribute("data-open-settings")) {
      store.setState({ ...state, modal: { kind: "settings" } });
      return;
    }

    if (target.dataset.confirmDeleteProvider) {
      await removeProviderById(state, target.dataset.confirmDeleteProvider);
      store.setState((prev) => ({ ...prev, modal: null }));
      return;
    }

    if (target.hasAttribute("data-window-minimise")) {
      api.minimiseWindow();
      return;
    }

    if (target.hasAttribute("data-window-toggle-maximise")) {
      api.toggleMaximiseWindow();
      return;
    }

    if (target.hasAttribute("data-window-close")) {
      api.closeWindow();
      return;
    }

    if (target.hasAttribute("data-test-connection")) {
      const provider = currentProvider(state);
      try {
        const result = await api.testConnection(provider.id);
        store.setState((prev) => ({
          ...prev,
          modal: { kind: "test-result", payload: result },
          logs: [`已测试连接：${provider.name}`, ...prev.logs]
        }));
      } catch (error) {
        showErrorModal(state, "连接测试失败", error);
      }
      return;
    }

    if (target.hasAttribute("data-fetch-models")) {
      const provider = currentProvider(state);
      try {
        const fetched = await api.fetchModels(provider.id);
        store.setState((prev) => ({
          ...prev,
          modal: {
            kind: "fetch-models",
            payload: {
              providerId: provider.id,
              models: fetched.map((model) => ({ ...model, selected: true }))
            }
          },
          logs: [`已获取模型：${provider.name}`, ...prev.logs]
        }));
      } catch (error) {
        showErrorModal(state, "获取可用模型失败", error);
      }
      return;
    }

    if (target.hasAttribute("data-import-models")) {
      const modal = state.modal;
      if (!modal || modal.kind !== "fetch-models") {
        return;
      }
      const selected = Array.from(root.querySelectorAll("[data-model-id]:checked")).map((checkbox) =>
        modal.payload.models.find((model) => model.id === checkbox.dataset.modelId)
      );
      await api.importModels(modal.payload.providerId, selected);
      store.setState((prev) => ({
        ...prev,
        providers: prev.providers.map((provider) =>
          provider.id === modal.payload.providerId
            ? {
                ...provider,
                models: selected,
                selectedModelId: selected[0]?.id ?? provider.selectedModelId
              }
            : provider
        ),
        modal: null,
        logs: [`已导入 ${selected.length} 个模型。`, ...prev.logs]
      }));
      return;
    }

    if (target.hasAttribute("data-save-provider")) {
      saveProviderForm();
      return;
    }

    if (target.hasAttribute("data-delete-provider")) {
      const provider = currentProvider(state);
      if (!provider) {
        return;
      }
      store.setState({
        ...state,
        modal: {
          kind: "confirm-delete",
          payload: { id: provider.id, name: provider.name }
        }
      });
      return;
    }

    if (target.hasAttribute("data-set-default")) {
      const provider = currentProvider(state);
      if (!provider) {
        return;
      }
      await api.setDefaultModel(provider.id, provider.selectedModelId);
      store.setState({
        ...state,
        defaultProviderId: provider.id,
        defaultModelId: provider.selectedModelId,
        logs: [`默认模型已切换为：${provider.selectedModelId}`, ...state.logs]
      });
      return;
    }

    if (target.hasAttribute("data-launch-pi")) {
      const provider = currentProvider(state);
      if (!provider) {
        return;
      }
      const payload = await api.launchPi(provider.id, provider.selectedModelId);
      store.setState((prev) => ({
        ...prev,
        modal: { kind: "launch-preview", payload },
        logs: [`准备启动：${payload.command}`, ...prev.logs]
      }));
      return;
    }

    if (target.hasAttribute("data-confirm-launch")) {
      const provider = currentProvider(state);
      if (!provider) {
        return;
      }
      await api.executeLaunchPi(provider.id, provider.selectedModelId);
      store.setState((prev) => ({
        ...prev,
        modal: null,
        logs: [`已尝试启动终端：${provider.id}/${provider.selectedModelId}`, ...prev.logs]
      }));
      return;
    }

    if (target.hasAttribute("data-save-settings")) {
      const nextSettings = {
        piCommand: root.querySelector('input[name="piCommand"]').value.trim(),
        piSettingsPath: root.querySelector('input[name="piSettingsPath"]').value.trim(),
        piModelsPath: root.querySelector('input[name="piModelsPath"]').value.trim(),
        piSwitchConfigPath: root.querySelector('input[name="piSwitchConfigPath"]').value.trim()
      };
      await api.updateSettings(nextSettings);
      store.setState((prev) => ({
        ...prev,
        settings: nextSettings,
        modal: null,
        logs: ["已保存 Pi 路径设置。", ...prev.logs]
      }));
      return;
    }

    if (target.hasAttribute("data-close-modal")) {
      store.setState({ ...state, modal: null });
    }
  });

  root.addEventListener("change", (event) => {
    const target = event.target;
    if (target.matches("[data-model-select]")) {
      syncSelectedModel(target.value);
      return;
    }
    if (target.closest(".drawer") && target.matches('input[name], select[name]')) {
      scheduleDrawerAutoSave();
    }
  });

  root.addEventListener("input", (event) => {
    const target = event.target;
    if (target.closest(".drawer") && target.matches('input[name]')) {
      scheduleDrawerAutoSave();
    }
  });
}

function render() {
  const state = store.getState();
  root.innerHTML = renderApp(state, PRESETS);
  document.title = `Pi Switch ${state.version}`;
}

store.subscribe(render);
bindEvents();
bootstrap();
