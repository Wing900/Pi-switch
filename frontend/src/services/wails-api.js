import { CheckEnvVar, CreateProvider, DeleteProvider, ExecuteLaunchPi, FetchModels, GetAppState, ImportModels, LaunchPi, SetDefaultModel, TestConnection, UpdateProvider, UpdateSettings } from "../../wailsjs/go/main/App";
import { Quit, WindowMinimise, WindowToggleMaximise } from "../../wailsjs/runtime/runtime";

export class WailsApi {
  async getAppState() {
    return GetAppState();
  }

  async listProviders() {
    const state = await GetAppState();
    return state.providers;
  }

  async createProvider(input) {
    return CreateProvider(input);
  }

  async updateProvider(id, input) {
    return UpdateProvider(id, input);
  }

  async deleteProvider(id) {
    return DeleteProvider(id);
  }

  async testConnection(id) {
    return TestConnection(id);
  }

  async fetchModels(id) {
    return FetchModels(id);
  }

  async importModels(providerId, models) {
    return ImportModels(providerId, models);
  }

  async setDefaultModel(providerId, modelId) {
    return SetDefaultModel(providerId, modelId);
  }

  async launchPi(providerId, modelId) {
    return LaunchPi(providerId, modelId);
  }

  async executeLaunchPi(providerId, modelId) {
    return ExecuteLaunchPi(providerId, modelId);
  }

  async updateSettings(settings) {
    return UpdateSettings(settings);
  }

  async checkEnvVar(name) {
    return CheckEnvVar(name);
  }

  minimiseWindow() {
    WindowMinimise();
  }

  closeWindow() {
    Quit();
  }

  toggleMaximiseWindow() {
    WindowToggleMaximise();
  }
}
