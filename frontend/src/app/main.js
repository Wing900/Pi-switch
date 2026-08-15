import { createAppActions } from "../actions/app-actions.js";
import { createOperationFeedback } from "../actions/operation-feedback.js";
import { createProviderActions, currentProvider } from "../actions/provider-actions.js";
import { renderApp, renderContentLayer, renderDrawerLayer } from "../components/app-shell.js";
import { modelManageList } from "../components/provider-drawer.js";
import { renderModal } from "../components/modals.js";
import { PRESETS } from "../config/presets.js";
import { createProviderFormController } from "../controllers/provider-form-controller.js";
import { WailsApi } from "../services/wails-api.js";
import { createStore } from "../state/store.js";
import { transitionState } from "../ui/transitions.js";

const root = document.querySelector("#app");
const api = new WailsApi();
const store = createStore({
  version: "0.0.0.14",
  providers: [],
  selectedProviderId: "",
  defaultProviderId: "",
  defaultModelId: "",
  settings: {
    piCommand: "pi",
    piSettingsPath: "",
    piModelsPath: "",
    darkMode: false
  },
  presets: PRESETS,
  modal: null,
  drawer: null,
  modelMenuOpen: false
});

const feedback = createOperationFeedback(store);
const providerForm = createProviderFormController({
  root,
  api,
  store,
  getCurrentProvider: currentProvider,
  onError: feedback.showError
});
const providerActions = createProviderActions({ root, api, store, providerForm, feedback });
const appActions = createAppActions({ root, api, store, feedback });

function applyDocumentTheme(settings) {
  const theme = settings?.darkMode ? "dark" : "light";
  document.documentElement.dataset.theme = theme;
  document.body.dataset.theme = theme;
  document.documentElement.style.colorScheme = theme;
}

async function bootstrap() {
  try {
    await reloadConfig();
  } catch (error) {
    feedback.showError("应用初始化失败", error);
  }
}

async function reloadConfig() {
  const data = await api.getAppState();
  store.setState((state) => {
    const stillExists = data.providers.some((p) => p.id === state.selectedProviderId);
    const selectedProviderId = stillExists ? state.selectedProviderId : data.selectedProviderId || data.providers[0]?.id || "";
    const drawerProviderStillExists =
      state.drawer && state.drawer.kind === "provider"
        ? data.providers.some((p) => p.id === state.drawer.providerId)
        : true;
    return {
      ...state,
      version: data.version,
      providers: data.providers,
      selectedProviderId,
      defaultProviderId: data.defaultProviderId,
      defaultModelId: data.defaultModelId,
      settings: data.settings,
      drawer: drawerProviderStillExists ? state.drawer : null,
      modelMenuOpen: false
    };
  });
  applyDocumentTheme(data.settings);
  api.applyWindowTheme(!!data.settings?.darkMode);
  return data;
}

function openModal(kind) {
  store.setState((state) => ({ ...state, modal: { kind } }));
}

function closeModal() {
  if (store.getState().modal?.kind === "operation-loading") return;
  store.setState((state) => ({ ...state, modal: null }));
}

async function selectProvider(providerId) {
  if (store.getState().drawer && !(await providerForm.commit())) return;
  transitionState(
    store,
    (state) => ({ ...state, selectedProviderId: providerId, modelMenuOpen: false }),
    "provider"
  );
}

async function openProvider(providerId) {
  if (store.getState().drawer && !(await providerForm.commit())) return;
  store.setState((state) => ({
    ...state,
    selectedProviderId: providerId,
    drawer: { kind: "provider", providerId }
  }));
}

function confirmProviderDeletion() {
  const state = store.getState();
  const provider = currentProvider(state);
  if (!provider) return;
  store.setState({
    ...state,
    modal: { kind: "confirm-delete", payload: { id: provider.id, name: provider.name } }
  });
}

function toggleModelMenu() {
  store.setState((state) => ({ ...state, modelMenuOpen: !state.modelMenuOpen }));
}

function selectModel(modelId) {
  store.setState((state) => ({
    ...state,
    modelMenuOpen: false,
    providers: state.providers.map((provider) =>
      provider.id === state.selectedProviderId
        ? { ...provider, selectedModelId: modelId }
        : provider
    )
  }));
}

async function closeDrawer() {
  if (!(await providerForm.commit())) return;
  store.setState((state) => ({ ...state, drawer: null }));
}

const clickActions = {
  "data-open-add-provider": () => openModal("add-provider"),
  "data-open-settings": () => openModal("settings"),
  "data-toggle-model-menu": toggleModelMenu,
  "data-close-modal": closeModal,
  "data-close-drawer": closeDrawer,
  "data-fetch-models": providerActions.fetchModels,
  "data-open-headers-editor": providerActions.openHeadersEditor,
  "data-import-models": providerActions.importModels,
  "data-toggle-model-selection-all": providerActions.toggleAllModelSelections,
  "data-open-manual-model": providerActions.openManualModel,
  "data-save-model": providerActions.saveModelEditor,
  "data-delete-provider": confirmProviderDeletion,
  "data-set-default": appActions.setDefault,
  "data-launch-pi": appActions.directLaunch,
  "data-window-minimise": () => api.minimiseWindow(),
  "data-window-toggle-maximise": () => api.toggleMaximiseWindow(),
  "data-window-close": () => api.closeWindow(),
  "data-check-update": appActions.checkUpdate,
  "data-skip-update": appActions.skipUpdate,
  "data-install-update": appActions.installUpdate
};

function findActionAttribute(target) {
  return Object.keys(clickActions).find((attribute) => target.hasAttribute(attribute));
}

function bindClickEvents() {
  root.addEventListener("click", async (event) => {
    if (store.getState().modelMenuOpen && !event.target.closest(".model-select")) {
      store.setState((state) => ({ ...state, modelMenuOpen: false }));
    }

    const target = event.target.closest("button, [data-confirm-delete-provider]");
    if (!target) {
      if (event.target.classList.contains("modal-backdrop")) closeModal();
      return;
    }

    if (target.dataset.providerId) return selectProvider(target.dataset.providerId);
    if (target.dataset.selectModel) return selectModel(target.dataset.selectModel);
    if (target.dataset.editModel) return providerActions.openModelEditor(target.dataset.editModel);
    if (target.hasAttribute("data-add-model")) return providerActions.openModelEditor();
    if (target.dataset.deleteModel) return providerActions.deleteModel(target.dataset.deleteModel);
    if (target.dataset.setHeaderMode) return providerActions.setHeaderMode(target.dataset.setHeaderMode);
    if (target.dataset.openProviderSettings) return openProvider(target.dataset.openProviderSettings);
    if (target.dataset.presetId) return providerActions.createFromPreset(target.dataset.presetId);
    if (target.dataset.confirmDeleteProvider) return providerActions.remove(target.dataset.confirmDeleteProvider);

    const attribute = findActionAttribute(target);
    if (attribute) {
      try {
        await clickActions[attribute]();
      } catch (error) {
        feedback.showError("操作失败", error);
      }
    }
  });
}

function bindFormEvents() {
  root.addEventListener("change", async (event) => {
    const target = event.target;
    if (target.matches("[data-model-thinking-mode]")) {
      const level = target.name.replace("modelThinkingMode_", "");
      const valueInput = root.querySelector(`[name="modelThinkingValue_${level}"]`);
      if (valueInput) {
        valueInput.disabled = target.value !== "custom";
        if (target.value !== "custom") valueInput.value = "";
      }
      return;
    }

    if (target.matches("[data-model-id]")) {
      providerActions.syncToggleAllButton();
      return;
    }

    if (target.closest(".drawer") && target.matches("input[name], select[name]")) {
      providerForm.schedule();
      return;
    }

    if (target.closest(".settings-form") && target.matches("input[name], select[name]")) {
      await appActions.saveSettings({
        closeModal: false,
        immediateTheme: target.name === "darkMode"
      });
    }
  });

  root.addEventListener("input", (event) => {
    if (event.target.closest(".drawer") && event.target.matches("input[name]")) {
      providerForm.schedule();
    }
    if (event.target.closest(".headers-editor") && event.target.matches("[data-header-name], [data-header-value]")) {
      providerActions.scheduleProviderHeadersSave();
    }
  });
}

let renderedState = null;

function sameDrawer(previous, next) {
  return previous?.drawer?.kind === next.drawer?.kind
    && previous?.drawer?.providerId === next.drawer?.providerId
    && previous?.selectedProviderId === next.selectedProviderId;
}

function currentProviderSnapshot(state) {
  return state.providers.find((provider) => provider.id === state.selectedProviderId) ?? state.providers[0];
}

function drawerContentChanged(previous, next) {
  if (!sameDrawer(previous, next)) return true;
  const previousProvider = currentProviderSnapshot(previous);
  const nextProvider = currentProviderSnapshot(next);
  return previousProvider?.id !== nextProvider?.id
    || previousProvider?.name !== nextProvider?.name
    || previousProvider?.baseUrl !== nextProvider?.baseUrl
    || previousProvider?.api !== nextProvider?.api
    || previousProvider?.apiKeyLiteral !== nextProvider?.apiKeyLiteral
    || previousProvider?.apiKeyEnv !== nextProvider?.apiKeyEnv;
}

function modelManageChanged(previous, next) {
  const previousProvider = currentProviderSnapshot(previous);
  const nextProvider = currentProviderSnapshot(next);
  return previousProvider?.models !== nextProvider?.models;
}

function updateModelManageLayer(drawerLayer, provider) {
  const current = drawerLayer.querySelector(".model-manage");
  if (!current || !provider) return;
  const holder = document.createElement("div");
  holder.innerHTML = modelManageList(provider);
  const next = holder.firstElementChild;
  if (next) current.replaceWith(next);
}

function contentChanged(previous, next) {
  return previous.selectedProviderId !== next.selectedProviderId
    || previous.defaultProviderId !== next.defaultProviderId
    || previous.defaultModelId !== next.defaultModelId
    || previous.modelMenuOpen !== next.modelMenuOpen
    || previous.settings !== next.settings
    || previous.providers !== next.providers;
}

function renderState(state) {
  if (!renderedState) {
    root.innerHTML = renderApp(state);
    renderedState = state;
    applyDocumentTheme(state.settings);
    document.title = `Pi Switch ${state.version}`;
    return;
  }

  const modalChanged = renderedState.modal !== state.modal
    || (state.modal?.kind === "settings" && renderedState.settings !== state.settings);
  const contentLayer = root.querySelector("[data-content-layer]");
  const drawerLayer = root.querySelector("[data-drawer-layer]");
  const modalLayer = root.querySelector("[data-modal-layer]");
  if (!contentLayer || !drawerLayer || !modalLayer) {
    root.innerHTML = renderApp(state);
  } else {
    if (contentChanged(renderedState, state)) contentLayer.innerHTML = renderContentLayer(state);
    if (drawerContentChanged(renderedState, state)) {
      drawerLayer.innerHTML = renderDrawerLayer(state);
    } else if (modelManageChanged(renderedState, state)) {
      updateModelManageLayer(drawerLayer, currentProviderSnapshot(state));
    }
    if (modalChanged) modalLayer.innerHTML = renderModal(state);
  }

  renderedState = state;
  applyDocumentTheme(state.settings);
  document.title = `Pi Switch ${state.version}`;
}

store.subscribe(renderState);

bindClickEvents();
bindFormEvents();
bootstrap();

api.onConfigChanged(() => {
  reloadConfig().catch((error) => feedback.showError("配置刷新失败", error));
});

api.onUpdateAvailable((result) => {
  store.setState((state) => ({
    ...state,
    modal: { kind: "update-available", payload: result }
  }));
});
