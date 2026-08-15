import { escapeHtml } from "./view-utils.js";
import { COMPAT_BOOLEAN_FIELDS, COMPAT_ENUM_FIELDS, THINKING_LEVELS } from "../config/model-fields.js";

function modalFrame({ tone = "", eyebrow = "", title, description = "", body = "", actions = "", wide = false, className = "" }) {
  return `
    <div class="modal-backdrop">
      <section class="modal-dialog ${wide ? "modal-dialog--wide" : ""} ${className} ${tone ? `modal-dialog--${tone}` : ""}" role="dialog" aria-modal="true">
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

function modelTextField({ label, name, value = "", description = "", type = "text", step = "", min = "" }) {
  return `
    <label class="model-editor-field">
      <span class="model-editor-field__label">${escapeHtml(label)}</span>
      <input name="${escapeHtml(name)}" type="${type}" value="${escapeHtml(value)}" autocomplete="off" ${step ? `step="${step}"` : ""} ${min ? `min="${min}"` : ""}>
      ${description ? `<small>${escapeHtml(description)}</small>` : ""}
    </label>
  `;
}

function modelJsonField({ label, name, value, placeholder, description }) {
  return `
    <label class="model-editor-json-field">
      <span class="model-editor-field__label">${escapeHtml(label)}</span>
      <textarea name="${escapeHtml(name)}" spellcheck="false" autocomplete="off" placeholder="${escapeHtml(placeholder || "{}")}">${value ? escapeHtml(JSON.stringify(value, null, 2)) : ""}</textarea>
      ${description ? `<small>${escapeHtml(description)}</small>` : ""}
    </label>
  `;
}

function modelRawJsonField({ label, name, value, placeholder, description }) {
  return `
    <label class="model-editor-json-field">
      <span class="model-editor-field__label">${escapeHtml(label)}</span>
      <textarea name="${escapeHtml(name)}" spellcheck="false" autocomplete="off" placeholder="${escapeHtml(placeholder || "{}")}">${value ? escapeHtml(value) : ""}</textarea>
      ${description ? `<small>${escapeHtml(description)}</small>` : ""}
    </label>
  `;
}

function modelTriStateField({ label, name, value, description }) {
  const selected = value === undefined ? "inherit" : value ? "true" : "false";
  return `
    <label class="model-editor-compat-field">
      <span>
        <strong>${escapeHtml(label)}</strong>
        <small>${escapeHtml(description)}</small>
      </span>
      <select name="${escapeHtml(name)}" data-model-compat-boolean>
        <option value="inherit" ${selected === "inherit" ? "selected" : ""}>继承 / 未指定</option>
        <option value="true" ${selected === "true" ? "selected" : ""}>是</option>
        <option value="false" ${selected === "false" ? "selected" : ""}>否</option>
      </select>
    </label>
  `;
}

function modelThinkingField(level, label, value) {
  const mode = value === undefined ? "inherit" : value === null ? "disabled" : "custom";
  return `
    <div class="model-thinking-row">
      <span class="model-thinking-row__name">${escapeHtml(label)}</span>
      <select name="modelThinkingMode_${level}" data-model-thinking-mode>
        <option value="inherit" ${mode === "inherit" ? "selected" : ""}>继承默认</option>
        <option value="disabled" ${mode === "disabled" ? "selected" : ""}>不支持（隐藏）</option>
        <option value="custom" ${mode === "custom" ? "selected" : ""}>发送自定义值</option>
      </select>
      <input name="modelThinkingValue_${level}" value="${escapeHtml(typeof value === "string" ? value : "")}" placeholder="例如 high" autocomplete="off" ${mode === "custom" ? "" : "disabled"}>
    </div>
  `;
}

function modelEditorModal(payload) {
  const model = payload.model || {};
  const isEdit = payload.mode === "edit";
  const inputTypes = Array.isArray(model.input) && model.input.length ? model.input : ["text"];
  const cost = model.cost || {};
  const compat = model.compat || {};
  const thinkingLevelMap = model.thinkingLevelMap || {};
  const checkbox = (name, value, label, description = "") => `
    <label class="model-editor-check">
      <input type="checkbox" name="${escapeHtml(name)}" value="${escapeHtml(value)}" ${inputTypes.includes(value) ? "checked" : ""}>
      <span>${escapeHtml(label)}</span>
      ${description ? `<small>${escapeHtml(description)}</small>` : ""}
    </label>
  `;
  const costField = (name, label, description) => modelTextField({
    label,
    name: `modelCost_${name}`,
    value: Object.prototype.hasOwnProperty.call(cost, name) ? cost[name] : "",
    type: "number",
    step: "any",
    description
  });

  return modalFrame({
    wide: true,
    className: "modal-dialog--model-editor",
    title: isEdit ? "编辑模型" : "添加模型",
    description: "按 Pi models.json 的模型定义编辑。灰色说明是字段含义和兼容性提示；保存后仍会保留其他未识别字段。",
    body: `
      <div class="model-editor">
        <section class="model-editor-section">
          <div class="model-editor-section__heading">
            <h4>基本信息</h4>
            <p>决定 Pi 如何识别此模型以及在列表中的显示方式。</p>
          </div>
          <div class="model-editor-grid model-editor-grid--two">
            ${modelTextField({ label: "模型 ID *", name: "modelId", value: model.id || "", description: "会原样传给服务端，也是模型在 Pi 中的唯一标识。" })}
            ${modelTextField({ label: "显示名称", name: "modelName", value: model.name || "", description: "仅用于界面显示和匹配；留空时使用模型 ID。" })}
          </div>
        </section>

        <section class="model-editor-section">
          <div class="model-editor-section__heading">
            <h4>能力与上下文</h4>
            <p>这些字段主要影响 Pi 是否显示图片、是否启用思考，以及何时压缩上下文。</p>
          </div>
          <div class="model-editor-grid model-editor-grid--two">
            ${modelTextField({ label: "上下文窗口（tokens）", name: "modelContextWindow", value: model.contextWindow ?? 128000, type: "number", min: "1", description: "模型一次能接收的最大上下文，不是字节数；填小了会提前压缩，填大了可能被服务端拒绝。" })}
            ${modelTextField({ label: "最大输出（tokens）", name: "modelMaxTokens", value: model.maxTokens ?? 16384, type: "number", min: "1", description: "单次回复最多生成的 token 数，不等于上下文窗口。" })}
          </div>
          <label class="model-editor-toggle">
            <input type="checkbox" name="modelReasoning" ${model.reasoning ? "checked" : ""}>
            <span>
              <strong>支持扩展思考 / reasoning</strong>
              <small>开启后 Pi 可能显示思考级别并发送思考相关参数；服务端不支持时不要开启。</small>
            </span>
          </label>
          <div class="model-editor-subheading">输入类型</div>
          <div class="model-editor-checks">
            ${checkbox("modelInput", "text", "文本", "几乎所有模型都支持。")}
            ${checkbox("modelInput", "image", "图片", "只有视觉模型和对应接口支持；填写后 Pi 可能发送图片内容。")}
          </div>
        </section>

        <section class="model-editor-section">
          <div class="model-editor-section__heading">
            <h4>思考级别映射</h4>
            <p>用于把 Pi 的 off / low / high 等级转换为服务商实际接受的值。选择“不支持”会在 Pi 中隐藏该级别。</p>
          </div>
          <div class="model-thinking-grid">
            ${THINKING_LEVELS.map(([level, label]) => modelThinkingField(level, label, thinkingLevelMap[level])).join("")}
          </div>
        </section>

        <details class="model-editor-section model-editor-section--collapsible">
          <summary class="model-editor-section__heading model-editor-section__summary">
            <h4>计费信息</h4>
            <p>单位是每百万 token 的价格，仅影响 Pi 的成本统计，不会改变服务端请求。</p>
          </summary>
          <div class="model-editor-section__content">
            <div class="model-editor-grid model-editor-grid--four">
              ${costField("input", "输入", "普通输入 token 单价。")}
              ${costField("output", "输出", "输出 token 单价。")}
              ${costField("cacheRead", "缓存读取", "命中提示缓存时的读取单价。")}
              ${costField("cacheWrite", "缓存写入", "写入提示缓存时的单价。")}
            </div>
            ${modelJsonField({ label: "分段价格 tiers", name: "modelCostTiers", value: cost.tiers, placeholder: '[{"inputTokensAbove":272000,"input":10,"output":45,"cacheRead":1,"cacheWrite":12.5}]', description: "当单次请求输入总量超过 inputTokensAbove 时，整次请求使用最高匹配档位；不是只对超出部分计价。" })}
          </div>
        </details>

        <details class="model-editor-section model-editor-section--collapsible">
          <summary class="model-editor-section__heading model-editor-section__summary">
            <h4>请求参数</h4>
            <p>这些 JSON 会按 Pi 支持的规则合并到请求中；格式错误会阻止保存。</p>
          </summary>
          <div class="model-editor-section__content">
            ${modelJsonField({ label: "samplingParams", name: "modelSamplingParams", value: model.samplingParams, placeholder: '{"temperature":0.7,"top_p":0.95,"min_p":0}', description: "OpenAI 兼容接口的自由参数，会原样合并到请求体；可填写服务商专属的 temperature、top_k、min_p 等字段。" })}
            ${modelJsonField({ label: "模型级 headers", name: "modelHeaders", value: model.headers, placeholder: '{"x-model-route":"fast"}', description: "仅此模型使用的请求头；敏感值仍建议使用环境变量或由 Provider 统一配置。" })}
          </div>
        </details>

        <details class="model-editor-section model-editor-section--collapsible">
          <summary class="model-editor-section__heading model-editor-section__summary">
            <h4>接口兼容性 compat</h4>
            <p>当服务商“看起来兼容”但字段名、角色或工具协议有差异时，在这里告诉 Pi 如何调整请求。</p>
          </summary>
          <div class="model-editor-section__content">
            <div class="model-editor-compat-list">
              ${COMPAT_BOOLEAN_FIELDS.map(([name, label, description]) => modelTriStateField({ label, name: `modelCompat_${name}`, value: compat[name], description })).join("")}
            </div>
            <div class="model-editor-grid model-editor-grid--two model-editor-enums">
              ${COMPAT_ENUM_FIELDS.map(([name, label, options, description]) => {
                const currentValue = compat[name];
                const hasKnownValue = options.some(([value]) => value === currentValue);
                return `
                  <label class="model-editor-field">
                    <span class="model-editor-field__label">${escapeHtml(label)}</span>
                    <select name="modelCompat_${escapeHtml(name)}">
                      <option value="inherit">继承 / 未指定</option>
                      ${!hasKnownValue && currentValue !== undefined ? `<option value="__custom__" selected>保留当前值：${escapeHtml(currentValue)}</option>` : ""}
                      ${options.map(([value, optionLabel]) => `<option value="${escapeHtml(value)}" ${currentValue === value ? "selected" : ""}>${escapeHtml(optionLabel)}</option>`).join("")}
                    </select>
                    <small>${escapeHtml(description)}</small>
                  </label>
                `;
              }).join("")}
            </div>
            ${modelJsonField({ label: "chatTemplateKwargs", name: "modelCompat_chatTemplateKwargs", value: compat.chatTemplateKwargs, placeholder: '{"thinking":{"$var":"thinking.enabled"}}', description: "chat-template 思考参数；$var 会由 Pi 替换为当前思考开关或等级。" })}
            ${modelJsonField({ label: "chatTemplateArgs", name: "modelCompat_chatTemplateArgs", value: compat.chatTemplateArgs, placeholder: '{"enable_thinking":{"$var":"thinking.enabled"}}', description: "Baseten 等接口使用的模板参数；只有对应 thinkingFormat 时才生效。" })}
            ${modelJsonField({ label: "openRouterRouting", name: "modelCompat_openRouterRouting", value: compat.openRouterRouting, placeholder: '{"only":["anthropic"],"allow_fallbacks":false}', description: "OpenRouter 路由约束，例如 only、order、ignore、价格或延迟偏好。" })}
            ${modelJsonField({ label: "vercelGatewayRouting", name: "modelCompat_vercelGatewayRouting", value: compat.vercelGatewayRouting, placeholder: '{"only":["anthropic"]}', description: "Vercel AI Gateway 的 Provider 路由顺序或白名单。" })}
            ${modelRawJsonField({ label: "其他 compat 字段", name: "modelCompatExtraJson", value: model.compatExtraFieldsJson, placeholder: '{"futureCompat":true}', description: "Pi 新版本或服务商扩展的 compat 字段；这里可以新增、修改或删除，结构化字段冲突时以上方表单为准。" })}
          </div>
        </details>

        <section class="model-editor-section model-editor-section--advanced">
          <div class="model-editor-section__heading">
            <h4>其他字段</h4>
            <p>Pi 新版本或扩展支持的字段可以放在这里；保存时会与上面的字段合并，字段名冲突时以上方表单为准。</p>
          </div>
          <textarea class="model-editor-extra" name="modelExtraJson" spellcheck="false" autocomplete="off" placeholder='{"vendorOption":true}'>${model.extraFieldsJson ? escapeHtml(model.extraFieldsJson) : ""}</textarea>
        </section>
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
      <button class="text-button text-button--accent" data-save-model>保存模型</button>
    `
  });
}

function headersModal(payload) {
  const headers = Object.entries(payload.headers ?? {});
  const rows = headers.length > 0 ? headers : [["", ""]];
  return modalFrame({
    wide: true,
    title: "自定义请求头",
    body: `
      <div class="headers-editor">
        <div class="headers-editor__head"><span>Header 名称</span><span>Header 值</span></div>
        <div class="headers-editor__rows">
          ${rows.map(([name, value]) => `
            <div class="headers-editor__row">
              <input class="headers-editor__input" data-header-name value="${escapeHtml(name)}" placeholder="user-agent" autocomplete="off">
              <input class="headers-editor__input" data-header-value value="${escapeHtml(value)}" placeholder="custom-client/1.0" autocomplete="off">
            </div>
          `).join("")}
        </div>
      </div>
    `,
    actions: '<span class="headers-editor__autosave">修改后自动保存</span><button class="text-button" data-close-modal>关闭</button>'
  });
}

function settingsModal(state) {
  const settings = state.settings;
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
        ${settingsField("工作目录", "workingDir", settings.workingDir || "")}
        <div class="settings-update">
          <div class="settings-update__row">
            <span class="settings-update__version">当前版本 v${escapeHtml(state.version || "")}</span>
            <button type="button" class="text-button text-button--accent" data-check-update>检查更新</button>
          </div>
          <div class="settings-update__status" data-update-status aria-live="polite"></div>
        </div>
      </div>
    `,
    actions: `
      <button class="text-button" data-close-modal>取消</button>
    `
  });
}

function updateAvailableModal(payload) {
  return modalFrame({
    tone: "success",
    title: "发现新版本",
    description: `新版本 v${escapeHtml(payload.latestVersion)} 已可用。`,
    body: `
      <div class="update-notes">${payload.releaseNotes ? escapeHtml(payload.releaseNotes) : "请前往 Release 页面查看更新说明。"}</div>
    `,
    actions: `
      <button class="text-button" data-skip-update>跳过（7 天内不再提醒）</button>
      <button class="text-button text-button--accent" data-install-update>立即更新</button>
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
  if (modal.kind === "model-editor") return modelEditorModal(modal.payload);
  if (modal.kind === "provider-headers") return headersModal(modal.payload);
  if (modal.kind === "settings") return settingsModal(state);
  if (modal.kind === "update-available") return updateAvailableModal(modal.payload);

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
