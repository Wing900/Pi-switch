// Package updater 提供基于 GitHub Releases 的版本检测与 Windows 自更新。
package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"piswitch/internal/config"
)

const (
	repoOwner     = "Wing900"
	repoName      = "Pi-switch"
	checkInterval = 7 * 24 * time.Hour
	httpTimeout   = 20 * time.Second
)

// CheckResult 描述一次版本检测的结果，直接映射到前端 JSON。
type CheckResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseURL     string `json:"releaseUrl"`
	AssetURL       string `json:"assetUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestAPIURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
}

// NeedsCheck 判断距上次检查是否已满 7 天，用于启动时的静默检测节流。
func NeedsCheck(settings config.AppSettings) bool {
	if settings.LastUpdateCheckAtUnix == 0 {
		return true
	}
	return time.Since(time.Unix(settings.LastUpdateCheckAtUnix, 0)) >= checkInterval
}

// CheckLatest 拉取 GitHub 最新 release，比对版本并定位 Windows 更新资产。
func CheckLatest(currentVersion string) (CheckResult, error) {
	req, err := http.NewRequest(http.MethodGet, latestAPIURL(), nil)
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("User-Agent", "Pi-Switch-updater/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CheckResult{}, fmt.Errorf("GitHub API 返回状态码 %d", resp.StatusCode)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return CheckResult{}, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	result := CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   strings.TrimSpace(release.Body),
	}

	cmp, err := Compare(currentVersion, latest)
	if err != nil {
		return result, err
	}
	result.HasUpdate = cmp < 0

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, windowsAssetSuffix()) {
			result.AssetURL = asset.BrowserDownloadURL
			break
		}
	}
	if result.HasUpdate && result.AssetURL == "" {
		return result, errors.New("最新版本缺少 Windows 更新包")
	}
	return result, nil
}

// Compare 比较 v1/v2（形如 0.0.0.11，可带 v 前缀），返回 1/0/-1。
func Compare(v1, v2 string) (int, error) {
	a := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	b := strings.Split(strings.TrimPrefix(v2, "v"), ".")
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		var x, y int
		var err error
		if i < len(a) {
			x, err = strconv.Atoi(a[i])
			if err != nil {
				return 0, fmt.Errorf("版本号格式错误: %s", a[i])
			}
		}
		if i < len(b) {
			y, err = strconv.Atoi(b[i])
			if err != nil {
				return 0, fmt.Errorf("版本号格式错误: %s", b[i])
			}
		}
		if x > y {
			return 1, nil
		}
		if x < y {
			return -1, nil
		}
	}
	return 0, nil
}

func windowsAssetSuffix() string {
	return "-windows-amd64.exe"
}

// Download 将 url 指向的资产下载到 dest 文件。
func Download(url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Pi-Switch-updater")
	// 大文件下载，不设总超时，仅依赖底层连接超时。
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// Install 将已下载的新 exe 通过延迟替换应用到当前可执行文件。
// 运行中的 exe 被系统锁定无法直接覆盖，故写入 update.bat 延迟执行：等待主进程
// 退出后用 move 替换再重启。调用方应在 Install 成功后退出主进程。
func Install(downloadedExe string) error {
	if runtime.GOOS != "windows" {
		return errors.New("当前平台不支持自动更新")
	}
	current, err := os.Executable()
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp("", "piswitch-update-")
	if err != nil {
		return err
	}
	script := filepath.Join(tmpDir, "update.bat")
	lines := []string{
		"@echo off",
		"setlocal",
		"timeout /t 3 /nobreak >nul",
		"move /y \"" + downloadedExe + "\" \"" + current + "\" >nul",
		"if errorlevel 1 goto :fail",
		"start \"\" \"" + current + "\"",
		"goto :end",
		":fail",
		"echo Pi Switch 更新替换失败，请手动安装新版本。",
		":end",
		"del \"%~f0\"",
	}
	if err := os.WriteFile(script, []byte(strings.Join(lines, "\r\n")+"\r\n"), 0o755); err != nil {
		return err
	}
	// 用 start 分离启动，主进程退出后 bat 仍能继续执行。
	cmd := exec.Command("cmd", "/c", "start", "", "/b", script)
	return cmd.Start()
}
