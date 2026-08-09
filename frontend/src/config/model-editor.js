import { COMPAT_BOOLEAN_FIELDS, COMPAT_ENUM_FIELDS, THINKING_LEVELS } from "./model-fields.js";

const COST_FIELDS = ["input", "output", "cacheRead", "cacheWrite"];
const COST_TIER_FIELDS = ["inputTokensAbove", ...COST_FIELDS];
const COMPAT_JSON_FIELDS = [
  "chatTemplateKwargs",
  "chatTemplateArgs",
  "openRouterRouting",
  "vercelGatewayRouting"
];

function objectEntries(value) {
  return value && typeof value === "object" && !Array.isArray(value)
    ? Object.entries(value)
    : [];
}

function parseJson(rawValue, label, fallback = undefined) {
  const value = String(rawValue ?? "").trim();
  if (!value) return fallback;
  try {
    return JSON.parse(value);
  } catch {
    throw new Error(`${label} 不是合法 JSON`);
  }
}

function parseObjectJson(rawValue, label) {
  const value = parseJson(rawValue, label);
  if (value === undefined) return undefined;
  if (!value || Array.isArray(value) || typeof value !== "object") {
    throw new Error(`${label} 必须是 JSON 对象`);
  }
  return value;
}

function readNumber(rawValue, label, { integer = false, positive = false } = {}) {
  const value = String(rawValue ?? "").trim();
  if (!value) return undefined;
  const number = Number(value);
  const invalid = !Number.isFinite(number)
    || (positive ? number <= 0 : number < 0)
    || (integer && !Number.isInteger(number));
  if (invalid) {
    throw new Error(`${label} 必须是${positive ? "正" : "非负"}${integer ? "整数" : "数字"}`);
  }
  return number;
}

function validateHeaders(value, label) {
  if (value === undefined) return undefined;
  for (const [name, headerValue] of Object.entries(value)) {
    if (typeof headerValue !== "string") {
      throw new Error(`${label} 的 ${name} 值必须是字符串`);
    }
  }
  return value;
}

function validateCostTiers(value) {
  if (value === undefined) return undefined;
  if (!Array.isArray(value)) throw new Error("cost.tiers 必须是 JSON 数组");

  value.forEach((tier, index) => {
    if (!tier || Array.isArray(tier) || typeof tier !== "object") {
      throw new Error(`cost.tiers[${index}] 必须是 JSON 对象`);
    }
    for (const field of COST_TIER_FIELDS) {
      if (typeof tier[field] !== "number" || !Number.isFinite(tier[field]) || tier[field] < 0) {
        throw new Error(`cost.tiers[${index}].${field} 必须是非负数字`);
      }
    }
  });
  return value;
}

function readThinkingLevelMap(originalModel, value) {
  const thinkingLevelMap = Object.fromEntries(
    objectEntries(originalModel.thinkingLevelMap)
      .filter(([level]) => !THINKING_LEVELS.some(([knownLevel]) => knownLevel === level))
  );

  for (const [level] of THINKING_LEVELS) {
    const mode = value(`modelThinkingMode_${level}`);
    const providerValue = value(`modelThinkingValue_${level}`);
    if (mode === "disabled") thinkingLevelMap[level] = null;
    if (mode === "custom") {
      if (!providerValue) throw new Error(`${level} 思考级别需要填写服务商值`);
      thinkingLevelMap[level] = providerValue;
    }
  }
  return Object.keys(thinkingLevelMap).length ? thinkingLevelMap : undefined;
}

function readCost(originalModel, value) {
  const cost = Object.fromEntries(
    objectEntries(originalModel.cost)
      .filter(([field]) => ![...COST_FIELDS, "tiers"].includes(field))
  );

  for (const field of COST_FIELDS) {
    const fieldValue = readNumber(value(`modelCost_${field}`), `cost.${field}`);
    if (fieldValue !== undefined) cost[field] = fieldValue;
  }

  const tiers = validateCostTiers(parseJson(value("modelCostTiers"), "cost.tiers"));
  if (tiers !== undefined) cost.tiers = tiers;
  return Object.keys(cost).length ? cost : undefined;
}

function readCompat(originalModel, value) {
  const compat = {};

  for (const [name] of COMPAT_BOOLEAN_FIELDS) {
    const fieldValue = value(`modelCompat_${name}`);
    if (fieldValue && fieldValue !== "inherit") compat[name] = fieldValue === "true";
  }
  for (const [name] of COMPAT_ENUM_FIELDS) {
    const fieldValue = value(`modelCompat_${name}`);
    if (fieldValue && fieldValue !== "inherit" && fieldValue !== "__custom__") {
      compat[name] = fieldValue;
    }
    if (fieldValue === "__custom__") {
      const originalValue = originalModel.compat?.[name];
      if (originalValue !== undefined) compat[name] = originalValue;
    }
  }
  for (const field of COMPAT_JSON_FIELDS) {
    const fieldValue = parseObjectJson(value(`modelCompat_${field}`), field);
    if (fieldValue !== undefined) compat[field] = fieldValue;
  }
  return Object.keys(compat).length ? compat : undefined;
}

/**
 * Converts model-editor form values into the exact model document to persist.
 * The caller supplies tiny form adapters; validation and merge precedence stay
 * inside this module so browser actions and tests exercise the same seam.
 */
export function readModelDraft({
  originalModel = {},
  readValue = () => "",
  readChecked = () => false,
  readCheckedValues = () => []
}) {
  const value = (name) => String(readValue(name) ?? "").trim();
  const id = value("modelId");
  if (!id) throw new Error("模型 id 不能为空");

  const contextWindow = readNumber(value("modelContextWindow"), "上下文窗口", { integer: true, positive: true });
  const maxTokens = readNumber(value("modelMaxTokens"), "最大输出", { integer: true, positive: true });
  const thinkingLevelMap = readThinkingLevelMap(originalModel, value);
  const cost = readCost(originalModel, value);
  const compat = readCompat(originalModel, value);
  const samplingParams = parseObjectJson(value("modelSamplingParams"), "samplingParams");
  const headers = validateHeaders(parseObjectJson(value("modelHeaders"), "headers"), "headers");
  const compatExtraFieldsJson = value("modelCompatExtraJson");
  parseObjectJson(compatExtraFieldsJson, "其他 compat 字段");
  const extraFieldsJson = value("modelExtraJson");
  const extra = parseJson(extraFieldsJson, "其他字段", {});
  if (!extra || Array.isArray(extra) || typeof extra !== "object") {
    throw new Error("其他字段必须是 JSON 对象");
  }

  const inputs = readCheckedValues("modelInput");
  const draft = {};
  Object.assign(draft, {
    id,
    ...(Object.prototype.hasOwnProperty.call(originalModel, "api") ? { api: originalModel.api } : {}),
    ...(Object.prototype.hasOwnProperty.call(originalModel, "baseUrl") ? { baseUrl: originalModel.baseUrl } : {}),
    name: value("modelName") || id,
    reasoning: !!readChecked("modelReasoning"),
    ...(contextWindow === undefined ? {} : { contextWindow }),
    ...(maxTokens === undefined ? {} : { maxTokens }),
    input: inputs.length ? inputs : ["text"],
    ...(thinkingLevelMap === undefined ? {} : { thinkingLevelMap }),
    ...(cost === undefined ? {} : { cost }),
    ...(samplingParams === undefined ? {} : { samplingParams }),
    ...(headers === undefined ? {} : { headers }),
    ...(compat === undefined ? {} : { compat }),
    ...(compatExtraFieldsJson ? { compatExtraFieldsJson } : {}),
    ...(extraFieldsJson ? { extraFieldsJson } : {}),
    ...(originalModel.revision ? { revision: originalModel.revision } : {}),
    replaceDocument: true,
    ...(originalModel.id && originalModel.id !== id ? { originalId: originalModel.id } : {})
  });

  return draft;
}
