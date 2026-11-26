import { createElement } from '@asymmetric-effort/specifyjs';
import { Hero } from './Hero.js';
import { Features } from './Features.js';
import { QuickStart } from './QuickStart.js';
import { MarkdownPage } from './MarkdownPage.js';
import { content } from '../generated/content.js';

function getHash(): string {
  return globalThis.location?.hash?.slice(1) || '/';
}

const contentRoutes: Record<string, string> = {
  '/getting-started': 'getting-started',
  '/user-guide': 'user-guide',
  '/charts': 'charts',
  '/encryption': 'encryption',
  '/cli': 'cli',
  '/editor': 'editor',
  '/report-binary': 'report-binary',
};

export function Router(): ReturnType<typeof createElement> {
  const route = getHash();

  if (route === '/') {
    return createElement(
      'main',
      null,
      createElement(Hero, null),
      createElement(Features, null),
      createElement(QuickStart, null),
    );
  }

  const contentKey = contentRoutes[route];
  if (contentKey) {
    const entry = content[contentKey];
    const title = entry?.title || contentKey;
    return createElement(MarkdownPage, { title, section: contentKey });
  }

  return createElement(
    'main',
    {
      style: {
        padding: '4rem 2rem',
        textAlign: 'center',
        color: 'var(--text-muted)',
        fontFamily: 'system-ui, sans-serif',
      },
    },
    createElement('h1', { style: { color: 'var(--text-primary)' } }, '404'),
    createElement('p', null, 'Page not found.'),
    createElement(
      'a',
      { href: '#/', style: { color: 'var(--accent)', textDecoration: 'none' } },
      'Go home',
    ),
  );
}
