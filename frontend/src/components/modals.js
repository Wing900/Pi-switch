import { escapeHtml } from "./view-utils.js";

function modalFrame({ tone = "", eyebrow = "", title, description = "", body = "", actions = "", wide = false }) {
  return `
    <div class="modal-backdrop">
      <section class="modal-dialog ${wide ? "modal-dialog--wide" : ""} ${tone ? `modal-dialog--${tone}` : ""}" role="dialog" aria-modal="true">
        <header class="modal-header">
          ${eyebrow ? `<span class="eyebrow">${escapeHtml(eyebrow)}</span>` : ""}
          <h3>${escapeHtml(title)}</h3>
          ${description ? `<p>${escapeHtml(description)}</p>` : ""}
        </header>
        ${body ? `<div class="modal-body">${body}</div>` : ""}
        <footer class="modal-footer">${actions}</footer>
      </section>
    </div>
  `;
}

function resultModal(payload) {
  const success = payload.status === "success";
  return modalFrame({
    tone: success ? "success" : "error",
    title: payload.title,
    description: payload.message,
    body: payload.details?.length
      ? `<div class="result-list">${payload.details
          .map((item) => `<div class="result-row"><span></span><p>${escapeHtml(item)}</p></div>`)
          .join("")}</div>`
      : "",
    actions: `
      ${payload.allowManualModel ? '<button class="text-button modal-footer__manual-action" data-open-manual-model>手动添加</button>' : ""}
      <button class="text-button text-button--accent" data-close-modal>知道了</button>
    `
  });
}

function loadingModal(payload) {
  return modalFrame({
    title: payload.title,
    description: payload.message,
    body: '<div class="loading-line" role="progressbar" aria-label="正在处理"><span></span></div>',
    actions: ""
  });
}

function providerInvalidModal(payload) {
  return modalFrame({
    tone: "error",
    title: payload.title,
    description: payload.message,
    actions: `
      <button class="text-button text-button--accent" data-close-modal>继续编辑</button>
    `
  });
}

function fetchModelsModal(payload) {
  const models = payload.models ?? [];
  const allSelected = models.length > 0 && models.every((model) => model.selected);
  return modalFrame({
    wide: true,
    title: "选择要导入的模型",
    description: `已获取 ${models.length} 个模型`,
    body: `
      <div class="model-check-list">
        ${models
          .map(
            (model) => {
              const cw = model.contextWindow ? Math.round(model.contextWindow / 1000) : 256;
              return `
              <div class="model-check-row">
                <label class="model-check-row__main">
                  <input type="checkbox" data-model-id="${escapeHtml(model.id)}" ${model.selected ? "checked" : ""}>
                  <span class="model-check-row__box"></span>
                  <span>
                    <strong>${escapeHtml(model.name)}</strong>
                    <small>${model.reasoning ? "推理模型" : "通用模型"}</small>
                  </span>
                </label>
                <div class="model-cw-wrap">
                  <input
                    type="number"
                    class="model-cw-input"
                    data-cw-model="${escapeHtml(model.id)}"
                    value="${cw}"
                    min="1"
                    placeholder="256"
                    title="上下文窗口 (K tokens)"
                    autocomplete="off"
                  >
                  <span class="model-cw-unit">K</span>
                </div>
              </div>
            `;
            }
          )
          .join("")}
      </div>
    `,
    actions: `
      <button class="text-button modal-footer__manual-action" data-toggle-model-selection-all>${allSelected ? "全不选" : "全选"}</button>
      <button class="text-button" data-close-modal>取消</button>
      <button class="text-button text-button--accent" data-import-models>导入所选模型</button>
    `
  });
}

function manualModelModal(payload) {
  const contextWindowK = payload.contextWindowK || 256;
  return modalFrame({
    wide: true,
    title: "手动添加模型",
    description: "获取模型失败时，可手动录入当前提供商的模型配置。",
    body: `
      <div class="manual-model-form">
        <div class="model-check-row model-check-row--manual">
          <label class="manual-model-field manual-model-field--id">
            <span class="manual-model-field__label">Model ID</span>
            <input
              type="text"
              class="manual-model-input"
              name="manualModelId"
              value="${escapeHtml(payload.modelId || "")}"
              placeholder="deepseek-v4-flash"
              autocomplete="off"
            >
          </label>
          <label class="manual-model-field manual-model-field--cw">
            <span class="manual-model-field__label">Context</span>
            <div class="model-cw-wrap">
              <input
                type="number"
                class="model-cw-input model-cw-input--manual"
                name="manualContextWindow"
                value="${contextWindowK}"
                min="1"
                placeholder="256"
                title="上下文窗口 (K tokens)"
                autocomplete="off"
              >
              <span class="model-cw-unit">K</span>
            </div>
          </label>
        </div>
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
      <button class="text-button text-button--accent" data-import-manual-model>导入模型</button>
    `
  });
}

function settingsModal(settings) {
  const settingsField = (label, name, value) => `
    <label class="form-field">
      <span class="form-field__label">${label}</span>
      <input name="${name}" value="${escapeHtml(value)}">
    </label>
  `;

  const toggleField = (label, name, checked) => `
    <label class="settings-toggle">
      <span class="settings-toggle__copy">
        <span class="form-field__label">${label}</span>
      </span>
      <span class="switch">
        <input type="checkbox" name="${name}" ${checked ? "checked" : ""}>
        <span class="switch__track" aria-hidden="true"></span>
      </span>
    </label>
  `;

  return modalFrame({
    wide: true,
    title: "应用设置",
    body: `
      <div class="settings-form">
        ${toggleField("黑暗模式", "darkMode", !!settings.darkMode)}
        ${settingsField("Pi 命令", "piCommand", settings.piCommand)}
        ${settingsField("Pi 设置文件", "piSettingsPath", settings.piSettingsPath)}
        ${settingsField("Pi 模型文件", "piModelsPath", settings.piModelsPath)}
        ${settingsField("Pi Switch 配置", "piSwitchConfigPath", settings.piSwitchConfigPath)}
        ${settingsField("工作目录", "workingDir", settings.workingDir || "")}
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
    `
  });
}

export function renderModal(state) {
  const modal = state.modal;
  if (!modal) {
    return "";
  }

  if (modal.kind === "operation-loading") return loadingModal(modal.payload);
  if (modal.kind === "operation-result") return resultModal(modal.payload);
  if (modal.kind === "provider-invalid") return providerInvalidModal(modal.payload);
  if (modal.kind === "fetch-models") return fetchModelsModal(modal.payload);
  if (modal.kind === "manual-model") return manualModelModal(modal.payload);
  if (modal.kind === "settings") return settingsModal(state.settings);

  if (modal.kind === "add-provider") {
    return modalFrame({
      wide: true,
      title: "选择配置模板",
      body: `<div class="preset-list">${state.presets
        .map(
          (preset) => `
            <button class="preset-row" data-preset-id="${escapeHtml(preset.id)}">
              <span><strong>${escapeHtml(preset.label)}</strong><small>${escapeHtml(preset.baseUrl || "手动配置")}</small></span>
              <span>→</span>
            </button>
          `
        )
        .join("")}</div>`,
      actions: '<button class="text-button" data-close-modal>取消</button>'
    });
  }

  if (modal.kind === "confirm-delete") {
    return modalFrame({
      tone: "error",
      title: "删除提供商",
      description: `确定删除 ${modal.payload.name}？本地配置将同步更新。`,
      actions: `
        <button class="text-button" data-close-modal>取消</button>
        <button class="text-button text-button--danger" data-confirm-delete-provider="${escapeHtml(modal.payload.id)}">确认删除</button>
      `
    });
  }

  return "";
}
