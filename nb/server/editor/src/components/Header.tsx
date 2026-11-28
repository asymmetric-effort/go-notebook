// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';

/** Cached version string, populated on first mount. */
let cachedVersion = '';

function fetchVersion(): void {
  fetch('/api/status')
    .then((r) => r.json())
    .then((data: { version?: string }) => {
      if (data.version) {
        cachedVersion = data.version;
        const el = document.getElementById('editor-version');
        if (el) el.textContent = `(${cachedVersion})`;
      }
    })
    .catch(() => {
      // Silently ignore fetch errors.
    });
}

function toggleDarkMode(): void {
  const html = document.documentElement;
  const isDark = html.getAttribute('data-theme') === 'dark';
  if (isDark) {
    html.removeAttribute('data-theme');
    localStorage.setItem('editor-theme', 'light');
  } else {
    html.setAttribute('data-theme', 'dark');
    localStorage.setItem('editor-theme', 'dark');
  }
  // Update toggle button text.
  const btn = document.getElementById('theme-toggle-header');
  if (btn) btn.textContent = isDark ? '\u263E' : '\u2600';
}

export function Header(): ReturnType<typeof createElement> {
  // Trigger version fetch on mount.
  setTimeout(fetchVersion, 0);

  const isDark = document.documentElement.getAttribute('data-theme') === 'dark';

  return createElement(
    'header',
    { className: 'editor-header' },
    createElement(
      'div',
      { className: 'editor-header-left' },
      createElement('img', {
        src: '/static/logo.png',
        alt: 'go-notebook',
        width: '28',
        height: '28',
        className: 'editor-logo',
      }),
      createElement(
        'h1',
        { className: 'editor-title' },
        'go-notebook editor ',
        createElement('span', { id: 'editor-version', className: 'editor-version' },
          cachedVersion ? `(${cachedVersion})` : ''),
      ),
    ),
    createElement(
      'button',
      {
        id: 'theme-toggle-header',
        className: 'theme-toggle',
        'aria-label': 'Toggle dark mode',
        onClick: toggleDarkMode,
      },
      isDark ? '\u2600' : '\u263E',
    ),
  );
}
