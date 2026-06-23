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
      ${renderProviderDrawer(state, provider)}
      ${renderModal(state)}
    </div>
  `;
}
