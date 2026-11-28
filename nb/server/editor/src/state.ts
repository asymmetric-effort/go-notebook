// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

/** Global state for the active editor view. */
export type ViewName = 'notebook' | 'secrets' | 'diagnostics';

let activeView: ViewName = 'notebook';
let rerenderFn: (() => void) | null = null;

export function getActiveView(): ViewName {
  return activeView;
}

export function setActiveView(view: ViewName): void {
  activeView = view;
  if (rerenderFn) rerenderFn();
}

export function setRerenderFn(fn: () => void): void {
  rerenderFn = fn;
}
