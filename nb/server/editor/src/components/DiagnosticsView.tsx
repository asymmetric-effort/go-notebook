// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { render } from '@asymmetric-effort/specifyjs/dom';

interface Diagnostic {
  file: string;
  line: number;
  col: number;
  message: string;
  severity: string;
}

interface DiagnosticsResponse {
  diagnostics: Diagnostic[];
}

let diagnostics: Diagnostic[] = [];
let diagLoading = true;

function rerenderDiagnostics(): void {
  const container = document.getElementById('diagnostics-content');
  if (container) {
    render(createElement(DiagnosticsContent, null), container);
  }
}

function fetchDiagnostics(): void {
  diagLoading = true;
  fetch('/api/lsp/diagnostics')
    .then((r) => r.json())
    .then((data: DiagnosticsResponse) => {
      diagnostics = data.diagnostics || [];
      diagLoading = false;
      rerenderDiagnostics();
    })
    .catch(() => {
      diagLoading = false;
      rerenderDiagnostics();
    });
}

function severityClass(severity: string): string {
  switch (severity) {
    case 'error': return 'diagnostic diagnostic-error';
    case 'warning': return 'diagnostic diagnostic-warning';
    case 'info': return 'diagnostic diagnostic-info';
    default: return 'diagnostic';
  }
}

function DiagnosticsContent(): ReturnType<typeof createElement> {
  if (diagLoading) {
    return createElement('div', null, 'Loading diagnostics...');
  }

  if (diagnostics.length === 0) {
    return createElement('p', { className: 'status ok' }, 'No diagnostics. Notebook is clean.');
  }

  return createElement(
    'div',
    { className: 'diagnostics-panel' },
    createElement('h3', null, `${diagnostics.length} diagnostic${diagnostics.length !== 1 ? 's' : ''}`),
    ...diagnostics.map((d) =>
      createElement(
        'div',
        { className: severityClass(d.severity) },
        d.line > 0
          ? `Line ${d.line}: ${d.message}`
          : d.message,
      ),
    ),
  );
}

export function DiagnosticsView(): ReturnType<typeof createElement> {
  setTimeout(fetchDiagnostics, 0);

  return createElement(
    'div',
    null,
    createElement('h2', null, 'Diagnostics'),
    createElement('button', { className: 'diagnostics-refresh-btn', onClick: fetchDiagnostics }, 'Refresh'),
    createElement('div', { id: 'diagnostics-content' }, 'Loading...'),
  );
}
