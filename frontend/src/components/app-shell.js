import { renderMainPanel } from "./main-panel.js";
import { renderModal } from "./modals.js";
import { renderProviderDrawer } from "./provider-drawer.js";
import { renderProviderSidebar } from "./provider-sidebar.js";

export function renderApp(state) {
  const provider =
    state.providers.find((item) => item.id === state.selectedProviderId) ?? state.providers[0];

  return `
    <div class="window-shell">
      ${renderProviderSidebar(state)}
      ${renderMainPanel(state, provider)}
      ${state.drawer ? '<button class="drawer-scrim" data-close-drawer aria-label="关闭配置"></button>' : ""}
      ${renderProviderDrawer(state, provider)}
      ${renderModal(state)}
    </div>
  `;
}
