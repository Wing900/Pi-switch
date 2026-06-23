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
    actions: '<button class="text-button text-button--accent" data-close-modal>知道了</button>'
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

function fetchModelsModal(payload) {
  const models = payload.models ?? [];
  return modalFrame({
    wide: true,
    title: "选择要导入的模型",
    description: `已获取 ${models.length} 个模型`,
    body: `
      <div class="model-check-list">
        ${models
          .map(
            (model) => `
              <label class="model-check-row">
                <input type="checkbox" data-model-id="${escapeHtml(model.id)}" ${model.selected ? "checked" : ""}>
                <span class="model-check-row__box"></span>
                <span>
                  <strong>${escapeHtml(model.name)}</strong>
                  <small>${model.reasoning ? "推理模型" : "通用模型"}</small>
                </span>
              </label>
            `
          )
          .join("")}
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
      <button class="text-button text-button--accent" data-import-models>导入所选模型</button>
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

  return modalFrame({
    wide: true,
    title: "应用设置",
    body: `
      <div class="settings-form">
        ${settingsField("Pi 命令", "piCommand", settings.piCommand)}
        ${settingsField("Pi 设置文件", "piSettingsPath", settings.piSettingsPath)}
        ${settingsField("Pi 模型文件", "piModelsPath", settings.piModelsPath)}
        ${settingsField("Pi Switch 配置", "piSwitchConfigPath", settings.piSwitchConfigPath)}
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
      <button class="text-button text-button--accent" data-save-settings>保存设置</button>
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
  if (modal.kind === "fetch-models") return fetchModelsModal(modal.payload);
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
