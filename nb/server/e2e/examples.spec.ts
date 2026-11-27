import { test, expect } from '@playwright/test';
import { readFileSync, existsSync } from 'fs';
import { resolve } from 'path';

const ROOT = resolve(__dirname, '..', '..', '..');

test('examples/hello.md exists in repo root', () => {
  const path = resolve(ROOT, 'examples', 'hello.md');
  expect(existsSync(path)).toBe(true);
});

test('examples/hello.md contains go:notebook fences', () => {
  const path = resolve(ROOT, 'examples', 'hello.md');
  const content = readFileSync(path, 'utf-8');
  expect(content).toContain('go:defs');
  expect(content).toContain('go:notebook');
});

test('site/public/examples/hello.md matches repo root copy', () => {
  const repoPath = resolve(ROOT, 'examples', 'hello.md');
  const sitePath = resolve(ROOT, 'site', 'public', 'examples', 'hello.md');
  expect(existsSync(sitePath)).toBe(true);
  const repoContent = readFileSync(repoPath, 'utf-8');
  const siteContent = readFileSync(sitePath, 'utf-8');
  expect(siteContent).toBe(repoContent);
});
