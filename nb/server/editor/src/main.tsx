// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { render } from '@asymmetric-effort/specifyjs/dom';
import { App } from './App.js';
import { setRerenderFn } from './state.js';

function mount(): void {
  render(createElement(App, null), document.getElementById('root')!);
}

// Initialize theme from localStorage or system preference.
(function initTheme(): void {
  const saved = localStorage.getItem('editor-theme');
  if (saved === 'dark' || (!saved && matchMedia('(prefers-color-scheme:dark)').matches)) {
    document.documentElement.setAttribute('data-theme', 'dark');
  }
})();

// Register the re-render function so view changes trigger a full re-render.
setRerenderFn(mount);

mount();
