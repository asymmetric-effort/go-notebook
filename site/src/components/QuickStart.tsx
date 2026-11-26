import { createElement } from '@asymmetric-effort/specifyjs';

export function QuickStart(): ReturnType<typeof createElement> {
  const installExample = `# Install gobook
go install github.com/asymmetric-effort/go-notebook/cmd/gobook@latest

# Verify installation
gobook version`;

  const notebookExample = `# My First Report

Some introductory prose in CommonMark Markdown.

\`\`\`go:defs
// Shared definitions available to all cells.
import "fmt"
\`\`\`

\`\`\`go:notebook
// This cell runs and produces output.
fmt.Println("Hello from go-notebook!")
\`\`\``;

  const buildExample = `# Build a standalone report binary
gobook build report.md -o report

# Run it to generate output
./report --format html -o report.html
./report --format pdf -o report.pdf

# Or use gobook run for quick iteration
gobook run report.md -- --format html -o out.html`;

  return createElement(
    'section',
    {
      style: {
        padding: '3rem 20px',
        backgroundColor: 'var(--bg-secondary)',
      },
    },
    createElement(
      'h2',
      {
        style: {
          textAlign: 'center',
          color: 'var(--text-primary)',
          marginBottom: '2rem',
          fontSize: '1.5rem',
        },
      },
      'Quick Start',
    ),
    createElement(
      'div',
      {
        style: {
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
          gap: '20px',
          maxWidth: '900px',
          margin: '0 auto',
        },
      },
      codeBlock('Install', installExample),
      codeBlock('Write a Notebook', notebookExample),
      codeBlock('Build & Run', buildExample),
    ),
  );
}

function codeBlock(
  title: string,
  code: string,
): ReturnType<typeof createElement> {
  return createElement(
    'div',
    null,
    createElement(
      'h3',
      {
        style: {
          color: 'var(--text-primary)',
          marginBottom: '0.5rem',
          fontSize: '1.2rem',
        },
      },
      title,
    ),
    createElement(
      'pre',
      {
        style: {
          padding: '12px 16px',
          backgroundColor: 'var(--code-bg)',
          borderRadius: '6px',
          overflow: 'auto',
          fontSize: '13px',
          lineHeight: '1.5',
          color: 'var(--code-text)',
          fontFamily: 'monospace',
        },
      },
      createElement('code', null, code),
    ),
  );
}
