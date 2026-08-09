import assert from "node:assert/strict";
import test from "node:test";

import { createProviderActions, modelsMatch } from "./provider-actions.js";
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
    modelCompatExtraJson: { value: "" },
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

test("stores the persisted provider returned after preset creation", async () => {
  const state = {
    selectedProviderId: "",
    defaultProviderId: "",
    defaultModelId: "",
    modal: { kind: "add-provider" },
    providers: []
  };
  const store = createStore(state);
  const actions = createProviderActions({
    root: modelFormRoot(),
    store,
    providerForm: { async commit() { return true; } },
    feedback: { showError() {}, showLoading() {} },
    api: {
      async createProvider(input) {
        return { ...input, modelsRevision: "list-revision", models: input.models.map((model) => ({ ...model, revision: "model-revision" })) };
      }
    }
  });

  await actions.createFromPreset("openai");

  assert.equal(store.getState().providers[0].modelsRevision, "list-revision");
  assert.equal(store.getState().providers[0].models[0].revision, "model-revision");
});
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
      providers: [{ id: "p", selectedModelId: "old-id", models: [{ id: "old-id", name: "Old", reasoning: false }] }]
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
  assert.equal(replaced[0].replaceDocument, true);
  assert.equal(replaced[0].originalId, "old-id");
  assert.equal(store.getState().defaultModelId, "new-id");
  assert.equal(store.getState().providers[0].selectedModelId, "new-id");
  assert.equal(store.getState().providers[0].models[0].replaceDocument, undefined);
  assert.equal(store.getState().providers[0].models[0].originalId, undefined);
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

test("opening the editor commits pending provider form changes", async () => {
  let commits = 0;
  const state = {
    selectedProviderId: "p",
    defaultProviderId: "",
    defaultModelId: "",
    modal: null,
    providers: [{ id: "p", selectedModelId: "demo", models: [{ id: "demo", name: "Demo" }] }]
  };
  const store = createStore(state);
  const actions = createProviderActions({
    root: modelFormRoot(),
    api: {},
    store,
    providerForm: { async commit() { commits += 1; return true; } },
    feedback: { showError() {}, showLoading() {} }
  });

  await actions.openModelEditor("demo");

  assert.equal(commits, 1);
  assert.equal(store.getState().modal?.kind, "model-editor");
});

test("stale editor refuses to overwrite an externally changed model", async () => {
  let writes = 0;
  const { actions, errors } = actionsFor({
    root: modelFormRoot(),
    state: {
      selectedProviderId: "p",
      defaultProviderId: "",
      defaultModelId: "",
      modal: {
        kind: "model-editor",
        payload: {
          providerId: "p",
          originalModelId: "demo",
          model: { id: "demo", name: "Before", contextWindow: 128000, maxTokens: 16384 }
        }
      },
      providers: [{
        id: "p",
        selectedModelId: "demo",
        models: [{ id: "demo", name: "External", contextWindow: 200000, maxTokens: 32000 }]
      }]
    },
    api: { async replaceModels() { writes += 1; } }
  });

  await actions.saveModelEditor();

  assert.equal(writes, 0);
  assert.match(errors[0].error.message, /外部修改|重新打开/);
});

test("editing a model preserves its list position", async () => {
  let ids;
  const root = modelFormRoot({ fields: { modelId: { value: "b" }, modelName: { value: "B edited" } } });
  const { actions } = actionsFor({
    root,
    state: {
      selectedProviderId: "p",
      defaultProviderId: "",
      defaultModelId: "",
      modal: {
        kind: "model-editor",
        payload: {
          providerId: "p",
          originalModelId: "b",
          model: { id: "b", name: "B", contextWindow: 128000, maxTokens: 16384 }
        }
      },
      providers: [{
        id: "p",
        selectedModelId: "b",
        models: [
          { id: "a", name: "A" },
          { id: "b", name: "B", contextWindow: 128000, maxTokens: 16384 },
          { id: "c", name: "C" }
        ]
      }]
    },
    api: { async replaceModels(_providerId, models) { ids = models.map((model) => model.id); } }
  });

  await actions.saveModelEditor();

  assert.deepEqual(ids, ["a", "b", "c"]);
});

test("stores the persisted model and list revisions returned by the backend", async () => {
  let expectedRevision;
  const root = modelFormRoot({ fields: { modelId: { value: "demo" }, modelName: { value: "After" } } });
  const { actions, store } = actionsFor({
    root,
    state: {
      selectedProviderId: "p",
      defaultProviderId: "",
      defaultModelId: "",
      modal: {
        kind: "model-editor",
        payload: {
          providerId: "p",
          originalModelId: "demo",
          modelsRevision: "list-old",
          model: { id: "demo", name: "Before", reasoning: false, contextWindow: 128000, maxTokens: 16384, revision: "old" }
        }
      },
      providers: [{
        id: "p",
        selectedModelId: "demo",
        modelsRevision: "list-old",
        models: [{ id: "demo", name: "Before", reasoning: false, contextWindow: 128000, maxTokens: 16384, revision: "old" }]
      }]
    },
    api: {
      async replaceModels(_providerId, _models, revision) {
        expectedRevision = revision;
        return {
          models: [{ id: "demo", name: "After", reasoning: false, contextWindow: 128000, maxTokens: 16384, revision: "fresh" }],
          revision: "list-fresh"
        };
      }
    }
  });

  await actions.saveModelEditor();

  assert.equal(expectedRevision, "list-old");
  assert.equal(store.getState().providers[0].models[0].revision, "fresh");
  assert.equal(store.getState().providers[0].modelsRevision, "list-fresh");
  assert.equal(store.getState().providers[0].models[0].replaceDocument, undefined);
});
test("model equality ignores object key order but detects value changes", () => {
  assert.equal(modelsMatch(
    { id: "m", compat: { future: true, nested: { b: 2, a: 1 } } },
    { compat: { nested: { a: 1, b: 2 }, future: true }, id: "m" }
  ), true);
  assert.equal(modelsMatch({ id: "m", maxTokens: 10 }, { id: "m", maxTokens: 20 }), false);
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
            extraFieldsJson: '{"fetchedFlag":2}'
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
          extraFieldsJson: '{"vendorFlag":1}'
        }]
      }]
    },
    api: { async replaceModels(_providerId, models) { replaced = models; } }
  });

  await actions.importModels();

  assert.equal(errors.length, 0);
  assert.equal(replaced[0].extraFieldsJson, '{"fetchedFlag":2}');
  assert.equal(replaced[0].api, "openai-completions");
  assert.equal(replaced[0].baseUrl, "https://model.example/v1");
  assert.equal(replaced[0].selected, undefined);
  assert.equal(replaced[0].contextWindow, 256000);
});
