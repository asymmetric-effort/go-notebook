import { createElement } from '@asymmetric-effort/specifyjs';

interface FeatureCardProps {
  title: string;
  description: string;
}

function FeatureCard(
  props: FeatureCardProps,
): ReturnType<typeof createElement> {
  return createElement(
    'div',
    {
      style: {
        flex: '1',
        minWidth: '250px',
        padding: '16px 20px',
        backgroundColor: 'var(--bg-secondary)',
        borderRadius: '6px',
        border: '1px solid var(--border)',
      },
    },
    createElement(
      'h3',
      {
        style: {
          color: 'var(--text-primary)',
          marginBottom: '0.5rem',
          fontSize: '1.1rem',
        },
      },
      props.title,
    ),
    createElement(
      'p',
      {
        style: {
          color: 'var(--text-muted)',
          lineHeight: '1.5',
          fontSize: '0.9rem',
          margin: '0',
        },
      },
      props.description,
    ),
  );
}

export function Features(): ReturnType<typeof createElement> {
  const features: FeatureCardProps[] = [
    {
      title: 'CommonMark Notebooks',
      description:
        'Notebooks are standard Markdown files with go:notebook code fences. ' +
        'Edit them in any text editor, track changes with git, and render them anywhere.',
    },
    {
      title: 'Standalone Report Binaries',
      description:
        'gobook build compiles a notebook into a self-contained executable. ' +
        'Distribute reports as single binaries that render PDF, HTML, or Markdown output on any platform.',
    },
    {
      title: 'PDF, HTML & Markdown Output',
      description:
        'Render notebooks in multiple formats. PDF via Typst (zero cgo, WASM-based), ' +
        'HTML with inline charts and images, or clean Markdown.',
    },
    {
      title: 'Post-Quantum Encryption',
      description:
        'Encrypt reports with ML-DSA-65 signatures and HPKE MLKEM768X25519 key encapsulation. ' +
        'Passphrase mode (Argon2id + AES-256-GCM) or public-key recipient mode.',
    },
    {
      title: 'Built-in Chart Rendering',
      description:
        'Seven chart types via ctx.Chart(): line, bar, scatter, pie, histogram, heatmap, and box plots. ' +
        'Charts render as SVG in HTML and inline in PDF.',
    },
    {
      title: 'LSP-Powered Editor',
      description:
        'gobook serve launches a browser-based notebook editor with Go autocompletion, ' +
        'diagnostics, and hover docs powered by gopls over a Unix socket kernel.',
    },
    {
      title: 'Cross-Platform Builds',
      description:
        'Build report binaries for any OS/architecture combination. ' +
        'The --os and --arch flags on gobook build enable cross-compilation.',
    },
    {
      title: 'Zero cgo Dependencies',
      description:
        'The entire toolchain is pure Go. Typst PDF rendering uses wazero (WebAssembly). ' +
        'No C compiler needed, no system library dependencies.',
    },
    {
      title: 'VSCode Extension',
      description:
        'Syntax highlighting and language support for go-notebook files in Visual Studio Code. ' +
        'Go code blocks get full language intelligence.',
    },
  ];

  return createElement(
    'section',
    {
      style: {
        padding: '3rem 20px',
        backgroundColor: 'var(--bg-primary)',
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
      'Features',
    ),
    createElement(
      'div',
      {
        style: {
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
          gap: '16px',
          maxWidth: '900px',
          margin: '0 auto',
        },
      },
      ...features.map((f) => createElement(FeatureCard, f)),
    ),
  );
}
