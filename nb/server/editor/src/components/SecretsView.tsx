// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { render } from '@asymmetric-effort/specifyjs/dom';

interface SecretsResponse {
  secrets: string[];
}

let secretNames: string[] = [];
let secretsLoading = true;

function rerenderSecrets(): void {
  const container = document.getElementById('secrets-content');
  if (container) {
    render(createElement(SecretsContent, null), container);
  }
}

function fetchSecrets(): void {
  secretsLoading = true;
  fetch('/api/secrets')
    .then((r) => r.json())
    .then((data: SecretsResponse) => {
      secretNames = data.secrets || [];
      secretsLoading = false;
      rerenderSecrets();
    })
    .catch(() => {
      secretsLoading = false;
      rerenderSecrets();
    });
}

function handleAddSecret(): void {
  const nameEl = document.getElementById('secret-name-input') as HTMLInputElement | null;
  const valueEl = document.getElementById('secret-value-input') as HTMLInputElement | null;
  if (!nameEl || !valueEl || !nameEl.value) return;

  fetch('/api/secrets', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: nameEl.value, value: valueEl.value }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Failed to save secret');
      nameEl.value = '';
      valueEl.value = '';
      fetchSecrets();
    })
    .catch((err: Error) => {
      alert(`Error: ${err.message}`);
    });
}

function handleDeleteSecret(name: string): void {
  fetch('/api/secrets', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name }),
  })
    .then((r) => {
      if (!r.ok) throw new Error('Failed to delete secret');
      fetchSecrets();
    })
    .catch((err: Error) => {
      alert(`Error: ${err.message}`);
    });
}

function handleClearAll(): void {
  if (!confirm('Clear all secrets?')) return;

  fetch('/api/secrets', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: '{}',
  })
    .then((r) => {
      if (!r.ok) throw new Error('Failed to clear secrets');
      fetchSecrets();
    })
    .catch((err: Error) => {
      alert(`Error: ${err.message}`);
    });
}

function SecretsContent(): ReturnType<typeof createElement> {
  if (secretsLoading) {
    return createElement('div', null, 'Loading secrets...');
  }

  const listItems = secretNames.map((name) =>
    createElement(
      'li',
      null,
      createElement('span', { className: 'secret-name' }, name),
      createElement('button', {
        className: 'delete-secret-btn',
        onClick: () => handleDeleteSecret(name),
      }, 'Delete'),
    ),
  );

  return createElement(
    'div',
    null,
    secretNames.length > 0
      ? createElement('ul', { className: 'secret-list' }, ...listItems)
      : createElement('p', { className: 'status' }, 'No secrets stored.'),
    createElement(
      'div',
      { className: 'block secret-block' },
      createElement('h3', null, 'Add / Update Secret'),
      createElement(
        'div',
        { className: 'field' },
        createElement('label', { htmlFor: 'secret-name-input' }, 'Name'),
        createElement('input', {
          type: 'text', id: 'secret-name-input',
          placeholder: 'SECRET_NAME',
        }),
      ),
      createElement(
        'div',
        { className: 'field' },
        createElement('label', { htmlFor: 'secret-value-input' }, 'Value'),
        createElement('input', {
          type: 'password', id: 'secret-value-input',
          placeholder: 'secret value',
        }),
      ),
      createElement(
        'div',
        { className: 'btn-row' },
        createElement('button', { onClick: handleAddSecret }, 'Save Secret'),
      ),
    ),
    secretNames.length > 0
      ? createElement('button', { className: 'clear-all-btn', onClick: handleClearAll }, 'Clear All')
      : null,
  );
}

export function SecretsView(): ReturnType<typeof createElement> {
  setTimeout(fetchSecrets, 0);

  return createElement(
    'div',
    null,
    createElement('h2', null, 'Secrets'),
    createElement('div', { id: 'secrets-content' }, 'Loading...'),
  );
}
