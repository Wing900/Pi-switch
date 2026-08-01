package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"piswitch/internal/config"
	"piswitch/internal/paths"
	"piswitch/internal/pi"
	"piswitch/internal/provider"
	"piswitch/internal/system"
	"piswitch/internal/updater"
)

const appVersion = "0.0.0.12"

const configChangedEvent = "pi:config-changed"
const updateAvailableEvent = "pi:update-available"

type App struct {
	ctx         context.Context
	service     *config.Service
	watcher     *fsnotify.Watcher
	watcherStop chan struct{}
	writeMu     sync.Mutex
	lastSelfWrite time.Time
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service = config.NewService(paths.DefaultPaths())
	a.startConfigWatcher()
	a.startBackgroundUpdateCheck()
}

func (a *App) shutdown(ctx context.Context) {
	a.stopConfigWatcher()
}

// markSelfWrite 在本应用主动写盘前调用，用于让 watcher 忽略自身触发的变更。
func (a *App) markSelfWrite() {
	a.writeMu.Lock()
	a.lastSelfWrite = time.Now()
	a.writeMu.Unlock()
}

func (a *App) recentSelfWrite() bool {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	return time.Since(a.lastSelfWrite) < 800*time.Millisecond
}

func (a *App) startConfigWatcher() {
	cfg, err := a.service.Load()
	if err != nil {
		return
	}
	targets := make(map[string]struct{}, 2)
	if cfg.Settings.PiSwitchConfigPath != "" {
		targets[cfg.Settings.PiSwitchConfigPath] = struct{}{}
	}
	if cfg.Settings.PiSettingsPath != "" {
		targets[cfg.Settings.PiSettingsPath] = struct{}{}
	}
	if len(targets) == 0 {
		return
	}

	dirs := make(map[string]struct{}, len(targets))
	for target := range targets {
		dirs[filepath.Dir(target)] = struct{}{}
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	for dir := range dirs {
		if err := w.Add(dir); err != nil {
			_ = w.Close()
			return
		}
	}
	a.watcher = w
	a.watcherStop = make(chan struct{})

	go func() {
		var debounce *time.Timer
		pending := false
		fire := func() {
			if !pending {
				return
			}
			pending = false
			if a.recentSelfWrite() {
				return
			}
			runtime.EventsEmit(a.ctx, configChangedEvent)
		}
		for {
			select {
			case <-a.watcherStop:
				if debounce != nil {
					debounce.Stop()
				}
				return
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				if _, ok := targets[event.Name]; !ok {
					continue
				}
				if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
					continue
				}
				pending = true
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(300*time.Millisecond, fire)
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				_ = err
			}
		}
	}()
}

func (a *App) stopConfigWatcher() {
	if a.watcherStop != nil {
		close(a.watcherStop)
		a.watcherStop = nil
	}
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
}

func (a *App) GetAppState() (config.AppState, error) {
	cfg, err := a.service.Load()
	if err != nil {
		return config.AppState{}, err
	}
	piDefaults, _ := pi.ReadDefaults(cfg.Settings.PiSettingsPath)
	selectedProvider := ""
	if len(cfg.Providers) > 0 {
		selectedProvider = cfg.Providers[0].ID
	}

	logs := []string{
		"就绪。",
		"Go/Wails 绑定已接入。",
		"当前版本会把 Provider 配置持久化到 ~/.piswitch/config.json。",
	}

	defaultProviderID := firstNonEmpty(piDefaults.DefaultProvider, cfg.Settings.LastDefaultProviderID, selectedProvider)
	defaultModelID := firstNonEmpty(piDefaults.DefaultModel, cfg.Settings.LastDefaultModelID)

	return config.AppState{
		Version:            appVersion,
		Providers:          cfg.Providers,
		SelectedProviderID: selectedProvider,
		DefaultProviderID:  defaultProviderID,
		DefaultModelID:     defaultModelID,
		Settings:           cfg.Settings,
		Logs:               logs,
	}, nil
}

func (a *App) ListProviders() ([]provider.Config, error) {
	cfg, err := a.service.Load()
	if err != nil {
		return nil, err
	}
	return cfg.Providers, nil
}

func (a *App) CreateProvider(input provider.Config) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	input = provider.Normalize(input)
	if err := config.ValidateProvider(input); err != nil {
		return err
	}
	cfg.UpsertProvider(input, "")
	return a.persistPiState(cfg)
}

func (a *App) UpdateProvider(id string, input provider.Config) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	input = provider.Normalize(input)
	if err := config.ValidateProvider(input); err != nil {
		return err
	}
	cfg.UpsertProvider(input, id)
	if cfg.Settings.LastDefaultProviderID == id {
		cfg.Settings.LastDefaultProviderID = input.ID
	}
	return a.persistPiState(cfg)
}

func (a *App) DeleteProvider(id string) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	cfg.DeleteProvider(id)
	if cfg.Settings.LastDefaultProviderID == id {
		cfg.Settings.LastDefaultProviderID = ""
		cfg.Settings.LastDefaultModelID = ""
	}
	return a.persistPiState(cfg)
}

func (a *App) TestConnection(id string) (provider.ConnectionTestResult, error) {
	cfg, err := a.service.Load()
	if err != nil {
		return provider.ConnectionTestResult{}, err
	}
	current, err := cfg.ProviderByID(id)
	if err != nil {
		return provider.ConnectionTestResult{}, err
	}

	key, envResult := system.ResolveAPIKey(current.APIKeyEnv, current.APIKeyLiteral)
	if current.APIKeyEnv != "" && !envResult.Found {
		return provider.ConnectionTestResult{
			OK:    false,
			Title: "连接测试失败",
			Lines: []string{
				"状态：失败",
				"问题：API Key 环境变量不存在",
				"修复：请先设置系统环境变量并重新打开 Pi Switch",
			},
		}, nil
	}

	models, err := provider.FetchModelsByAPI(current, key)
	if err != nil {
		return provider.ConnectionTestResult{
			OK:    false,
			Title: "连接测试失败",
			Lines: []string{
				"状态：失败",
				"问题：" + err.Error(),
				"修复：请检查 Base URL、代理设置、API 密钥和服务端模型列表接口",
			},
		}, nil
	}

	endpoint := "模型列表接口"
	if current.API == "anthropic-messages" || current.API == "anthropic" {
		endpoint = "/v1/models"
	} else {
		endpoint = "/models"
	}

	return provider.ConnectionTestResult{
		OK:    true,
		Title: "连接测试",
		Lines: []string{
			"状态：正常",
			"Base URL：可访问",
			envResult.Message,
			endpoint + "：可用",
			"检测到模型：" + provider.FormatModelCount(len(models)),
		},
	}, nil
}

func (a *App) FetchModels(id string) ([]provider.ModelInfo, error) {
	cfg, err := a.service.Load()
	if err != nil {
		return nil, err
	}
	current, err := cfg.ProviderByID(id)
	if err != nil {
		return nil, err
	}
	key, envResult := system.ResolveAPIKey(current.APIKeyEnv, current.APIKeyLiteral)
	if current.APIKeyEnv != "" && !envResult.Found {
		return nil, errors.New("环境变量 " + current.APIKeyEnv + " 不存在")
	}
	return provider.FetchModelsByAPI(current, key)
}

func (a *App) ImportModels(providerID string, models []provider.ModelInfo) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	current, err := cfg.ProviderByID(providerID)
	if err != nil {
		return err
	}
	current.Models = provider.MergeModels(current.Models, models)
	current = provider.Normalize(current)
	cfg.UpsertProvider(current, providerID)
	return a.persistPiState(cfg)
}

// ReplaceModels 用给定列表整体替换该 provider 的模型集合（替换语义，未传入的将被删除）。
func (a *App) ReplaceModels(providerID string, models []provider.ModelInfo) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	current, err := cfg.ProviderByID(providerID)
	if err != nil {
		return err
	}
	current.Models = provider.NormalizeModels(models)
	current = provider.Normalize(current)
	cfg.UpsertProvider(current, providerID)
	return a.persistPiState(cfg)
}

func (a *App) SetDefaultModel(providerID string, modelID string) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	current, err := cfg.ProviderByID(providerID)
	if err != nil {
		return err
	}
	current.SelectedModelID = modelID
	cfg.UpsertProvider(current, providerID)
	cfg.Settings.LastDefaultProviderID = providerID
	cfg.Settings.LastDefaultModelID = modelID

	return a.persistPiState(cfg)
}

func (a *App) LaunchPi(providerID string, modelID string) (pi.LaunchPreview, error) {
	cfg, err := a.service.Load()
	if err != nil {
		return pi.LaunchPreview{}, err
	}
	command := pi.BuildCommand(cfg.Settings.PiCommand, providerID, modelID)
	return pi.LaunchPreview{
		Command: command,
		Checklist: []string{
			"Provider 配置存在",
			"默认模型已选定",
			"可在下一版补成真正拉起新终端执行",
		},
	}, nil
}

func (a *App) OpenConfigFolder() error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	target := filepath.Dir(cfg.Settings.PiSwitchConfigPath)
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(target))
	return nil
}

func (a *App) CheckEnvVar(name string) (system.EnvCheckResult, error) {
	return system.CheckEnvVar(name), nil
}

func (a *App) UpdateSettings(input config.AppSettings) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	cfg.Settings = config.NormalizeSettings(input)
	a.markSelfWrite()
	return a.service.Save(cfg)
}

// startBackgroundUpdateCheck 在启动时静默检测更新：距上次检查满 7 天才会拉取
// GitHub，有新版本则向前端广播事件。检测结果（无论有无更新）都会刷新时间戳。
func (a *App) startBackgroundUpdateCheck() {
	cfg, err := a.service.Load()
	if err != nil {
		return
	}
	if !updater.NeedsCheck(cfg.Settings) {
		return
	}
	go func() {
		res, err := updater.CheckLatest(appVersion)
		a.recordUpdateCheck()
		if err != nil || !res.HasUpdate {
			return
		}
		runtime.EventsEmit(a.ctx, updateAvailableEvent, res)
	}()
}

// recordUpdateCheck 刷新 LastUpdateCheckAt，作为“跳过则 7 天不检测”的持久化依据。
func (a *App) recordUpdateCheck() error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	cfg.Settings.LastUpdateCheckAtUnix = time.Now().Unix()
	a.markSelfWrite()
	return a.service.Save(cfg)
}

// CheckForUpdate 手动检查更新（不受 7 天节流限制），并刷新检测时间戳。
func (a *App) CheckForUpdate() (updater.CheckResult, error) {
	res, err := updater.CheckLatest(appVersion)
	_ = a.recordUpdateCheck()
	return res, err
}

// MarkUpdateChecked 记录“本次检测已处理（跳过）”，一周内不再静默提醒。
func (a *App) MarkUpdateChecked() error {
	return a.recordUpdateCheck()
}

// InstallUpdate 下载最新 Windows 更新包并写入延迟替换脚本；调用方随后应退出主进程。
func (a *App) InstallUpdate() error {
	res, err := updater.CheckLatest(appVersion)
	if err != nil {
		return err
	}
	if !res.HasUpdate || res.AssetURL == "" {
		return errors.New("当前已是最新版本，无需更新")
	}
	tmp, err := os.CreateTemp("", "PiSwitch-update-*.exe")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	if err := updater.Download(res.AssetURL, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return updater.Install(tmpPath)
}

func (a *App) ExecuteLaunchPi(providerID string, modelID string) error {
	cfg, err := a.service.Load()
	if err != nil {
		return err
	}
	current, err := cfg.ProviderByID(providerID)
	if err != nil {
		return err
	}
	current.SelectedModelID = modelID
	cfg.UpsertProvider(current, providerID)
	cfg.Settings.LastDefaultProviderID = providerID
	cfg.Settings.LastDefaultModelID = modelID
	if err := a.persistPiState(cfg); err != nil {
		return err
	}
	command := pi.BuildCommand(cfg.Settings.PiCommand, providerID, modelID)
	return pi.OpenCommandInTerminal(command, cfg.Settings.WorkingDir)
}

func (a *App) persistPiState(cfg config.SwitchConfig) error {
	a.markSelfWrite()
	if err := a.service.Save(cfg); err != nil {
		return err
	}

	a.markSelfWrite()
	if err := pi.WriteAllModels(cfg.Settings.PiModelsPath, cfg.Providers); err != nil {
		return err
	}
	a.markSelfWrite()
	return pi.MergeDefaults(cfg.Settings.PiSettingsPath, pi.DefaultSettings{
		DefaultProvider:      cfg.Settings.LastDefaultProviderID,
		DefaultModel:         cfg.Settings.LastDefaultModelID,
		DefaultThinkingLevel: "off",
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func init() {
	_ = os.Setenv("WAILS_SAVE_FILE_OVERWRITE_PROMPT", "false")
}
