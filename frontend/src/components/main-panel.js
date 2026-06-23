import { escapeHtml } from "./view-utils.js";

export function renderMainPanel(state, provider) {
  const models = provider?.models ?? [];
  const selectedModelId = provider?.selectedModelId ?? models[0]?.id ?? "";
  const selectedModel = models.find((model) => model.id === selectedModelId);
  const defaultModelText = state.defaultModelId || "未设置";

  return `
    <main class="main-panel">
      <div class="main-topbar" data-wails-drag>
        <span class="topbar-context">${escapeHtml(provider?.host || "Provider workspace")}</span>
        <div class="window-actions">
          <button class="window-symbol" data-window-minimise aria-label="最小化">−</button>
          <button class="window-symbol" data-window-toggle-maximise aria-label="最大化">□</button>
          <button class="window-symbol window-symbol--danger" data-window-close aria-label="关闭">×</button>
        </div>
      </div>

      <div class="main-content">
        <div class="provider-heading">
          <span class="provider-heading__signal"></span>
          <span>${escapeHtml(provider?.name ?? "未选择提供商")}</span>
        </div>

        <section class="launch-workspace">
          <span class="field-label">运行模型</span>
          <div class="model-select ${state.modelMenuOpen ? "is-open" : ""}">
            <button
              class="model-select__trigger"
              data-toggle-model-menu
              aria-haspopup="listbox"
              aria-expanded="${state.modelMenuOpen}"
              ${models.length ? "" : "disabled"}
            >
              <span>${escapeHtml(selectedModel?.name ?? "请先获取模型")}</span>
              <span class="model-select__arrow" aria-hidden="true">⌄</span>
            </button>
            ${
              models.length
                ? `<div class="model-select__menu" role="listbox">
                    ${models
                      .map(
                        (model) => `
                          <button
                            class="model-select__option ${model.id === selectedModelId ? "is-selected" : ""}"
                            data-select-model="${escapeHtml(model.id)}"
                            role="option"
                            aria-selected="${model.id === selectedModelId}"
                          >
                            <span>${escapeHtml(model.name)}</span>
                            ${model.id === selectedModelId ? '<span aria-hidden="true">·</span>' : ""}
                          </button>
                        `
                      )
                      .join("")}
                  </div>`
                : ""
            }
          </div>

          <div class="launch-actions">
            <button class="primary-action text-button text-button--accent" data-launch-pi ${provider && selectedModelId ? "" : "disabled"}>
              <span>启动 Pi Agent</span>
              <span aria-hidden="true">↗</span>
            </button>
            <button class="text-button" data-set-default ${selectedModelId ? "" : "disabled"}>设为默认模型</button>
          </div>
        </section>

        <div class="default-summary">
          <span>DEFAULT ROUTE</span>
          <strong>${escapeHtml(state.defaultProviderId || "未设置")} / ${escapeHtml(defaultModelText)}</strong>
        </div>
      </div>

      <footer class="status-footer">
        <span><i class="status-dot"></i>配置服务已连接</span>
        <span class="status-path">${escapeHtml(state.settings.piSettingsPath || "未指定配置路径")}</span>
      </footer>
    </main>
  `;
}
