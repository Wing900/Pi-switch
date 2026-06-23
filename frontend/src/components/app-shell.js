function providerSubtitle(provider) {
  return provider.host || "未配置";
}

function modelOptions(items, selectedValue) {
  return items
    .map(
      (item) =>
        `<option value="${item.id}" ${item.id === selectedValue ? "selected" : ""}>${item.name}</option>`
    )
    .join("");
}

function apiModeOptions(selectedValue) {
  const options = [
    { value: "openai-completions", label: "OpenAI Completions" },
    { value: "openai-responses", label: "OpenAI Responses" }
  ];

  return options
    .map(
      (item) =>
        `<option value="${item.value}" ${item.value === selectedValue ? "selected" : ""}>${item.label}</option>`
    )
    .join("");
}

function providerCards(state) {
  return state.providers
    .map(
      (provider) => `
        <button class="provider-card ${provider.id === state.selectedProviderId ? "active" : ""}" data-provider-id="${provider.id}">
          <span class="provider-card__header">
            <span class="provider-card__name">${provider.name}</span>
            <span class="provider-card__configure" data-open-provider-settings="${provider.id}">配置</span>
          </span>
          <span class="provider-card__meta">${providerSubtitle(provider)}</span>
        </button>
      `
    )
    .join("");
}

function providerDrawer(state, provider) {
  if (!state.drawer || state.drawer.kind !== "provider") {
    return "";
  }

  return `
    <section class="drawer ${state.drawer ? "open" : ""}">
      <div class="drawer-header">
        <div class="drawer-heading">
          <span class="drawer-title">配置</span>
          <span class="drawer-subtitle">${provider?.name ?? ""}</span>
        </div>
        <button class="drawer-close" data-close-drawer>×</button>
      </div>
      <div class="drawer-body">
        <div class="form-group">
          <label>Provider 名称</label>
          <input type="text" name="name" value="${provider?.name ?? ""}">
        </div>
        <div class="form-group">
          <label>API URL</label>
          <input type="text" name="baseUrl" value="${provider?.baseUrl ?? ""}">
        </div>
        <div class="form-group">
          <label>API Key</label>
          <input type="text" name="apiKey" value="${provider?.apiKeyLiteral || provider?.apiKeyEnv || ""}">
        </div>
        <div class="form-group">
          <label>API 模式</label>
          <select name="api">${apiModeOptions(provider?.api ?? "openai-completions")}</select>
        </div>
        <div class="form-group">
          <label>Provider ID</label>
          <input type="text" name="id" value="${provider?.id ?? ""}">
        </div>
        <div class="form-group form-group--actions">
          <div class="section-caption">模型发现</div>
          <button class="drawer-ghost" data-test-connection>测试连接</button>
          <button class="drawer-ghost" data-fetch-models>获取可用模型</button>
        </div>
      </div>
      <div class="drawer-footer">
        <button class="action-text action-text--danger" data-delete-provider>删除</button>
        <span class="drawer-status">${state.drawerSaveStatus === "saving" ? "保存中..." : state.drawerSaveStatus === "saved" ? "已自动保存" : ""}</span>
      </div>
    </section>
  `;
}

function modalMarkup(state) {
  if (!state.modal) {
    return "";
  }

  if (state.modal.kind === "test-result") {
    return `
      <div class="modal-backdrop" data-close-modal>
        <section class="modal-card">
          <header class="modal-header">
            <h3>${state.modal.payload.title}</h3>
          </header>
          <div class="modal-body modal-stack">
            ${state.modal.payload.lines.map((line) => `<p>${line}</p>`).join("")}
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>关闭</button>
          </footer>
        </section>
      </div>
    `;
  }

  if (state.modal.kind === "fetch-models") {
    return `
      <div class="modal-backdrop">
        <section class="modal-card modal-card--wide">
          <header class="modal-header">
            <h3>模型导入</h3>
          </header>
          <div class="modal-body">
            <p class="modal-note">${state.modal.payload.models.length} 个模型已检测到。</p>
            <div class="check-list">
              ${state.modal.payload.models
                .map(
                  (model) => `
                    <label class="check-item">
                      <input type="checkbox" data-model-id="${model.id}" ${model.selected ? "checked" : ""}>
                      <span>${model.name}</span>
                    </label>
                  `
                )
                .join("")}
            </div>
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>关闭</button>
            <button class="action-text" data-import-models>导入</button>
          </footer>
        </section>
      </div>
    `;
  }

  if (state.modal.kind === "settings") {
    const settings = state.settings;
    return `
      <div class="modal-backdrop">
        <section class="modal-card modal-card--wide">
          <header class="modal-header">
            <h3>设置</h3>
          </header>
          <div class="modal-body settings-grid">
            <label>
              <span>Pi 命令</span>
              <input name="piCommand" value="${settings.piCommand}">
            </label>
            <label>
              <span>Pi Settings 路径</span>
              <input name="piSettingsPath" value="${settings.piSettingsPath}">
            </label>
            <label>
              <span>Pi Models 路径</span>
              <input name="piModelsPath" value="${settings.piModelsPath}">
            </label>
            <label>
              <span>Pi Switch 配置路径</span>
              <input name="piSwitchConfigPath" value="${settings.piSwitchConfigPath}">
            </label>
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>关闭</button>
            <button class="action-text" data-save-settings>保存</button>
          </footer>
        </section>
      </div>
    `;
  }

  if (state.modal.kind === "launch-preview") {
    return `
      <div class="modal-backdrop">
        <section class="modal-card modal-card--wide">
          <header class="modal-header">
            <h3>启动 Pi</h3>
          </header>
          <div class="modal-body modal-stack">
            <p>命令：${state.modal.payload.command}</p>
            ${state.modal.payload.checklist.map((item) => `<p>✓ ${item}</p>`).join("")}
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>关闭</button>
            <button class="action-text" data-confirm-launch>启动终端</button>
          </footer>
        </section>
      </div>
    `;
  }

  if (state.modal.kind === "add-provider") {
    return `
      <div class="modal-backdrop">
        <section class="modal-card modal-card--wide">
          <header class="modal-header">
            <h3>新增 Provider</h3>
          </header>
          <div class="modal-body">
            <div class="preset-list">
              ${state.presets
                .map(
                  (preset) => `
                    <button class="preset-card" data-preset-id="${preset.id}">
                      <strong>${preset.label}</strong>
                      <span>${preset.baseUrl || "手动填写配置"}</span>
                    </button>
                  `
                )
                .join("")}
            </div>
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>关闭</button>
          </footer>
        </section>
      </div>
    `;
  }

  if (state.modal.kind === "confirm-delete") {
    return `
      <div class="modal-backdrop">
        <section class="modal-card modal-card--compact">
          <header class="modal-header">
            <h3>删除 Provider</h3>
          </header>
          <div class="modal-body modal-stack">
            <p>确定删除 ${state.modal.payload.name} 吗？</p>
            <p class="modal-note">该操作会更新本地配置。</p>
          </div>
          <footer class="modal-footer">
            <button class="action-text" data-close-modal>取消</button>
            <button class="action-text action-text--danger" data-confirm-delete-provider="${state.modal.payload.id}">删除</button>
          </footer>
        </section>
      </div>
    `;
  }

  return "";
}

export function renderApp(state, presets) {
  const provider = state.providers.find((item) => item.id === state.selectedProviderId) ?? state.providers[0];
  const models = provider?.models ?? [];
  const selectedModelId = provider?.selectedModelId ?? models[0]?.id ?? "";
  const defaultModelText = state.defaultModelId || "未设置默认模型";

  return `
    <div class="window-shell">
      <aside class="sidebar">
        <div class="aside-header">Pi Switch</div>
        <div class="provider-grid">${providerCards(state)}</div>
        <div class="sidebar-footer">
          <button class="action-text" data-open-add-provider>+ 添加 Provider</button>
          <button class="action-text" data-open-settings>设置</button>
        </div>
      </aside>

      <main class="main-panel">
        <div class="main-topbar" data-wails-drag>
          <div class="topbar-spacer"></div>
          <div class="window-actions">
            <button class="window-symbol" data-window-minimise>−</button>
            <button class="window-symbol" data-window-toggle-maximise>□</button>
            <button class="window-symbol window-symbol--danger" data-window-close>×</button>
          </div>
        </div>

        <div class="main-content">
          <div class="hero-icon" data-open-provider-settings="${provider?.id ?? ""}">
            <svg width="32" height="32" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
          </div>
          <h2 class="main-title">${provider?.name ?? "未选择 Provider"}</h2>

          <div class="model-select-area">
            <label>当前模型</label>
            <select data-model-select>${modelOptions(models, selectedModelId)}</select>
          </div>

          <button class="launch-frame" data-launch-pi>启动 piAgent</button>

          <div class="action-row">
            <button class="action-text" data-set-default>设为默认模型</button>
          </div>

          <p class="default-line">默认：${state.defaultProviderId || "未设置"} / ${defaultModelText}</p>
        </div>

        <div class="status-footer">
          系统状态：<span class="status-ok">已连接</span> · 配置：${state.settings.piSettingsPath}
        </div>
      </main>

      ${providerDrawer(state, provider)}
      ${modalMarkup({ ...state, presets })}
    </div>
  `;
}
