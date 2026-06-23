import { createAppActions } from "../actions/app-actions.js";
import { createOperationFeedback } from "../actions/operation-feedback.js";
import { createProviderActions, currentProvider } from "../actions/provider-actions.js";
import { renderApp } from "../components/app-shell.js";
import { PRESETS } from "../config/presets.js";
import { createProviderFormController } from "../controllers/provider-form-controller.js";
import { WailsApi } from "../services/wails-api.js";
import { createStore } from "../state/store.js";
import { transitionState } from "../ui/transitions.js";

const root = document.querySelector("#app");
const api = new WailsApi();
const store = createStore({
  version: "0.0.0.2",
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

async function bootstrap() {
  try {
    const data = await api.getAppState();
    store.setState((state) => ({
      ...state,
      version: data.version,
      providers: data.providers,
      selectedProviderId: data.selectedProviderId || data.providers[0]?.id || "",
      defaultProviderId: data.defaultProviderId,
      defaultModelId: data.defaultModelId,
      settings: data.settings,
      logs: data.logs,
      drawer: null,
      modelMenuOpen: false
    }));
  } catch (error) {
    feedback.showError("应用初始化失败", error);
  }
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
  "data-import-models": providerActions.importModels,
  "data-delete-provider": confirmProviderDeletion,
  "data-set-default": appActions.setDefault,
  "data-launch-pi": appActions.directLaunch,
  "data-save-settings": appActions.saveSettings,
  "data-window-minimise": () => api.minimiseWindow(),
  "data-window-toggle-maximise": () => api.toggleMaximiseWindow(),
  "data-window-close": () => api.closeWindow()
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
    if (target.dataset.openProviderSettings) return openProvider(target.dataset.openProviderSettings);
    if (target.dataset.presetId) return providerActions.createFromPreset(target.dataset.presetId);
    if (target.dataset.confirmDeleteProvider) return providerActions.remove(target.dataset.confirmDeleteProvider);

    const attribute = findActionAttribute(target);
    if (attribute) await clickActions[attribute]();
  });
}

function bindFormEvents() {
  root.addEventListener("change", (event) => {
    const target = event.target;
    if (target.closest(".drawer") && target.matches("input[name], select[name]")) {
      providerForm.schedule();
    }
  });

  root.addEventListener("input", (event) => {
    if (event.target.closest(".drawer") && event.target.matches("input[name]")) {
      providerForm.schedule();
    }
  });
}

store.subscribe((state) => {
  root.innerHTML = renderApp(state);
  document.title = `Pi Switch ${state.version}`;
});

bindClickEvents();
bindFormEvents();
bootstrap();
