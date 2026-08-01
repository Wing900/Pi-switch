import { createProviderFromPreset } from "../config/presets.js";
import { customHeadersForProvider, headersForMode } from "../config/header-presets.js";
import { withTimeout } from "../utils/async.js";

const FETCH_MODELS_TIMEOUT_MS = 10_000;

export function currentProvider(state) {
  return state.providers.find((provider) => provider.id === state.selectedProviderId) ?? state.providers[0];
}

function getModelCheckboxes(root) {
  return Array.from(root.querySelectorAll("[data-model-id]"));
}

function syncToggleAllButton(root) {
  const button = root.querySelector("[data-toggle-model-selection-all]");
  if (!button) return;

  const checkboxes = getModelCheckboxes(root);
  const allSelected = checkboxes.length > 0 && checkboxes.every((checkbox) => checkbox.checked);
  button.textContent = allSelected ? "全不选" : "全选";
}

function mergeModels(existing, incoming) {
  const merged = [];
  const indexById = new Map();

  for (const model of [...(existing ?? []), ...(incoming ?? [])]) {
    if (!model?.id) continue;
    if (indexById.has(model.id)) {
      merged[indexById.get(model.id)] = {
        ...merged[indexById.get(model.id)],
        ...model,
        name: model.name?.trim?.() || model.id
      };
      continue;
    }
    indexById.set(model.id, merged.length);
    merged.push({
      ...model,
      name: model.name?.trim?.() || model.id
    });
  }

  return merged;
}

export function createProviderActions({ root, api, store, providerForm, feedback }) {
  let headerSaveTimer = null;

  function refreshHeaderModeUI(mode) {
    root.querySelectorAll("[data-set-header-mode]").forEach((button) => {
      button.classList.toggle("is-active", button.dataset.setHeaderMode === mode);
    });
    const editButton = root.querySelector("[data-open-headers-editor]");
    editButton?.classList.toggle("is-hidden", mode !== "custom");
  }

  async function persistProvider(nextProvider) {
    await api.updateProvider(nextProvider.id, nextProvider);
    store.setState((state) => ({
      ...state,
      providers: state.providers.map((item) => (item.id === nextProvider.id ? nextProvider : item)),
      drawer: { kind: "provider", providerId: nextProvider.id }
    }), { notify: false });
    refreshHeaderModeUI(nextProvider.headerMode);
  }

  async function setHeaderMode(mode) {
    if (!(await providerForm.commit())) return;
    const provider = currentProvider(store.getState());
    if (!provider) return;
    const headers = headersForMode(mode, provider.api, customHeadersForProvider(provider));
    try {
      await persistProvider({ ...provider, headerMode: mode, headers });
      if (mode === "custom") {
        store.setState((state) => ({
          ...state,
          modal: { kind: "provider-headers", payload: { providerId: provider.id, headers } }
        }));
      }
    } catch (error) {
      feedback.showError("保存请求头失败", error);
    }
  }

  async function openHeadersEditor() {
    if (!(await providerForm.commit())) return;
    const provider = currentProvider(store.getState());
    if (!provider) return;
    store.setState((state) => ({
      ...state,
      modal: { kind: "provider-headers", payload: { providerId: provider.id, headers: customHeadersForProvider(provider) } }
    }));
  }

  function scheduleProviderHeadersSave() {
    if (headerSaveTimer) window.clearTimeout(headerSaveTimer);
    headerSaveTimer = window.setTimeout(() => {
      headerSaveTimer = null;
      void saveProviderHeaders();
    }, 500);
  }

  async function saveProviderHeaders() {
    const modal = store.getState().modal;
    if (!modal || modal.kind !== "provider-headers") return;
    const names = Array.from(root.querySelectorAll("[data-header-name]"));
    const values = Array.from(root.querySelectorAll("[data-header-value]"));
    const headers = {};
    for (let index = 0; index < names.length; index += 1) {
      const name = names[index]?.value.trim();
      const value = values[index]?.value ?? "";
      if (!name && !value.trim()) continue;
      if (!name || !value.trim()) continue;
      headers[name] = value;
    }
    const provider = store.getState().providers.find((item) => item.id === modal.payload.providerId);
    if (!provider) return;
    if (Object.keys(headers).length === 0
      && Object.keys(provider.headers ?? {}).length === 0
      && Object.keys(provider.customHeaders ?? {}).length === 0) {
      return;
    }
    try {
      await persistProvider({ ...provider, headerMode: "custom", headers, customHeaders: headers });
    } catch (error) {
      feedback.showError("保存请求头失败", error);
    }
  }

  async function createFromPreset(presetId) {
    const nextProvider = createProviderFromPreset(presetId);
    const exists = store.getState().providers.some((provider) => provider.id === nextProvider.id);
    const providerToUse = exists
      ? { ...nextProvider, id: `${nextProvider.id}-${Date.now().toString().slice(-4)}` }
      : nextProvider;

    try {
      await api.createProvider(providerToUse);
      store.setState((state) => ({
        ...state,
        providers: [...state.providers, providerToUse],
        selectedProviderId: providerToUse.id,
        modal: null,
        drawer: { kind: "provider", providerId: providerToUse.id }
      }));
    } catch (error) {
      feedback.showError("创建提供商失败", error);
    }
  }

  async function remove(providerId) {
    const provider = store.getState().providers.find((item) => item.id === providerId);
    if (!provider) return;

    try {
      await api.deleteProvider(provider.id);
      store.setState((state) => {
        const rest = state.providers.filter((item) => item.id !== provider.id);
        const fallback = rest[0];
        return {
          ...state,
          providers: rest,
          selectedProviderId: fallback?.id ?? "",
          defaultProviderId: state.defaultProviderId === provider.id ? fallback?.id ?? "" : state.defaultProviderId,
          defaultModelId:
            state.defaultProviderId === provider.id ? fallback?.selectedModelId ?? "" : state.defaultModelId,
          drawer: null,
          modal: null
        };
      });
    } catch (error) {
      feedback.showError("删除提供商失败", error);
    }
  }

  async function fetchModels() {
    if (!(await providerForm.commit())) return;
    const provider = currentProvider(store.getState());
    if (!provider) return;

    feedback.showLoading("正在获取模型", `正在连接 ${provider.name}`);
    try {
      const fetched = await withTimeout(
        api.fetchModels(provider.id),
        FETCH_MODELS_TIMEOUT_MS,
        "获取模型超时，请检查服务地址或网络连接。"
      );
      store.setState((state) => ({
        ...state,
        modal: {
          kind: "fetch-models",
          payload: {
            providerId: provider.id,
            models: fetched.map((model) => ({ ...model, selected: true }))
          }
        }
      }));
    } catch (error) {
      store.setState((state) => ({
        ...state,
        modal: {
          kind: "operation-result",
          payload: {
            status: "error",
            title: "获取模型失败",
            message: error instanceof Error ? error.message : String(error ?? "未知错误"),
            details: ["请检查配置后重试。"],
            allowManualModel: true,
            providerId: provider.id
          }
        }
      }));
    }
  }

  async function importModels() {
    const modal = store.getState().modal;
    if (!modal || modal.kind !== "fetch-models") return;

    const selected = Array.from(root.querySelectorAll("[data-model-id]:checked"))
      .map((checkbox) => {
        const model = modal.payload.models.find((m) => m.id === checkbox.dataset.modelId);
        if (!model) return null;
        const cwInput = root.querySelector(`[data-cw-model="${model.id}"]`);
        const cwK = parseInt(cwInput?.value, 10) || 256;
        return { ...model, contextWindow: cwK * 1000 };
      })
      .filter(Boolean);

    try {
      await api.replaceModels(modal.payload.providerId, selected);
      store.setState((state) => ({
        ...state,
        providers: state.providers.map((provider) => {
          if (provider.id !== modal.payload.providerId) {
            return provider;
          }
          const nextModels = selected;
          const selectedModelId =
            nextModels.some((model) => model.id === provider.selectedModelId)
              ? provider.selectedModelId
              : nextModels[0]?.id ?? "";
          return { ...provider, models: nextModels, selectedModelId };
        }),
        modal: null
      }));
    } catch (error) {
      feedback.showError("导入模型失败", error);
    }
  }

  function openManualModel() {
    const modal = store.getState().modal;
    const providerId =
      modal?.payload?.providerId || currentProvider(store.getState())?.id || "";
    if (!providerId) return;
    store.setState((state) => ({
      ...state,
      modal: {
        kind: "manual-model",
        payload: {
          providerId,
          modelId: "",
          contextWindowK: 256
        }
      }
    }));
  }

  async function importManualModel() {
    const modal = store.getState().modal;
    if (!modal || modal.kind !== "manual-model") return;

    const modelId = root.querySelector('input[name="manualModelId"]')?.value?.trim();
    const contextWindowK = parseInt(root.querySelector('input[name="manualContextWindow"]')?.value, 10) || 256;
    if (!modelId) {
      feedback.showError("导入模型失败", new Error("Model ID 不能为空"));
      return;
    }

    const manualModel = {
      id: modelId,
      name: modelId,
      contextWindow: contextWindowK * 1000,
      reasoning: false
    };

    try {
      await api.importModels(modal.payload.providerId, [manualModel]);
      store.setState((state) => ({
        ...state,
        providers: state.providers.map((provider) => {
          if (provider.id !== modal.payload.providerId) {
            return provider;
          }
          const mergedModels = mergeModels(provider.models, [manualModel]);
          const selectedModelId =
            provider.selectedModelId?.trim() || mergedModels[0]?.id || manualModel.id;
          return { ...provider, models: mergedModels, selectedModelId };
        }),
        modal: null
      }));
    } catch (error) {
      feedback.showError("导入模型失败", error);
    }
  }

  function toggleAllModelSelections() {
    const checkboxes = getModelCheckboxes(root);
    const nextChecked = checkboxes.some((checkbox) => !checkbox.checked);
    checkboxes.forEach((checkbox) => {
      checkbox.checked = nextChecked;
    });
    syncToggleAllButton(root);
  }

  async function deleteModel(modelId) {
    const provider = currentProvider(store.getState());
    if (!provider || !modelId) return;
    const nextModels = (provider.models ?? []).filter((m) => m.id !== modelId);
    try {
      await api.replaceModels(provider.id, nextModels);
      store.setState((state) => ({
        ...state,
        providers: state.providers.map((item) => {
          if (item.id !== provider.id) return item;
          const selectedModelId =
            nextModels.some((m) => m.id === item.selectedModelId)
              ? item.selectedModelId
              : nextModels[0]?.id ?? "";
          return { ...item, models: nextModels, selectedModelId };
        })
      }));
    } catch (error) {
      feedback.showError("移除模型失败", error);
    }
  }

  return {
    createFromPreset,
    remove,
    fetchModels,
    importModels,
    openManualModel,
    importManualModel,
    toggleAllModelSelections,
    deleteModel,
    setHeaderMode,
    openHeadersEditor,
    scheduleProviderHeadersSave,
    syncToggleAllButton: () => syncToggleAllButton(root)
  };
}
