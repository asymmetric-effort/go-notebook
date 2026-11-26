import { defineConfig } from 'vite';
import { readFileSync, existsSync } from 'fs';
import { resolve } from 'path';
import { specifyJsSeoPlugin, specifyJsNoscriptPlugin } from '@asymmetric-effort/specifyjs/build';

// Load generated content for noscript fallback (built by prebuild script)
const contentPath = resolve(__dirname, 'src/generated/content.json');
let noscriptSections: Array<{ id: string; title: string; html: string }> = [];

if (existsSync(contentPath)) {
  try {
    const parsed = JSON.parse(readFileSync(contentPath, 'utf-8')) as Record<string, { title: string; html: string }>;
    for (const [id, entry] of Object.entries(parsed)) {
      noscriptSections.push({ id, title: entry.title, html: entry.html });
    }
  } catch {
    // Fall through to fallback
  }
}

if (noscriptSections.length === 0) {
  noscriptSections = [
    { id: 'home', title: 'Home', html: '<h1>go-notebook</h1><p>A Jupyter-like notebook built in Go, with Go as the cell language. Notebooks are CommonMark Markdown files that compile into standalone report binaries.</p>' },
  ];
}

const allRoutes = [
  '/',
  '#/getting-started',
  '#/user-guide',
  '#/charts',
  '#/encryption',
  '#/cli',
  '#/editor',
  '#/report-binary',
];

export default defineConfig({
  base: '/',
  plugins: [
    specifyJsSeoPlugin({
      siteUrl: 'https://go-notebooks.asymmetric-effort.com',
      title: 'go-notebook',
      description:
        'A Jupyter-like notebook built in Go, with Go as the cell language. ' +
        'Notebooks are CommonMark Markdown files that compile into standalone report binaries ' +
        'rendering Markdown, HTML, or PDF.',
      routes: allRoutes,
      npmPackage: undefined,
      author: 'Asymmetric Effort, LLC',
      license: 'MIT',
      repository: 'https://github.com/asymmetric-effort/go-notebook',
      robotsRules: ['User-agent: *', 'Allow: /'],
      jsonLd: {
        '@context': 'https://schema.org',
        '@type': 'SoftwareSourceCode',
        name: 'go-notebook',
        description:
          'A Jupyter-like notebook built in Go. Notebooks are CommonMark Markdown files ' +
          'with Go code cells that compile into standalone report binaries with PDF, HTML, and Markdown output. ' +
          'Features post-quantum encryption (ML-DSA-65, HPKE MLKEM768X25519), chart rendering, and an LSP-powered editor.',
        runtimePlatform: 'Go',
        applicationCategory: 'DeveloperApplication',
        codeRepository: 'https://github.com/asymmetric-effort/go-notebook',
        programmingLanguage: 'Go',
        license: 'https://opensource.org/licenses/MIT',
        author: {
          '@type': 'Organization',
          name: 'Asymmetric Effort, LLC',
        },
      },
    }),
    specifyJsNoscriptPlugin({
      title: 'go-notebook',
      sections: noscriptSections,
    }),
  ],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
});
