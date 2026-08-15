package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"piswitch/internal/config"
	"piswitch/internal/configsync"
	"piswitch/internal/paths"
	"piswitch/internal/pi"
	"piswitch/internal/provider"
	"piswitch/internal/system"
	"piswitch/internal/updater"
	"piswitch/internal/wsl"
)

const appVersion = "0.0.0.14"

const configChangedEvent = "pi:config-changed"
const updateAvailableEvent = "pi:update-available"

type App struct {
	ctx          context.Context
	coordinator  *configsync.Coordinator
	watcher      *fsnotify.Watcher
	watcherStop  chan struct{}
	selfWriteMu  sync.Mutex
	selfWriteMap map[string]selfWriteFingerprint
}

type selfWriteFingerprint struct {
	hash       [sha256.Size]byte
	recordedAt time.Time
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	service := config.NewService(paths.DefaultPaths())
	a.coordinator = configsync.New(service, a.recordSelfWrite)
	a.startConfigWatcher()
	a.startBackgroundUpdateCheck()
}

func (a *App) shutdown(ctx context.Context) {
	a.stopConfigWatcher()
}

// recordSelfWrite records the actual file content so watcher suppression is
// scoped to this exact write. A concurrent external change gets a new hash and
// still reaches the UI.
func (a *App) recordSelfWrite(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	a.selfWriteMu.Lock()
	defer a.selfWriteMu.Unlock()
	if a.selfWriteMap == nil {
		a.selfWriteMap = map[string]selfWriteFingerprint{}
	}
	a.selfWriteMap[filepath.Clean(path)] = selfWriteFingerprint{
		hash:       sha256.Sum256(data),
		recordedAt: time.Now(),
	}
}

func (a *App) isSelfWrite(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	cleanPath := filepath.Clean(path)
	a.selfWriteMu.Lock()
	defer a.selfWriteMu.Unlock()
	recorded, ok := a.selfWriteMap[cleanPath]
	if !ok {
		return false
	}
	if time.Since(recorded.recordedAt) > 3*time.Second {
		delete(a.selfWriteMap, cleanPath)
		return false
	}
	return recorded.hash == sha256.Sum256(data)
}

func (a *App) startConfigWatcher() {
	cfg, err := a.coordinator.Load()
	if err != nil {
		return
	}
	targets := make(map[string]struct{}, 3)
	if cfg.Settings.PiSwitchConfigPath != "" {
		targets[cfg.Settings.PiSwitchConfigPath] = struct{}{}
	}
	if cfg.Settings.PiSettingsPath != "" {
		targets[cfg.Settings.PiSettingsPath] = struct{}{}
	}
	if cfg.Settings.PiModelsPath != "" {
		targets[cfg.Settings.PiModelsPath] = struct{}{}
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
	watchedDirs := 0
	for dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			continue
		}
		if err := w.Add(dir); err != nil {
			continue
		}
		watchedDirs++
	}
	if watchedDirs == 0 {
		_ = w.Close()
		return
	}
	a.watcher = w
	a.watcherStop = make(chan struct{})

	go func() {
		var debounce *time.Timer
		var debounceC <-chan time.Time
		pendingPaths := map[string]struct{}{}
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
				if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 && event.Op&fsnotify.Remove == 0 && event.Op&fsnotify.Rename == 0 {
					continue
				}
				pendingPaths[event.Name] = struct{}{}
				if debounce == nil {
					debounce = time.NewTimer(300 * time.Millisecond)
				} else {
					if !debounce.Stop() {
						select {
						case <-debounce.C:
						default:
						}
					}
					debounce.Reset(300 * time.Millisecond)
				}
				debounceC = debounce.C
			case <-debounceC:
				shouldReload := false
				for path := range pendingPaths {
					if !a.isSelfWrite(path) {
						shouldReload = true
						break
					}
				}
				pendingPaths = map[string]struct{}{}
				debounceC = nil
				if shouldReload {
					runtime.EventsEmit(a.ctx, configChangedEvent)
				}
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
	cfg, err := a.coordinator.Load()
	if err != nil {
		return config.AppState{}, err
	}
	piDefaults, err := pi.ReadDefaults(cfg.Settings.PiSettingsPath)
	if err != nil {
		return config.AppState{}, err
	}
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

	providerTransports, err := provider.ConfigTransports(cfg.Providers)
	if err != nil {
		return config.AppState{}, err
	}

	return config.AppState{
		Version:            appVersion,
		Providers:          providerTransports,
		SelectedProviderID: selectedProvider,
		DefaultProviderID:  defaultProviderID,
		DefaultModelID:     defaultModelID,
		Settings:           cfg.Settings,
		Logs:               logs,
	}, nil
}

func (a *App) GetWSLDetection() (wsl.Detection, error) {
	return wsl.Detect()
}

func (a *App) GetWSLPiDetection(distro string) (wsl.PiDetection, error) {
	return wsl.DetectPi(distro)
}

func (a *App) ListProviders() ([]provider.ConfigTransport, error) {
	cfg, err := a.coordinator.Load()
	if err != nil {
		return nil, err
	}
	return provider.ConfigTransports(cfg.Providers)
}

func (a *App) CreateProvider(input provider.ConfigTransport) (provider.ConfigTransport, error) {
	converted, err := input.Config()
	if err != nil {
		return provider.ConfigTransport{}, err
	}
	if err := a.coordinator.UpsertProvider("", converted); err != nil {
		return provider.ConfigTransport{}, err
	}
	cfg, err := a.coordinator.Load()
	if err != nil {
		return provider.ConfigTransport{}, err
	}
	persisted, err := cfg.ProviderByID(converted.ID)
	if err != nil {
		return provider.ConfigTransport{}, err
	}
	return provider.NewConfigTransport(persisted)
}

func (a *App) UpdateProvider(id string, input provider.ConfigTransport) error {
	converted, err := input.Config()
	if err != nil {
		return err
	}
	return a.coordinator.UpsertProvider(id, converted)
}

func (a *App) DeleteProvider(id string) error {
	return a.coordinator.DeleteProvider(id)
}

func (a *App) TestConnection(id string) (provider.ConnectionTestResult, error) {
	cfg, err := a.coordinator.Load()
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

func (a *App) FetchModels(id string) ([]provider.ModelTransport, error) {
	cfg, err := a.coordinator.Load()
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
	models, err := provider.FetchModelsByAPI(current, key)
	if err != nil {
		return nil, err
	}
	return provider.ModelTransports(models)
}

func (a *App) ImportModels(providerID string, models []provider.ModelTransport) error {
	converted, err := provider.ModelsFromTransport(models)
	if err != nil {
		return err
	}
	return a.coordinator.MergeModels(providerID, converted)
}

// ReplaceModels 用给定列表整体替换该 provider 的模型集合（替换语义，未传入的将被删除）。
func (a *App) ReplaceModels(providerID string, models []provider.ModelTransport, expectedRevision string) (provider.ModelListTransport, error) {
	converted, err := provider.ModelsFromTransport(models)
	if err != nil {
		return provider.ModelListTransport{}, err
	}
	replaced, err := a.coordinator.ReplaceModels(providerID, converted, expectedRevision)
	if err != nil {
		return provider.ModelListTransport{}, err
	}
	return provider.NewModelListTransport(replaced)
}

func (a *App) SetDefaultModel(providerID string, modelID string) error {
	return a.coordinator.SetDefault(providerID, modelID)
}

func (a *App) LaunchPi(providerID string, modelID string) (pi.LaunchPreview, error) {
	cfg, err := a.coordinator.Load()
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
	cfg, err := a.coordinator.Load()
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
	if err := a.coordinator.UpdateSettings(input); err != nil {
		return err
	}
	a.stopConfigWatcher()
	a.startConfigWatcher()
	return nil
}

// startBackgroundUpdateCheck 在启动时静默检测更新：距上次检查满 7 天才会拉取
// GitHub，有新版本则向前端广播事件。检测结果（无论有无更新）都会刷新时间戳。
func (a *App) startBackgroundUpdateCheck() {
	cfg, err := a.coordinator.Load()
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
	return a.coordinator.RecordUpdateCheck()
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
	cfg, err := a.coordinator.Load()
	if err != nil {
		return err
	}
	current, err := cfg.ProviderByID(providerID)
	if err != nil {
		return err
	}
	if err := provider.EnsureModel(current, modelID); err != nil {
		return err
	}
	command := pi.BuildCommand(cfg.Settings.PiCommand, providerID, modelID)
	return pi.OpenCommandInTerminal(command, cfg.Settings.WorkingDir)
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
