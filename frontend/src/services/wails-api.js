import { CheckEnvVar, CheckForUpdate, CreateProvider, DeleteProvider, ExecuteLaunchPi, FetchModels, GetAppState, ImportModels, InstallUpdate, LaunchPi, MarkUpdateChecked, ReplaceModels, SetDefaultModel, TestConnection, UpdateProvider, UpdateSettings } from "../../wailsjs/go/main/App";
import { Quit, WindowMinimise, WindowSetDarkTheme, WindowSetLightTheme, WindowToggleMaximise, EventsOn } from "../../wailsjs/runtime/runtime";

export class WailsApi {
  onConfigChanged(callback) {
    return EventsOn("pi:config-changed", () => callback());
  }

  onUpdateAvailable(callback) {
    return EventsOn("pi:update-available", (result) => callback(result));
  }

  applyWindowTheme(isDark) {
    if (isDark) {
      WindowSetDarkTheme();
      return;
    }
    WindowSetLightTheme();
  }

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

  async replaceModels(providerId, models, expectedRevision = "") {
    return ReplaceModels(providerId, models, expectedRevision);
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
    await UpdateSettings(settings);
    this.applyWindowTheme(settings.darkMode);
  }

  async checkForUpdate() {
    return CheckForUpdate();
  }

  async markUpdateChecked() {
    return MarkUpdateChecked();
  }

  async installUpdate() {
    return InstallUpdate();
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
