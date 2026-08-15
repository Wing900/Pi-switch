export const THINKING_LEVELS = [
  ["off", "关闭"],
  ["minimal", "Minimal"],
  ["low", "Low"],
  ["medium", "Medium"],
  ["high", "High"],
  ["xhigh", "XHigh"],
  ["max", "Max"]
];

export const COMPAT_BOOLEAN_FIELDS = [
  ["supportsStore", "支持 store", "是否接受请求中的 store 字段。"],
  ["supportsDeveloperRole", "支持 developer 角色", "不支持时，Pi 会将系统提示改用 system 角色。"],
  ["supportsReasoningEffort", "支持 reasoning_effort", "服务端是否接受 reasoning_effort 参数。"],
  ["supportsUsageInStreaming", "流式返回 usage", "是否支持 stream_options.include_usage。"],
  ["supportsFinishReason", "流式返回 finish_reason", "流式响应是否会返回 finish_reason；不返回时 Pi 可根据结束状态推断。"],
  ["requiresToolResultName", "工具结果需要 name", "工具结果消息是否必须带 name。"],
  ["requiresAssistantAfterToolResult", "工具结果后需要 assistant", "某些兼容接口要求工具结果后插入 assistant 消息。"],
  ["requiresThinkingAsText", "将思考转成文本", "服务端不支持原生 reasoning 时，将思考内容作为普通文本重放。"],
  ["requiresReasoningContentOnAssistantMessages", "补充 reasoning_content", "重放 assistant 消息时补充空的 reasoning_content。"],
  ["supportsOpenAIGrammarTools", "支持 OpenAI Grammar Tools", "是否支持 Lark/正则语法约束工具。"],
  ["supportsStrictMode", "支持严格工具模式", "是否接受严格 JSON Schema 工具定义。"],
  ["sendSessionAffinityHeaders", "发送会话亲和 Header", "将会话 ID 放入请求头，帮助代理复用缓存或路由。"],
  ["supportsLongCacheRetention", "支持长缓存保留", "是否接受较长的提示缓存保留参数。"],
  ["supportsEagerToolInputStreaming", "支持 eager 工具流", "Anthropic 兼容接口是否接受工具的 eager_input_streaming 字段。"],
  ["supportsCacheControlOnTools", "工具支持 cache_control", "Anthropic 兼容接口是否接受工具定义上的缓存标记。"],
  ["supportsTemperature", "支持 temperature", "Anthropic 接口是否接受 temperature 参数。"],
  ["forceAdaptiveThinking", "强制 adaptive thinking", "使用 Anthropic 的 adaptive thinking，而不是旧的 budget 模式。"],
  ["allowEmptySignature", "允许空 thinking signature", "仅适用于会返回空思考签名的 Anthropic 兼容代理。"],
  ["supportsStrictTools", "Anthropic 严格工具", "Anthropic 接口是否接受严格 JSON Schema 工具。"],
  ["supportsToolReferences", "支持工具引用", "Anthropic 兼容接口是否支持工具引用。"],
  ["supportsToolSearch", "支持工具搜索", "Responses API 是否支持工具搜索能力。"],
  ["zaiToolStream", "支持 Z.ai 工具流", "是否发送 tool_stream 以流式接收 Z.ai 工具调用。"],
  ["supportsThinkingTokenBudget", "支持思考 token 预算", "是否支持 vLLM 风格的 thinking_token_budget。"],
  ["supportsAdditionalTools", "支持 additional_tools", "Responses API 是否支持按消息挂载 additional_tools。"],
  ["supportsExplicitPromptCacheMode", "支持显式提示缓存", "是否支持 prompt_cache_options 的显式缓存模式。"]
];

export const COMPAT_ENUM_FIELDS = [
  ["maxTokensField", "最大输出字段", [["max_completion_tokens", "max_completion_tokens"], ["max_tokens", "max_tokens"]], "不同 OpenAI 兼容服务对最大输出参数的字段名不同。"],
  ["thinkingFormat", "思考参数格式", [["openai", "OpenAI"], ["openrouter", "OpenRouter"], ["together", "Together"], ["baseten", "Baseten"], ["deepseek", "DeepSeek"], ["zai", "Z.ai"], ["qwen", "Qwen"], ["chat-template", "Chat Template"], ["qwen-chat-template", "Qwen Chat Template"], ["string-thinking", "String Thinking"], ["ant-ling", "Ant Ling"]], "告诉 Pi 应该用哪种服务商专属格式发送思考开关。"],
  ["cacheControlFormat", "缓存标记格式", [["anthropic", "Anthropic cache_control"]], "只有支持 Anthropic 风格 cache_control 的 OpenAI 兼容代理才填写。"],
  ["deferredToolsMode", "延迟工具模式", [["kimi", "Kimi"]], "仅 Kimi 兼容接口使用；不要为了普通接口开启。"],
  ["sessionAffinityFormat", "会话 Header 格式", [["openai", "OpenAI"], ["openai-nosession", "OpenAI（不发送 session_id）"], ["openrouter", "OpenRouter"]], "决定会话 ID 应该使用哪种请求头格式。"]
];
