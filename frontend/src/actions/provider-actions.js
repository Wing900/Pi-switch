import { createProviderFromPreset } from "../config/presets.js";
import { customHeadersForProvider, headersForMode } from "../config/header-presets.js";
import { readModelDraft } from "../config/model-editor.js";
import { withTimeout } from "../utils/async.js";

const FETCH_MODELS_TIMEOUT_MS = 10_000;

export function currentProvider(state) {
  return state.providers.find((provider) => provider.id === state.selectedProviderId) ?? state.providers[0];
}

function getModelCheckboxes(root) {
  return Array.from(root.querySelectorAll("[data-model-id]"));
}

function definedFields(value) {
  return Object.fromEntries(Object.entries(value ?? {}).filter(([, fieldValue]) => fieldValue !== undefined));
}

function modelContextInput(root, modelId) {
  return Array.from(root.querySelectorAll("[data-cw-model]"))
    .find((input) => input.dataset.cwModel === modelId);
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

function canonicalValue(value) {
  if (Array.isArray(value)) return value.map(canonicalValue);
  if (value && typeof value === "object") {
    return Object.fromEntries(Object.keys(value).sort().map((key) => [key, canonicalValue(value[key])]));
  }
  return value;
}

export function modelsMatch(left, right) {
  return JSON.stringify(canonicalValue(left ?? null)) === JSON.stringify(canonicalValue(right ?? null));
}

function replaceModelAtPosition(models, oldId, draft) {
  const next = [...(models ?? [])];
  const index = next.findIndex((model) => model.id === oldId);
  if (index >= 0) {
    next[index] = draft;
    return next.filter((model, modelIndex) => modelIndex === index || model.id !== draft.id);
  }
  return [...next.filter((model) => model.id !== draft.id), draft];
}

function modelListResult(result, fallbackModels, fallbackRevision = "") {
  if (Array.isArray(result)) {
    return { models: result, revision: fallbackRevision };
  }
  return {
    models: Array.isArray(result?.models) ? result.models : fallbackModels,
    revision: result?.revision ?? fallbackRevision
  };
}

function syncDefaultModelState(state, providerId, models, { oldId = "", newId = "" } = {}) {
  if (state.defaultProviderId !== providerId) return state;

  let defaultModelId = state.defaultModelId;
  if (oldId && defaultModelId === oldId) defaultModelId = newId;
  if (defaultModelId && !models.some((model) => model.id === defaultModelId)) {
    defaultModelId = "";
  }

  return { ...state, defaultModelId };
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
      const persistedProvider = await api.createProvider(providerToUse);
      const nextProvider = persistedProvider && typeof persistedProvider === "object"
        ? persistedProvider
        : providerToUse;
      store.setState((state) => ({
        ...state,
        providers: [...state.providers, nextProvider],
        selectedProviderId: nextProvider.id,
        modal: null,
        drawer: { kind: "provider", providerId: nextProvider.id }
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
          defaultProviderId: state.defaultProviderId === provider.id ? "" : state.defaultProviderId,
          defaultModelId: state.defaultProviderId === provider.id ? "" : state.defaultModelId,
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
            modelsRevision: provider.modelsRevision || "",
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

    const provider = store.getState().providers.find((item) => item.id === modal.payload.providerId);
    if (!provider) return;
    const selected = Array.from(root.querySelectorAll("[data-model-id]:checked"))
      .map((checkbox) => {
        const model = modal.payload.models.find((m) => m.id === checkbox.dataset.modelId);
        if (!model) return null;
        const existingModel = provider.models?.find((item) => item.id === model.id) ?? {};
        const cwInput = modelContextInput(root, model.id);
        const cwK = parseInt(cwInput?.value, 10) || 256;
        const { selected: _selected, ...modelDraft } = model;
        const extraFieldsJson = model.extraFieldsJson || existingModel.extraFieldsJson;
        return {
          ...existingModel,
          ...definedFields(modelDraft),
          ...(extraFieldsJson ? { extraFieldsJson } : {}),
          contextWindow: cwK * 1000
        };
      })
      .filter(Boolean);

    try {
      const result = await api.replaceModels(modal.payload.providerId, selected, modal.payload.modelsRevision || "");
      const { models: nextModels, revision } = modelListResult(result, selected, modal.payload.modelsRevision || "");
      store.setState((state) => {
        const nextState = {
          ...state,
          providers: state.providers.map((provider) => {
            if (provider.id !== modal.payload.providerId) return provider;
            const selectedModelId =
              nextModels.some((model) => model.id === provider.selectedModelId)
                ? provider.selectedModelId
                : nextModels[0]?.id ?? "";
            return { ...provider, models: nextModels, modelsRevision: revision, selectedModelId };
          }),
          modal: null
        };
        return syncDefaultModelState(nextState, modal.payload.providerId, nextModels);
      });
    } catch (error) {
      feedback.showError("导入模型失败", error);
    }
  }

  async function openModelEditor(modelId = "") {
    if (!(await providerForm.commit())) return;
    const state = store.getState();
    const provider = currentProvider(state);
    const operation = state.modal;
    const providerId = operation?.payload?.providerId || provider?.id || "";
    if (!providerId) return;

    const model = provider?.id === providerId
      ? provider.models?.find((item) => item.id === modelId)
      : undefined;
    store.setState((nextState) => ({
      ...nextState,
      modal: {
        kind: "model-editor",
        payload: {
          providerId,
          mode: model ? "edit" : "add",
          originalModelId: model?.id || "",
          modelsRevision: provider?.modelsRevision || "",
          model: model || {
            id: "",
            name: "",
            reasoning: false,
            contextWindow: 128000,
            maxTokens: 16384
          }
        }
      }
    }));
  }

  async function saveModelEditor() {
    const modal = store.getState().modal;
    if (!modal || modal.kind !== "model-editor") return;

    const readInput = (name) => root.querySelector(`[name="${name}"]`);
    let draft;
    try {
      draft = readModelDraft({
        originalModel: modal.payload.model || {},
        readValue: (name) => readInput(name)?.value ?? "",
        readChecked: (name) => !!readInput(name)?.checked,
        readCheckedValues: (name) => Array.from(root.querySelectorAll(`input[name="${name}"]:checked`))
          .map((input) => input.value)
      });
    } catch (error) {
      feedback.showError("保存模型失败", error);
      return;
    }

    const provider = store.getState().providers.find((item) => item.id === modal.payload.providerId);
    if (!provider) {
      feedback.showError("保存模型失败", new Error("Provider 已被修改或删除，请重新打开模型编辑器"));
      return;
    }
    const oldId = modal.payload.originalModelId;
    const currentModel = oldId ? provider.models?.find((model) => model.id === oldId) : undefined;
    if (oldId && (!currentModel || !modelsMatch(currentModel, modal.payload.model))) {
      feedback.showError("保存模型失败", new Error("模型已被外部修改，请重新打开编辑器后再保存"));
      return;
    }
    const duplicate = (provider.models ?? []).find((model) => model.id === draft.id && model.id !== oldId);
    if (duplicate) {
      feedback.showError("保存模型失败", new Error(`模型 ID 已存在：${draft.id}`));
      return;
    }
    const nextModels = replaceModelAtPosition(provider.models, oldId, draft);

    try {
      const result = await api.replaceModels(provider.id, nextModels, modal.payload.modelsRevision || "");
      const { models: persistedModels, revision } = modelListResult(
        result,
        nextModels,
        modal.payload.modelsRevision || ""
      );
      const persistedStateModels = persistedModels.map((model) => {
        const {
          replaceDocument: _replaceDocument,
          originalId: _originalId,
          ...stateModel
        } = model;
        return stateModel;
      });
      store.setState((state) => {
        const nextState = {
          ...state,
          providers: state.providers.map((item) => {
            if (item.id !== provider.id) return item;
            const selectedModelId = item.selectedModelId === oldId
              ? draft.id
              : persistedStateModels.some((model) => model.id === item.selectedModelId)
                ? item.selectedModelId
                : persistedStateModels[0]?.id ?? "";
            return { ...item, models: persistedStateModels, modelsRevision: revision, selectedModelId };
          }),
          modal: null
        };
        return syncDefaultModelState(nextState, provider.id, persistedStateModels, { oldId, newId: draft.id });
      });
    } catch (error) {
      feedback.showError("保存模型失败", error);
    }
  }

  async function openManualModel() {
    await openModelEditor();
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
    if (!(await providerForm.commit())) return;
    const provider = currentProvider(store.getState());
    if (!provider || !modelId) return;
    const nextModels = (provider.models ?? []).filter((m) => m.id !== modelId);
    try {
      const result = await api.replaceModels(provider.id, nextModels, provider.modelsRevision || "");
      const { models: stateModels, revision } = modelListResult(result, nextModels, provider.modelsRevision || "");
      store.setState((state) => {
        const nextState = {
          ...state,
          providers: state.providers.map((item) => {
            if (item.id !== provider.id) return item;
            const selectedModelId =
              stateModels.some((m) => m.id === item.selectedModelId)
                ? item.selectedModelId
                : stateModels[0]?.id ?? "";
            return { ...item, models: stateModels, modelsRevision: revision, selectedModelId };
          })
        };
        return syncDefaultModelState(nextState, provider.id, stateModels);
      });
    } catch (error) {
      feedback.showError("移除模型失败", error);
    }
  }

  return {
    createFromPreset,
    remove,
    fetchModels,
    importModels,
    openModelEditor,
    saveModelEditor,
    openManualModel,
    toggleAllModelSelections,
    deleteModel,
    setHeaderMode,
    openHeadersEditor,
    scheduleProviderHeadersSave,
    syncToggleAllButton: () => syncToggleAllButton(root)
  };
}
