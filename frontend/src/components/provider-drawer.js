import { escapeHtml } from "./view-utils.js";

function apiModeOptions(selectedValue) {
  const options = [
    { value: "openai-completions", label: "OpenAI Chat Completions" },
    { value: "openai-responses", label: "OpenAI Responses" }
  ];

  return options
    .map(
      (item) =>
        `<option value="${item.value}" ${item.value === selectedValue ? "selected" : ""}>${item.label}</option>`
    )
    .join("");
}

function field({ label, name, value }) {
  return `
    <label class="form-field">
      <span class="form-field__label">${label}</span>
      <input type="text" name="${name}" value="${escapeHtml(value)}" autocomplete="off">
    </label>
  `;
}

export function renderProviderDrawer(state, provider) {
  if (!state.drawer || state.drawer.kind !== "provider" || !provider) {
    return "";
  }

  return `
    <section class="drawer open" aria-label="提供商配置">
      <header class="drawer-header">
        <div class="drawer-heading">
          <span class="eyebrow">PROVIDER CONFIG</span>
          <h2>${escapeHtml(provider.name)}</h2>
        </div>
        <button class="text-button drawer-close" data-close-drawer aria-label="关闭配置">关闭</button>
      </header>

      <div class="drawer-body">
        <section class="form-section">
          ${field({ label: "显示名称", name: "name", value: provider.name })}
          ${field({ label: "API 基础地址", name: "baseUrl", value: provider.baseUrl })}
          <label class="form-field">
            <span class="form-field__label">接口协议</span>
            <select name="api">${apiModeOptions(provider.api || "openai-completions")}</select>
          </label>
          ${field({ label: "Provider ID", name: "id", value: provider.id })}
        </section>
      </div>

      <footer class="drawer-footer">
        <div class="drawer-footer__status">
          <span class="drawer-status" aria-live="polite"></span>
        </div>
        <div class="drawer-footer__actions">
          <button class="text-button text-button--danger" data-delete-provider>删除</button>
          <button class="text-button text-button--accent" data-fetch-models>获取模型</button>
        </div>
      </footer>
    </section>
  `;
}
