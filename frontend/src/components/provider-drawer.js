import { escapeHtml } from "./view-utils.js";
import { HEADER_MODES } from "../config/header-presets.js";

function apiModeOptions(selectedValue) {
  const options = [
    { value: "openai-completions", label: "OpenAI Chat Completions" },
    { value: "openai-responses", label: "OpenAI Responses" },
    { value: "anthropic-messages", label: "Anthropic Messages" }
  ];

  return options
    .map(
      (item) =>
        `<option value="${item.value}" ${item.value === selectedValue ? "selected" : ""}>${item.label}</option>`
    )
    .join("");
}

function field({ label, name, value, type = "text" }) {
  return `
    <label class="form-field">
      <span class="form-field__label">${label}</span>
      <input type="${type}" name="${name}" value="${escapeHtml(value)}" autocomplete="off">
    </label>
  `;
}

export function modelManageList(provider) {
  const models = provider.models ?? [];
  return `
    <section class="form-section model-manage">
      <div class="form-section__head">
        <h4 class="form-section__title">模型${models.length ? `（${models.length}）` : ""}</h4>
        <button class="text-button" type="button" data-add-model>添加模型</button>
      </div>
      ${models.length ? `
        <div class="model-manage__list">
          ${models
            .map((model) => {
              const cw = model.contextWindow ? Math.round(model.contextWindow / 1000) + "K" : "";
              const label = model.name || model.id;
              return `
                <div class="model-manage__row">
                  <span class="model-manage__main">
                    <span class="model-manage__name">${escapeHtml(label)}</span>
                    ${label !== model.id ? `<span class="model-manage__id">${escapeHtml(model.id)}</span>` : ""}
                  </span>
                  <span class="model-manage__meta">
                    ${cw ? `<span>${escapeHtml(cw)}</span>` : ""}
                    ${model.reasoning ? "<span>推理</span>" : ""}
                  </span>
                  <button type="button" class="model-manage__edit" data-edit-model="${escapeHtml(model.id)}" aria-label="编辑 ${escapeHtml(label)}" title="编辑">编辑</button>
                  <button type="button" class="model-manage__remove" data-delete-model="${escapeHtml(model.id)}" aria-label="移除 ${escapeHtml(label)}" title="移除">×</button>
                </div>
              `;
            })
            .join("")}
        </div>
      ` : '<p class="model-manage__empty">尚未添加模型，可点「添加模型」录入，或用底部「获取模型」自动导入。</p>'}
    </section>
  `;
}

function headerModeSection(provider) {
  const savedMode = provider.headerMode === "auto" ? "none" : provider.headerMode;
  const activeMode = savedMode || (Object.keys(provider.headers ?? {}).length ? "custom" : "none");
  return `
    <section class="form-section">
      <h4 class="form-section__title">请求头身份</h4>
      <div class="header-mode-grid">
        ${HEADER_MODES.map((mode) => `
          <button
            class="header-mode-button ${mode.id === activeMode ? "is-active" : ""}"
            type="button"
            data-set-header-mode="${mode.id}"
          >${escapeHtml(mode.label)}</button>
        `).join("")}
      </div>
      <button class="text-button header-mode-edit ${activeMode === "custom" ? "" : "is-hidden"}" type="button" data-open-headers-editor>编辑自定义 Header</button>
    </section>
  `;
}

export function renderProviderDrawer(state, provider) {
  if (!state.drawer || state.drawer.kind !== "provider" || !provider) {
    return "";
  }

  return `
    <section class="drawer open" aria-label="提供商配置">
      <header class="drawer-header">
        <h2>${escapeHtml(provider.name)}</h2>
        <button class="text-button drawer-close" data-close-drawer aria-label="关闭配置">关闭</button>
      </header>

      <div class="drawer-body">
        <section class="form-section">
          <h4 class="form-section__title">连接</h4>
          ${field({ label: "显示名称", name: "name", value: provider.name })}
          ${field({ label: "基础地址", name: "baseUrl", value: provider.baseUrl })}
          ${field({ label: "API 密钥", name: "apiKeyLiteral", value: provider.apiKeyLiteral, type: "password" })}
          <label class="form-field">
            <span class="form-field__label">接口协议</span>
            <select name="api">${apiModeOptions(provider.api || "openai-completions")}</select>
          </label>
          ${field({ label: "Provider ID", name: "id", value: provider.id })}
        </section>

        ${headerModeSection(provider)}
        ${modelManageList(provider)}
      </div>

      <footer class="drawer-footer">
        <div class="drawer-footer__status">
          <span class="drawer-status" aria-live="polite"></span>
        </div>
        <div class="drawer-footer__actions">
          <button class="text-button text-button--danger" data-delete-provider>删除</button>
          <button class="primary-action" data-fetch-models>获取模型</button>
        </div>
      </footer>
    </section>
  `;
}
