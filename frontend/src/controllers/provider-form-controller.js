function readField(root, name, fallback = "") {
  return root.querySelector(`.drawer [name="${name}"]`)?.value.trim() ?? fallback;
}

function setStatus(root, status, message) {
  const element = root.querySelector(".drawer-status");
  if (!element) {
    return;
  }
  element.dataset.status = status;
  element.textContent = message;
}

function providerFromForm(root, provider) {
  return {
    ...provider,
    id: readField(root, "id", provider.id),
    name: readField(root, "name", provider.name),
    baseUrl: readField(root, "baseUrl", provider.baseUrl),
    api: readField(root, "api", provider.api),
    apiKeyLiteral: readField(root, "apiKeyLiteral", provider.apiKeyLiteral),
    apiKeyEnv: ""
  };
}

function validateProviderDraft(provider) {
  if (!provider.id) return "Provider ID 不能为空";
  if (!provider.name) return "Provider 名称不能为空";
  if (!provider.baseUrl) return "Base URL 不能为空";
  if (!provider.api) return "API 模式不能为空";
  return "";
}

export function createProviderFormController({ root, api, store, getCurrentProvider, onError }) {
  let timer = null;
  let saveChain = Promise.resolve();
  let revision = 0;

  function showInvalidModal(message) {
    store.setState((state) => ({
      ...state,
      modal: {
        kind: "provider-invalid",
        payload: {
          title: "配置未完成",
          message
        }
      }
    }));
  }

  function commit({ showInvalidModal: shouldShowInvalidModal = true } = {}) {
    if (timer) {
      window.clearTimeout(timer);
      timer = null;
    }
    const provider = getCurrentProvider(store.getState());
    if (!provider) {
      return Promise.resolve();
    }

    const nextProvider = providerFromForm(root, provider);
    const validationMessage = validateProviderDraft(nextProvider);
    if (validationMessage) {
      setStatus(root, "error", validationMessage);
      if (shouldShowInvalidModal) {
        showInvalidModal(validationMessage);
      }
      return Promise.resolve(false);
    }

    const currentRevision = ++revision;
    setStatus(root, "saving", "正在保存");

    saveChain = saveChain
      .catch(() => undefined)
      .then(async () => {
        const persistedProvider = getCurrentProvider(store.getState());
        const persistedId = persistedProvider?.id ?? provider.id;
        await api.updateProvider(persistedId, nextProvider);
        return persistedId;
      })
      .then((persistedId) => {
        store.setState(
          (state) => ({
            ...state,
            providers: state.providers.map((item) => (item.id === persistedId ? nextProvider : item)),
            selectedProviderId: nextProvider.id,
            defaultProviderId:
              state.defaultProviderId === persistedId ? nextProvider.id : state.defaultProviderId,
            drawer: { kind: "provider", providerId: nextProvider.id }
          }),
          { notify: false }
        );
        if (currentRevision === revision) {
          setStatus(root, "saved", "已自动保存");
        }
        return true;
      })
      .catch((error) => {
        setStatus(root, "error", "保存失败");
        onError("配置保存失败", error);
        return false;
      });

    return saveChain;
  }

  function schedule() {
    if (timer) {
      window.clearTimeout(timer);
    }

    const provider = getCurrentProvider(store.getState());
    if (!provider) {
      return;
    }

    const nextProvider = providerFromForm(root, provider);
    const validationMessage = validateProviderDraft(nextProvider);
    if (validationMessage) {
      setStatus(root, "error", validationMessage);
      return;
    }

    const delay = 500;
    setStatus(root, "saving", "等待保存");
    timer = window.setTimeout(() => {
      timer = null;
      void commit({ showInvalidModal: false });
    }, delay);
  }

  function cancel() {
    if (timer) {
      window.clearTimeout(timer);
      timer = null;
    }
  }

  return { schedule, commit, cancel };
}
