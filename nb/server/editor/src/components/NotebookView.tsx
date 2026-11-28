// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { render } from '@asymmetric-effort/specifyjs/dom';
import { BlockEditor } from './BlockEditor.js';

interface BlockInfo {
  kind: string;
  line: number;
  end_line: number;
  body: string;
  info: string;
  name: string;
  exported: string;
}

/** Cached blocks from last fetch. */
let blocks: BlockInfo[] = [];
let loading = true;
let errorMsg = '';
let allCollapsed = false;

function rerenderNotebook(): void {
  const container = document.getElementById('notebook-view');
  if (!container) return;
  render(createElement(NotebookContent, null), container);
}

function fetchBlocks(): void {
  loading = true;
  fetch('/api/blocks')
    .then((r) => r.json())
    .then((data: BlockInfo[]) => {
      blocks = data;
      loading = false;
      errorMsg = '';
      rerenderNotebook();
    })
    .catch((err: Error) => {
      loading = false;
      errorMsg = err.message;
      rerenderNotebook();
    });
}

function handleRunAll(): void {
  const statusEl = document.getElementById('notebook-status');
  if (statusEl) {
    statusEl.textContent = 'Running all cells...';
    statusEl.className = 'status';
  }
  fetch('/api/run/all', { method: 'POST' })
    .then((r) => r.json())
    .then(() => {
      if (statusEl) {
        statusEl.textContent = 'All cells executed.';
        statusEl.className = 'status ok';
      }
      fetchBlocks();
    })
    .catch((err: Error) => {
      if (statusEl) {
        statusEl.textContent = `Error: ${err.message}`;
        statusEl.className = 'status error';
      }
    });
}

function toggleCollapseAll(): void {
  allCollapsed = !allCollapsed;
  const bodies = document.querySelectorAll('.block-body');
  bodies.forEach((el) => {
    if (allCollapsed) {
      el.classList.add('collapsed');
    } else {
      el.classList.remove('collapsed');
    }
  });
  // Update toggle buttons.
  const toggles = document.querySelectorAll('.collapse-toggle');
  toggles.forEach((el) => {
    el.textContent = allCollapsed ? '\u25B6' : '\u25BC';
  });
}

function handleInsertBlock(kind: string, afterLine: number): void {
  const info = kind === 'cell' ? 'go:notebook name=cell1' : kind === 'defs' ? 'go:defs' : '';
  const body = kind === 'prose' ? 'New paragraph.' : '';
  fetch('/api/blocks/add', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ kind, after_line: afterLine, body, info }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Failed to add block');
      fetchBlocks();
    })
    .catch((err: Error) => {
      const statusEl = document.getElementById('notebook-status');
      if (statusEl) {
        statusEl.textContent = `Error: ${err.message}`;
        statusEl.className = 'status error';
      }
    });
}

function handleInitNotebook(): void {
  fetch('/api/notebook/init', { method: 'POST' })
    .then((r) => {
      if (!r.ok) throw new Error('Failed to create notebook');
      fetchBlocks();
    })
    .catch((err: Error) => {
      const statusEl = document.getElementById('notebook-status');
      if (statusEl) {
        statusEl.textContent = `Error: ${err.message}`;
        statusEl.className = 'status error';
      }
    });
}

function InsertToolbar(props: { afterLine: number }): ReturnType<typeof createElement> {
  return createElement(
    'div',
    { className: 'insert-toolbar' },
    createElement('button', { onClick: () => handleInsertBlock('prose', props.afterLine) }, '+ Prose'),
    createElement('button', { onClick: () => handleInsertBlock('cell', props.afterLine) }, '+ Cell'),
    createElement('button', { onClick: () => handleInsertBlock('defs', props.afterLine) }, '+ Defs'),
  );
}

function GetStartedPanel(): ReturnType<typeof createElement> {
  return createElement(
    'div',
    { className: 'get-started' },
    createElement('h2', null, 'No blocks found'),
    createElement('p', null, 'This notebook is empty. Create a starter notebook to get going.'),
    createElement(
      'button',
      { className: 'starter-btn', onClick: handleInitNotebook },
      'Create Starter Notebook',
    ),
  );
}

function NotebookContent(): ReturnType<typeof createElement> {
  if (loading) {
    return createElement('div', null, 'Loading...');
  }
  if (errorMsg) {
    return createElement('div', { className: 'status error' }, `Error: ${errorMsg}`);
  }
  if (blocks.length === 0) {
    return createElement(GetStartedPanel, null);
  }

  const children: Array<ReturnType<typeof createElement>> = [];

  // Insert toolbar at top.
  children.push(createElement(InsertToolbar, { afterLine: 0 }));

  for (const block of blocks) {
    children.push(createElement(BlockEditor, { block, onRefresh: fetchBlocks }));
    children.push(createElement(InsertToolbar, { afterLine: block.end_line }));
  }

  return createElement('div', null, ...children);
}

export function NotebookView(): ReturnType<typeof createElement> {
  // Trigger fetch on mount.
  setTimeout(fetchBlocks, 0);

  return createElement(
    'div',
    null,
    createElement(
      'div',
      { className: 'toolbar' },
      createElement('button', { className: 'run-all-btn', onClick: handleRunAll }, 'Run All'),
      createElement('button', { className: 'collapse-toolbar-btn', onClick: toggleCollapseAll },
        allCollapsed ? 'Expand All' : 'Collapse All'),
    ),
    createElement('div', { id: 'notebook-status', className: 'status' }),
    createElement('div', { id: 'notebook-view' }, 'Loading...'),
  );
}
