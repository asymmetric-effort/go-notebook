import { createElement } from '@asymmetric-effort/specifyjs';
import { content } from '../generated/content.js';
import { DocsSidebar } from './DocsSidebar.js';

interface MarkdownPageProps {
  title: string;
  section: string;
}

export function MarkdownPage(
  props: MarkdownPageProps,
): ReturnType<typeof createElement> {
  const entry = content[props.section];
  const title = entry?.title || props.title;
  const html = entry?.html || '<p>Content not found.</p>';
  const activeRoute = `#/${props.section}`;

  if (typeof document !== 'undefined') {
    document.title = `${title} | go-notebook`;
  }

  return createElement(
    'div',
    {
      style: {
        display: 'flex',
        minHeight: 'calc(100vh - 120px)',
      },
    },
    createElement(DocsSidebar, { activeRoute }),
    createElement(
      'main',
      {
        className: 'ns-section-content',
        style: {
          flex: '1',
          maxWidth: '780px',
          padding: '2rem 40px',
          color: 'var(--text-secondary)',
          lineHeight: '1.6',
          overflowY: 'auto',
        },
      },
      createElement('div', { dangerouslySetInnerHTML: { __html: html } }),
      createElement(
        'div',
        {
          style: {
            marginTop: '2rem',
            paddingTop: '1rem',
            borderTop: '1px solid var(--border)',
            fontSize: '13px',
          },
        },
        createElement(
          'a',
          {
            href: '#/',
            style: {
              color: 'var(--accent)',
              textDecoration: 'none',
            },
          },
          '\u2190 Back to home',
        ),
      ),
    ),
  );
}
