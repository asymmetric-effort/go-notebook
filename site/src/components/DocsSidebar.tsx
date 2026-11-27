import { createElement } from '@asymmetric-effort/specifyjs';

interface DocEntry {
  title: string;
  route: string;
}

interface DocSection {
  title: string;
  children: DocEntry[];
}

const docsTree: DocSection[] = [
  {
    title: 'Getting Started',
    children: [
      { title: 'Installation & First Notebook', route: '#/getting-started' },
    ],
  },
  {
    title: 'User Guide',
    children: [
      { title: 'Notebook Structure', route: '#/user-guide' },
      { title: 'Charts', route: '#/charts' },
      { title: 'Encryption', route: '#/encryption' },
    ],
  },
  {
    title: 'Reference',
    children: [
      { title: 'CLI Commands', route: '#/cli' },
      { title: 'Browser Editor', route: '#/editor' },
      { title: 'Report Binaries', route: '#/report-binary' },
    ],
  },
];

function SidebarSection(props: {
  section: DocSection;
  activeRoute: string;
}): ReturnType<typeof createElement> {
  const { section, activeRoute } = props;
  const hasActiveChild = section.children.some((c) => c.route === activeRoute);

  return createElement(
    'div',
    { style: { margin: '4px 0' } },
    createElement(
      'div',
      {
        style: {
          padding: '8px 12px',
          fontSize: '12px',
          fontWeight: '600',
          color: hasActiveChild ? 'var(--text-primary)' : 'var(--text-muted)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
        },
      },
      section.title,
    ),
    ...section.children.map((entry) => {
      const isActive = activeRoute === entry.route;
      return createElement(
        'a',
        {
          href: entry.route,
          style: {
            display: 'block',
            padding: '6px 12px 6px 24px',
            borderRadius: '4px',
            fontSize: '13px',
            fontWeight: isActive ? '600' : '400',
            color: isActive ? 'var(--accent)' : 'var(--text-muted)',
            backgroundColor: isActive ? 'var(--bg-tertiary)' : 'transparent',
            textDecoration: 'none',
            transition: 'all 0.15s',
          },
        },
        entry.title,
      );
    }),
  );
}

export function DocsSidebar(props: {
  activeRoute: string;
}): ReturnType<typeof createElement> {
  return createElement(
    'nav',
    {
      className: 'docs-sidebar',
      style: {
        width: '240px',
        minWidth: '240px',
        borderRight: '1px solid var(--border)',
        backgroundColor: 'var(--bg-secondary)',
        overflowY: 'auto',
        padding: '16px 8px',
      },
    },
    createElement(
      'div',
      {
        style: {
          padding: '4px 12px 16px',
          borderBottom: '1px solid var(--border)',
          marginBottom: '8px',
          fontSize: '15px',
          fontWeight: '700',
          color: 'var(--text-primary)',
        },
      },
      'Documentation',
    ),
    ...docsTree.map((section) =>
      createElement(SidebarSection, {
        section,
        activeRoute: props.activeRoute,
      }),
    ),
  );
}
