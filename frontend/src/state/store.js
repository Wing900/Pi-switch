export function createStore(initialState) {
  let state = initialState;
  const listeners = new Set();

  return {
    getState() {
      return state;
    },
    setState(updater, options = {}) {
      const nextState = typeof updater === "function" ? updater(state) : updater;
      state = nextState;
      if (options.notify !== false) {
        listeners.forEach((listener) => listener(state));
      }
    },
    subscribe(listener) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    }
  };
}
