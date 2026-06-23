import { createProviderFromPreset } from "../config/presets.js";
import { withTimeout } from "../utils/async.js";

const FETCH_MODELS_TIMEOUT_MS = 10_000;

export function currentProvider(state) {
  return state.providers.find((provider) => provider.id === state.selectedProviderId) ?? state.providers[0];
}

export function createProviderActions({ root, api, store, providerForm, feedback }) {
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
      feedback.showError("获取模型失败", error);
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
      await api.importModels(modal.payload.providerId, selected);
      store.setState((state) => ({
        ...state,
        providers: state.providers.map((provider) =>
          provider.id === modal.payload.providerId
            ? { ...provider, models: selected, selectedModelId: selected[0]?.id ?? "" }
            : provider
        ),
        modal: null
      }));
    } catch (error) {
      feedback.showError("导入模型失败", error);
    }
  }

  return { createFromPreset, remove, fetchModels, importModels };
}
