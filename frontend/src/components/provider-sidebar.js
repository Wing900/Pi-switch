import { escapeHtml, extractHost } from "./view-utils.js";

function providerEndpoint(provider) {
  return extractHost(provider.baseUrl);
}

function providerRows(state) {
  return state.providers
    .map(
      (provider) => `
        <div class="provider-row ${provider.id === state.selectedProviderId ? "active" : ""}">
          <button class="provider-row__select" data-provider-id="${escapeHtml(provider.id)}">
            <span class="provider-row__marker"></span>
            <span class="provider-row__content">
              <span class="provider-row__name">${escapeHtml(provider.name)}</span>
              <span class="provider-row__meta">${escapeHtml(providerEndpoint(provider))}</span>
            </span>
          </button>
          <button
            class="provider-row__configure"
            data-open-provider-settings="${escapeHtml(provider.id)}"
            aria-label="配置 ${escapeHtml(provider.name)}"
          >配置</button>
        </div>
      `
    )
    .join("");
}

export function renderProviderSidebar(state) {
  return `
    <aside class="sidebar">
      <header class="aside-header">
        <img class="brand-mark" src="./src/assets/pi-logo.png" alt="Pi Switch">
        <strong>Pi Switch</strong>
      </header>

      <div class="sidebar-section-label">模型提供商</div>
      <nav class="provider-list" aria-label="模型提供商">
        ${providerRows(state)}
      </nav>

      <footer class="sidebar-footer">
        <button class="text-button text-button--accent" data-open-add-provider>添加提供商</button>
        <button class="text-button" data-open-settings>应用设置</button>
      </footer>
    </aside>
  `;
}
