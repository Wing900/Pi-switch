import { renderMainPanel } from "./main-panel.js";
import { renderModal } from "./modals.js";
import { renderProviderDrawer } from "./provider-drawer.js";
import { renderProviderSidebar } from "./provider-sidebar.js";

function providerForState(state) {
  return state.providers.find((item) => item.id === state.selectedProviderId) ?? state.providers[0];
}

export function renderContentLayer(state) {
  const provider = providerForState(state);
  return `<div data-content-layer>${renderProviderSidebar(state)}${renderMainPanel(state, provider)}</div>`;
}

export function renderDrawerLayer(state) {
  const provider = providerForState(state);
  return `<div data-drawer-layer>${state.drawer ? '<button class="drawer-scrim" data-close-drawer aria-label="关闭配置"></button>' : ""}${renderProviderDrawer(state, provider)}</div>`;
}

export function renderModalLayer(state) {
  return `<div data-modal-layer>${renderModal(state)}</div>`;
}

export function renderApp(state) {
  const theme = state.settings?.darkMode ? "dark" : "light";

  return `
    <div class="window-shell" data-theme="${theme}">
      ${renderContentLayer(state)}
      ${renderDrawerLayer(state)}
      ${renderModalLayer(state)}
    </div>
  `;
}
