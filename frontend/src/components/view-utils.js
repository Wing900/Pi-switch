export function escapeHtml(value = "") {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

export function extractHost(baseUrl) {
  if (!baseUrl) return "尚未配置端点";
  let trimmed = baseUrl.trim();
  trimmed = trimmed.replace(/^https?:\/\//, "");
  const idx = trimmed.indexOf("/");
  return idx >= 0 ? trimmed.slice(0, idx) : trimmed;
}
