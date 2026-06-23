export function errorMessage(error) {
  return error instanceof Error ? error.message : String(error ?? "未知错误");
}

export function createOperationFeedback(store) {
  function showResult({ status, title, message, details = [] }) {
    store.setState((state) => ({
      ...state,
      modal: {
        kind: "operation-result",
        payload: { status, title, message, details }
      }
    }));
  }

  function showError(title, error) {
    const message = errorMessage(error);
    showResult({
      status: "error",
      title,
      message,
      details: ["请检查配置后重试。"],
      log: `${title}：${message}`
    });
  }

  function showLoading(title, message) {
    store.setState((state) => ({
      ...state,
      modal: { kind: "operation-loading", payload: { title, message } }
    }));
  }

  return { showResult, showError, showLoading };
}
