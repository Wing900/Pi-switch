import assert from "node:assert/strict";
import test from "node:test";

import { createProviderActions } from "./provider-actions.js";
import { createStore } from "../state/store.js";

function modelFormRoot(overrides = {}) {
  const fields = {
    modelId: { value: "demo" },
    modelName: { value: "Demo" },
    modelContextWindow: { value: "128000" },
    modelMaxTokens: { value: "16384" },
    modelReasoning: { checked: false },
    modelCostTiers: { value: "" },
    modelSamplingParams: { value: "" },
    modelHeaders: { value: "" },
    modelExtraJson: { value: "" },
    ...overrides.fields
  };
  return {
    querySelector(selector) {
      const match = selector.match(/^\[name="([^"]+)"\]$/);
      return match ? fields[match[1]] ?? null : null;
    },
    querySelectorAll(selector) {
      if (selector === 'input[name="modelInput"]:checked') {
        return (overrides.inputTypes ?? ["text"]).map((value) => ({ value }));
      }
      return [];
    }
  };
}

function actionsFor({ root, state, api = {} }) {
  const errors = [];
  const store = createStore(state);
  const actions = createProviderActions({
    root,
    api,
    store,
    providerForm: { commit: async () => true },
    feedback: {
      showError(title, error) {
        errors.push({ title, error });
      },
      showLoading() {}
    }
  });
  return { actions, store, errors };
}

test("renames selected and default model without a second default call", async () => {
  let replaced;
  const root = modelFormRoot({
    fields: {
      modelId: { value: "new-id" },
      modelName: { value: "Renamed" }
    }
  });
  const { actions, store, errors } = actionsFor({
    root,
    state: {
      selectedProviderId: "p",
      defaultProviderId: "p",
      defaultModelId: "old-id",
      modal: {
        kind: "model-editor",
        payload: {
          providerId: "p",
          originalModelId: "old-id",
          model: { id: "old-id", name: "Old", reasoning: false }
        }
      },
      providers: [{ id: "p", selectedModelId: "old-id", models: [{ id: "old-id", name: "Old" }] }]
    },
    api: {
      async replaceModels(providerId, models) {
        assert.equal(providerId, "p");
        replaced = models;
      },
      async setDefaultModel() {
        assert.fail("saveModelEditor must not issue a second default-model write");
      }
    }
  });

  await actions.saveModelEditor();

  assert.equal(errors.length, 0);
  assert.equal(replaced[0].__piSwitchReplaceDocument, true);
  assert.equal(replaced[0].__piSwitchOriginalId, "old-id");
  assert.equal(store.getState().defaultModelId, "new-id");
  assert.equal(store.getState().providers[0].selectedModelId, "new-id");
  assert.equal(store.getState().providers[0].models[0].__piSwitchReplaceDocument, undefined);
  assert.equal(store.getState().providers[0].models[0].__piSwitchOriginalId, undefined);
});

test("rejects a duplicate model id before writing", async () => {
  let writes = 0;
  const { actions, errors } = actionsFor({
    root: modelFormRoot({ fields: { modelId: { value: "taken" } } }),
    state: {
      selectedProviderId: "p",
      defaultProviderId: "",
      defaultModelId: "",
      modal: {
        kind: "model-editor",
        payload: {
          providerId: "p",
          originalModelId: "old-id",
          model: { id: "old-id", name: "Old" }
        }
      },
      providers: [{
        id: "p",
        selectedModelId: "old-id",
        models: [{ id: "old-id", name: "Old" }, { id: "taken", name: "Taken" }]
      }]
    },
    api: { async replaceModels() { writes += 1; } }
  });

  await actions.saveModelEditor();

  assert.equal(writes, 0);
  assert.equal(errors[0].error.message, "模型 ID 已存在：taken");
});

test("deleting a default model clears the frontend default state", async () => {
  const { actions, store } = actionsFor({
    root: modelFormRoot(),
    state: {
      selectedProviderId: "p",
      defaultProviderId: "p",
      defaultModelId: "remove",
      modal: null,
      providers: [{
        id: "p",
        selectedModelId: "remove",
        models: [{ id: "remove", name: "Remove" }, { id: "keep", name: "Keep" }]
      }]
    },
    api: { async replaceModels() {} }
  });

  await actions.deleteModel("remove");

  assert.equal(store.getState().defaultModelId, "");
  assert.equal(store.getState().providers[0].selectedModelId, "keep");
});

test("bulk import keeps existing and fetched unknown fields in the transport envelope", async () => {
  let replaced;
  const checkboxes = [{ dataset: { modelId: "demo" } }];
  const root = {
    querySelector() {
      return null;
    },
    querySelectorAll(selector) {
      if (selector === "[data-model-id]:checked") return checkboxes;
      if (selector === "[data-cw-model]") return [{ dataset: { cwModel: "demo" }, value: "256" }];
      return [];
    }
  };
  const { actions, errors } = actionsFor({
    root,
    state: {
      selectedProviderId: "p",
      defaultProviderId: "",
      defaultModelId: "",
      modal: {
        kind: "fetch-models",
        payload: {
          providerId: "p",
          models: [{
            id: "demo",
            name: "Fetched",
            api: undefined,
            baseUrl: undefined,
            selected: true,
            extraFields: { fetchedFlag: 2 }
          }]
        }
      },
      providers: [{
        id: "p",
        selectedModelId: "demo",
        models: [{
          id: "demo",
          name: "Existing",
          api: "openai-completions",
          baseUrl: "https://model.example/v1",
          extraFields: { vendorFlag: 1 }
        }]
      }]
    },
    api: { async replaceModels(_providerId, models) { replaced = models; } }
  });

  await actions.importModels();

  assert.equal(errors.length, 0);
  assert.deepEqual(replaced[0].extraFields, { vendorFlag: 1, fetchedFlag: 2 });
  assert.equal(replaced[0].api, "openai-completions");
  assert.equal(replaced[0].baseUrl, "https://model.example/v1");
  assert.equal(replaced[0].selected, undefined);
  assert.equal(replaced[0].contextWindow, 256000);
});
