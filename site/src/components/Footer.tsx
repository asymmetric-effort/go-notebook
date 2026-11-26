import { createElement } from '@asymmetric-effort/specifyjs';
import { Footer as SpecFooter } from '@asymmetric-effort/specifyjs/components';

export function Footer(): ReturnType<typeof createElement> {
  return SpecFooter({
    left: createElement('span', null, 'go-notebook'),
    center: createElement(
      'span',
      null,
      'MIT License \u00A9 2026 Asymmetric Effort, LLC',
    ),
    right: createElement(
      'span',
      null,
      createElement(
        'a',
        {
          href: 'https://github.com/asymmetric-effort/go-notebook',
          style: { color: 'var(--text-dimmed)', textDecoration: 'none' },
          target: '_blank',
          rel: 'noopener noreferrer',
        },
        'GitHub',
      ),
      ' \u00B7 ',
      createElement(
        'a',
        {
          href: 'https://github.com/asymmetric-effort/go-notebook/blob/main/SECURITY.md',
          style: { color: 'var(--text-dimmed)', textDecoration: 'none' },
          target: '_blank',
          rel: 'noopener noreferrer',
        },
        'Security',
      ),
      ' \u00B7 ',
      createElement(
        'a',
        {
          href: 'https://github.com/asymmetric-effort/go-notebook/blob/main/LICENSE',
          style: { color: 'var(--text-dimmed)', textDecoration: 'none' },
          target: '_blank',
          rel: 'noopener noreferrer',
        },
        'License',
      ),
    ),
    background: 'var(--bg-primary)',
    color: 'var(--text-dimmed)',
    borderTop: '1px solid var(--border)',
    fontSize: '13px',
    padding: '1rem 20px',
  });
}
