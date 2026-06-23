export function transitionState(store, updater, transition = "default") {
  const commit = () => store.setState(updater);

  if (!document.startViewTransition) {
    commit();
    return;
  }

  document.documentElement.dataset.transition = transition;
  const viewTransition = document.startViewTransition(commit);
  viewTransition.finished.finally(() => {
    delete document.documentElement.dataset.transition;
  });
}
