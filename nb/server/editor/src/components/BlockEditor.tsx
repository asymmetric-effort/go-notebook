// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';

interface ReadFile {
  name: string;
  mode: 'embed' | 'runtime';
}

interface BlockInfo {
  kind: string;
  line: number;
  end_line: number;
  body: string;
  info: string;
  name: string;
  exported: string;
}

interface RunResponse {
  stdout: string;
  stderr: string;
  exit_code: number;
  error?: string;
  errors?: Array<{ message: string; kind: string; line?: number }>;
  displays?: Array<{ mime: string; data: string }>;
}

interface BlockEditorProps {
  block: BlockInfo;
  onRefresh: () => void;
}

/** Parse info string to extract cell attributes. */
function parseInfo(info: string): {
  name: string; exported: string; reads: ReadFile[];
  writes: string[]; secrets: string[]; data: string; timeout: string;
} {
  const result = { name: '', exported: '', reads: [] as ReadFile[], writes: [] as string[], secrets: [] as string[], data: '', timeout: '' };
  if (!info) return result;

  const nameMatch = info.match(/name=(\S+)/);
  if (nameMatch) result.name = nameMatch[1];

  const exportedMatch = info.match(/exported\(([^)]*)\)/);
  if (exportedMatch) result.exported = exportedMatch[1];

  const readsMatches = info.matchAll(/reads\(([^)]*)\)/g);
  for (const m of readsMatches) {
    const parts = m[1].split(',').map((s) => s.trim()).filter(Boolean);
    for (const p of parts) {
      const [name, mode] = p.split(':');
      result.reads.push({ name: name.trim(), mode: (mode?.trim() === 'runtime' ? 'runtime' : 'embed') });
    }
  }

  const writesMatch = info.match(/writes\(([^)]*)\)/);
  if (writesMatch) result.writes = writesMatch[1].split(',').map((s) => s.trim()).filter(Boolean);

  const secretsMatch = info.match(/secrets\(([^)]*)\)/);
  if (secretsMatch) result.secrets = secretsMatch[1].split(',').map((s) => s.trim()).filter(Boolean);

  const dataMatch = info.match(/data=(\S+)/);
  if (dataMatch) result.data = dataMatch[1];

  const timeoutMatch = info.match(/timeout=(\S+)/);
  if (timeoutMatch) result.timeout = timeoutMatch[1];

  return result;
}

function toggleCollapse(e: Event): void {
  const btn = e.currentTarget as HTMLButtonElement;
  const body = btn.closest('.block')?.querySelector('.block-body');
  if (!body) return;
  const isCollapsed = body.classList.toggle('collapsed');
  btn.textContent = isCollapsed ? '\u25B6' : '\u25BC';
}

function handleSaveCell(block: BlockInfo, onRefresh: () => void): void {
  const prefix = `block-${block.line}`;
  const bodyEl = document.getElementById(`${prefix}-body`) as HTMLTextAreaElement | null;
  const nameEl = document.getElementById(`${prefix}-name`) as HTMLInputElement | null;
  const exportedEl = document.getElementById(`${prefix}-exported`) as HTMLInputElement | null;
  const readsEl = document.getElementById(`${prefix}-reads`) as HTMLInputElement | null;
  const writesEl = document.getElementById(`${prefix}-writes`) as HTMLInputElement | null;
  const secretsEl = document.getElementById(`${prefix}-secrets`) as HTMLInputElement | null;
  const dataEl = document.getElementById(`${prefix}-data`) as HTMLSelectElement | null;
  const timeoutEl = document.getElementById(`${prefix}-timeout`) as HTMLInputElement | null;

  const reads: ReadFile[] = [];
  if (readsEl?.value) {
    for (const part of readsEl.value.split(',')) {
      const trimmed = part.trim();
      if (!trimmed) continue;
      const [name, mode] = trimmed.split(':');
      reads.push({ name: name.trim(), mode: mode?.trim() === 'runtime' ? 'runtime' : 'embed' });
    }
  }

  const payload = {
    line: block.line,
    body: bodyEl?.value ?? block.body,
    name: nameEl?.value ?? block.name,
    exported: exportedEl?.value ?? '',
    reads,
    writes: writesEl?.value ? writesEl.value.split(',').map((s) => s.trim()).filter(Boolean) : [],
    secrets: secretsEl?.value ? secretsEl.value.split(',').map((s) => s.trim()).filter(Boolean) : [],
    data: dataEl?.value ?? '',
    timeout: timeoutEl?.value ?? '',
  };

  fetch('/api/cells/update', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Save failed');
      onRefresh();
    })
    .catch((err: Error) => {
      alert(`Save error: ${err.message}`);
    });
}

function handleSaveProse(block: BlockInfo, onRefresh: () => void): void {
  const bodyEl = document.getElementById(`block-${block.line}-body`) as HTMLTextAreaElement | null;
  if (!bodyEl) return;

  fetch('/api/prose/update', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line: block.line, body: bodyEl.value }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Save failed');
      onRefresh();
    })
    .catch((err: Error) => {
      alert(`Save error: ${err.message}`);
    });
}

function handleDelete(block: BlockInfo, onRefresh: () => void): void {
  if (!confirm(`Delete this ${block.kind} block?`)) return;

  fetch('/api/blocks/delete', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line: block.line }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Delete failed');
      onRefresh();
    })
    .catch((err: Error) => {
      alert(`Delete error: ${err.message}`);
    });
}

function handleReorder(block: BlockInfo, direction: string, onRefresh: () => void): void {
  fetch('/api/blocks/reorder', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line: block.line, direction }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Reorder failed');
      onRefresh();
    })
    .catch((err: Error) => {
      alert(`Reorder error: ${err.message}`);
    });
}

function handleRunCell(cellName: string): void {
  const outputEl = document.getElementById(`output-${cellName}`);
  if (outputEl) {
    outputEl.textContent = 'Running...';
    outputEl.className = 'cell-output';
  }

  fetch('/api/run', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ cell: cellName }),
  })
    .then((r) => r.json())
    .then((data: RunResponse) => {
      if (!outputEl) return;
      if (data.error) {
        outputEl.textContent = data.error;
        outputEl.className = 'cell-output cell-errors';
      } else if (data.stderr) {
        outputEl.textContent = data.stderr;
        outputEl.className = 'cell-output cell-errors';
      } else {
        outputEl.textContent = data.stdout || '(no output)';
        outputEl.className = 'cell-output';
      }
    })
    .catch((err: Error) => {
      if (outputEl) {
        outputEl.textContent = `Error: ${err.message}`;
        outputEl.className = 'cell-output cell-errors';
      }
    });
}

function handleKeyDown(e: KeyboardEvent, block: BlockInfo, onRefresh: () => void): void {
  // Shift+Enter: run cell.
  if (e.shiftKey && e.key === 'Enter' && block.kind === 'cell' && block.name) {
    e.preventDefault();
    handleRunCell(block.name);
  }
  // Ctrl+S: save.
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault();
    if (block.kind === 'cell') {
      handleSaveCell(block, onRefresh);
    } else if (block.kind === 'prose') {
      handleSaveProse(block, onRefresh);
    }
  }
}

function CellBlockEditor(props: BlockEditorProps): ReturnType<typeof createElement> {
  const { block, onRefresh } = props;
  const attrs = parseInfo(block.info);
  const prefix = `block-${block.line}`;

  return createElement(
    'div',
    null,
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-name` }, 'Name'),
      createElement('input', {
        type: 'text', id: `${prefix}-name`,
        value: attrs.name, placeholder: 'cell_name',
      }),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-exported` }, 'Exported'),
      createElement('input', {
        type: 'text', id: `${prefix}-exported`,
        value: attrs.exported, placeholder: 'field1 int, field2 string',
      }),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-body` }, 'Body'),
      createElement('textarea', {
        id: `${prefix}-body`, className: 'mono',
        rows: Math.max(4, block.body.split('\n').length + 1),
        onKeyDown: (e: KeyboardEvent) => handleKeyDown(e, block, onRefresh),
      }, block.body),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-reads` }, 'Reads'),
      createElement('input', {
        type: 'text', id: `${prefix}-reads`,
        value: attrs.reads.map((r) => `${r.name}:${r.mode}`).join(', '),
        placeholder: 'file.txt:embed, data.csv:runtime',
      }),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-writes` }, 'Writes'),
      createElement('input', {
        type: 'text', id: `${prefix}-writes`,
        value: attrs.writes.join(', '), placeholder: 'output.txt',
      }),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-secrets` }, 'Secrets'),
      createElement('input', {
        type: 'text', id: `${prefix}-secrets`,
        value: attrs.secrets.join(', '), placeholder: 'API_KEY',
      }),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-data` }, 'Data'),
      createElement(
        'select',
        { id: `${prefix}-data` },
        createElement('option', { value: '', selected: !attrs.data }, 'none'),
        createElement('option', { value: 'embedded', selected: attrs.data === 'embedded' }, 'embedded'),
        createElement('option', { value: 'live', selected: attrs.data === 'live' }, 'live'),
      ),
    ),
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-timeout` }, 'Timeout'),
      createElement('input', {
        type: 'text', id: `${prefix}-timeout`,
        value: attrs.timeout, placeholder: '30s',
      }),
    ),
    createElement(
      'div',
      { className: 'btn-row' },
      createElement('button', {
        className: 'run-btn',
        onClick: () => { if (attrs.name) handleRunCell(attrs.name); },
      }, 'Run'),
      createElement('button', { onClick: () => handleSaveCell(block, onRefresh) }, 'Save'),
      createElement('button', {
        className: 'btn-danger',
        onClick: () => handleDelete(block, onRefresh),
      }, 'Delete'),
    ),
    createElement('pre', { id: `output-${attrs.name}`, className: 'cell-output' }),
  );
}

function ProseBlockEditor(props: BlockEditorProps): ReturnType<typeof createElement> {
  const { block, onRefresh } = props;
  const prefix = `block-${block.line}`;

  return createElement(
    'div',
    null,
    createElement(
      'div',
      { className: 'field' },
      createElement('label', { htmlFor: `${prefix}-body` }, 'Content'),
      createElement('textarea', {
        id: `${prefix}-body`, className: 'prose-editor',
        rows: Math.max(3, block.body.split('\n').length + 1),
        onKeyDown: (e: KeyboardEvent) => handleKeyDown(e, block, onRefresh),
      }, block.body),
    ),
    createElement(
      'div',
      { className: 'btn-row' },
      createElement('button', { onClick: () => handleSaveProse(block, onRefresh) }, 'Save'),
      createElement('button', {
        className: 'btn-danger',
        onClick: () => handleDelete(block, onRefresh),
      }, 'Delete'),
    ),
  );
}

function DefsBlockEditor(props: BlockEditorProps): ReturnType<typeof createElement> {
  const { block, onRefresh } = props;

  return createElement(
    'div',
    null,
    createElement('pre', null, block.body),
    createElement(
      'div',
      { className: 'btn-row' },
      createElement('button', {
        className: 'btn-danger',
        onClick: () => handleDelete(block, onRefresh),
      }, 'Delete'),
    ),
  );
}

function OutputBlockEditor(props: BlockEditorProps): ReturnType<typeof createElement> {
  return createElement(
    'div',
    null,
    createElement('pre', null, props.block.body),
  );
}

function blockClass(kind: string): string {
  switch (kind) {
    case 'cell': return 'block cell-block';
    case 'prose': return 'block prose-block';
    case 'defs': return 'block defs-block';
    case 'output': return 'block output-block';
    default: return 'block';
  }
}

function blockTitle(block: BlockInfo): string {
  switch (block.kind) {
    case 'cell': return `Cell: ${block.name || '(unnamed)'}`;
    case 'prose': return 'Prose';
    case 'defs': return 'Defs';
    case 'output': return 'Output';
    default: return block.kind;
  }
}

export function BlockEditor(props: BlockEditorProps): ReturnType<typeof createElement> {
  const { block, onRefresh } = props;

  let editor: ReturnType<typeof createElement>;
  switch (block.kind) {
    case 'cell':
      editor = createElement(CellBlockEditor, { block, onRefresh });
      break;
    case 'prose':
      editor = createElement(ProseBlockEditor, { block, onRefresh });
      break;
    case 'defs':
      editor = createElement(DefsBlockEditor, { block, onRefresh });
      break;
    default:
      editor = createElement(OutputBlockEditor, { block, onRefresh });
  }

  return createElement(
    'div',
    { className: blockClass(block.kind) },
    createElement(
      'div',
      { className: 'block-header' },
      createElement('h3', null, blockTitle(block)),
      createElement(
        'div',
        { className: 'block-actions' },
        createElement('button', { onClick: () => handleReorder(block, 'up', onRefresh), 'aria-label': 'Move up' }, '\u2191'),
        createElement('button', { onClick: () => handleReorder(block, 'down', onRefresh), 'aria-label': 'Move down' }, '\u2193'),
        createElement('button', { className: 'collapse-toggle', onClick: toggleCollapse, 'aria-label': 'Toggle collapse' }, '\u25BC'),
      ),
    ),
    createElement(
      'div',
      { className: 'block-body' },
      editor,
    ),
  );
}
