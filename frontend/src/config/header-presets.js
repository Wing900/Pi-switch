export const HEADER_MODES = [
  { id: "none", label: "不添加" },
  { id: "claude-code", label: "Claude Code" },
  { id: "codex", label: "Codex" },
  { id: "grok-build", label: "Grok Build" },
  { id: "custom", label: "自定义" }
];

const PRESETS = {
  "claude-code": {
    "user-agent": "claude-cli/2.1.219 (external, sdk-cli)",
    "x-stainless-arch": "x64",
    "x-stainless-lang": "js",
    "x-stainless-os": "Windows",
    "x-stainless-package-version": "0.94.0",
    "x-stainless-retry-count": "0",
    "x-stainless-runtime": "node",
    "x-stainless-runtime-version": "v26.3.0",
    "x-stainless-timeout": "300",
    "anthropic-version": "2023-06-01",
    "x-app": "cli"
  },
  codex: {
    "user-agent": "codex-tui/0.145.0 (Windows 10.0.26200; x86_64) WindowsTerminal (codex-tui; 0.145.0)",
    originator: "codex-tui",
    "x-codex-beta-features": "remote_compaction_v2"
  },
  "grok-build": {
    "user-agent": "grok-build"
  }
};

export function cloneHeaders(headers = {}) {
  return Object.fromEntries(
    Object.entries(headers)
      .filter(([name, value]) => name.trim() && typeof value === "string")
      .map(([name, value]) => [name.trim(), value])
  );
}

export function headersForMode(mode, api, customHeaders = {}) {
  if (mode === "none") return {};
  if (mode === "custom") return cloneHeaders(customHeaders);
  if (mode === "auto") {
    if (api === "anthropic-messages") return cloneHeaders(PRESETS["claude-code"]);
    if (api === "openai-completions" || api === "openai-responses") return cloneHeaders(PRESETS.codex);
    return {};
  }
  return cloneHeaders(PRESETS[mode]);
}

export function headerModeLabel(mode) {
  return HEADER_MODES.find((item) => item.id === mode)?.label ?? "不添加";
}

export function customHeadersForProvider(provider) {
  if (provider?.headerMode !== "custom") return {};
  return cloneHeaders(provider.customHeaders ?? provider.headers);
}
