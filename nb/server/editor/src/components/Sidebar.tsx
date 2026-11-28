// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { createElement } from '@asymmetric-effort/specifyjs';
import { getActiveView, setActiveView } from '../state.js';
import type { ViewName } from '../state.js';

const navItems: Array<{ key: ViewName; label: string }> = [
  { key: 'notebook', label: 'Notebook' },
  { key: 'secrets', label: 'Secrets' },
  { key: 'diagnostics', label: 'Diagnostics' },
];

function handleNavClick(key: ViewName): void {
  setActiveView(key);
  // Close mobile sidebar.
  const sidebar = document.getElementById('sidebar');
  if (sidebar) sidebar.classList.remove('open');
}

function toggleMobileSidebar(): void {
  const sidebar = document.getElementById('sidebar');
  if (sidebar) sidebar.classList.toggle('open');
}

export function Sidebar(): ReturnType<typeof createElement> {
  const active = getActiveView();

  return createElement(
    'div',
    null,
    createElement(
      'button',
      {
        className: 'mobile-menu-btn',
        'aria-label': 'Toggle menu',
        onClick: toggleMobileSidebar,
      },
      '\u2630',
    ),
    createElement(
      'nav',
      { className: 'editor-sidebar', id: 'sidebar' },
      createElement('div', { className: 'sidebar-header' }, 'Navigation'),
      ...navItems.map((item) =>
        createElement(
          'button',
          {
            className: `sidebar-item${active === item.key ? ' active' : ''}`,
            onClick: () => handleNavClick(item.key),
          },
          item.label,
        ),
      ),
      createElement('div', { className: 'sidebar-spacer' }),
    ),
  );
}
