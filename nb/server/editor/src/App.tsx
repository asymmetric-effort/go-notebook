// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { Header } from './components/Header.js';
import { Sidebar } from './components/Sidebar.js';
import { NotebookView } from './components/NotebookView.js';
import { SecretsView } from './components/SecretsView.js';
import { DiagnosticsView } from './components/DiagnosticsView.js';
import { getActiveView } from './state.js';

function ActiveSection(): ReturnType<typeof createElement> {
  switch (getActiveView()) {
    case 'secrets':
      return createElement(SecretsView, null);
    case 'diagnostics':
      return createElement(DiagnosticsView, null);
    default:
      return createElement(NotebookView, null);
  }
}

export function App(): ReturnType<typeof createElement> {
  return createElement(
    'div',
    { className: 'editor-layout' },
    createElement(Sidebar, null),
    createElement(
      'div',
      { className: 'editor-main' },
      createElement(Header, null),
      createElement(ActiveSection, null),
    ),
  );
}
