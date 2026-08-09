import assert from "node:assert/strict";
import test from "node:test";

import { modelExtraFields, readModelDraft } from "./model-editor.js";

const levels = ["off", "minimal", "low", "medium", "high", "xhigh", "max"];
const compatEnums = ["maxTokensField", "thinkingFormat", "cacheControlFormat", "deferredToolsMode", "sessionAffinityFormat"];

function form(overrides = {}) {
  const values = {
    modelId: "demo",
    modelName: "Demo",
    modelContextWindow: "128000",
    modelMaxTokens: "16384",
    modelCostTiers: "",
    modelSamplingParams: "",
    modelHeaders: "",
    modelExtraJson: "",
    ...Object.fromEntries(levels.map((level) => [`modelThinkingMode_${level}`, "inherit"])),
    ...Object.fromEntries(compatEnums.map((field) => [`modelCompat_${field}`, "inherit"])),
    ...overrides.values
  };
  const checked = { modelReasoning: false, ...overrides.checked };
  const checkedValues = { modelInput: ["text"], ...overrides.checkedValues };
  return {
    originalModel: overrides.originalModel || {},
    readValue: (name) => values[name] ?? "",
    readChecked: (name) => checked[name] ?? false,
    readCheckedValues: (name) => checkedValues[name] ?? []
  };
}

function errorMessage(callback) {
  assert.throws(callback, (error) => {
    assert.ok(error instanceof Error);
    return true;
  });
  try {
    callback();
  } catch (error) {
    return error.message;
  }
  return "";
}

test("builds a complete model draft and keeps hidden API fields", () => {
  const draft = readModelDraft(form({
    originalModel: { id: "old-id", api: "openai-completions", baseUrl: "https://model.example/v1" },
    values: {
      modelId: " renamed ",
      modelName: " Renamed ",
      modelContextWindow: "200000",
      modelMaxTokens: "32000",
      modelThinkingMode_high: "custom",
      modelThinkingValue_high: "high",
      modelCost_input: "1.25",
      modelSamplingParams: '{"temperature":0.7}',
      modelHeaders: '{"x-model-route":"fast"}',
      modelExtraJson: '{"vendorOption":{"fast":true}}'
    },
    checked: { modelReasoning: true },
    checkedValues: { modelInput: ["text", "image"] }
  }));

  assert.deepEqual(draft, {
    vendorOption: { fast: true },
    id: "renamed",
    api: "openai-completions",
    baseUrl: "https://model.example/v1",
    name: "Renamed",
    reasoning: true,
    contextWindow: 200000,
    maxTokens: 32000,
    input: ["text", "image"],
    thinkingLevelMap: { high: "high" },
    cost: { input: 1.25 },
    samplingParams: { temperature: 0.7 },
    headers: { "x-model-route": "fast" },
    __piSwitchReplaceDocument: true,
    __piSwitchOriginalId: "old-id"
  });
});

test("preserves unknown nested fields and unknown compat enum values", () => {
  const draft = readModelDraft(form({
    originalModel: {
      thinkingLevelMap: { future: "ultra" },
      cost: { input: 1, vendorRate: 2 },
      compat: { thinkingFormat: "future-format", vendorCompat: true }
    },
    values: {
      modelCost_input: "1",
      modelCompat_thinkingFormat: "__custom__"
    }
  }));

  assert.deepEqual(draft.thinkingLevelMap, { future: "ultra" });
  assert.deepEqual(draft.cost, { vendorRate: 2, input: 1 });
  assert.deepEqual(draft.compat, { vendorCompat: true, thinkingFormat: "future-format" });
});

test("does not expand a partial cost object to four zero fields", () => {
  const draft = readModelDraft(form({
    originalModel: { cost: { input: 1 } },
    values: { modelCost_input: "1" }
  }));
  assert.deepEqual(draft.cost, { input: 1 });
});

test("validates JSON object fields and header values", () => {
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelSamplingParams: "[]" } }))),
    "samplingParams 必须是 JSON 对象"
  );
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelHeaders: '{"x-route":1}' } }))),
    "headers 的 x-route 值必须是字符串"
  );
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelCompat_openRouterRouting: '"invalid"' } }))),
    "openRouterRouting 必须是 JSON 对象"
  );
});

test("validates integers, JSON syntax, and complete cost tiers", () => {
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelContextWindow: "12.5" } }))),
    "上下文窗口 必须是非负整数"
  );
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelExtraJson: "{" } }))),
    "其他字段 不是合法 JSON"
  );
  assert.equal(
    errorMessage(() => readModelDraft(form({
      values: { modelCostTiers: '[{"inputTokensAbove":100,"input":1}]' }
    }))),
    "cost.tiers[0].output 必须是非负数字"
  );
});

test("requires a provider value for custom thinking levels", () => {
  assert.equal(
    errorMessage(() => readModelDraft(form({ values: { modelThinkingMode_high: "custom" } }))),
    "high 思考级别需要填写服务商值"
  );
});

test("filters known and UI-only fields from the extra JSON editor", () => {
  assert.deepEqual(modelExtraFields({
    id: "demo",
    name: "Demo",
    compat: {},
    selected: true,
    extraFields: { legacy: true },
    vendorOption: 1
  }), {
    legacy: true,
    vendorOption: 1
  });
});
