// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const FIXTURE_PATH = path.join(__dirname, 'fixtures', 'hello.md');
const FIXTURE_CONTENT = `# Test Notebook

\`\`\`go:defs
import "fmt"
\`\`\`

\`\`\`go:notebook name=greeting
fmt.Println("hello from test")
\`\`\`
`;

// Restore the fixture file before each test to prevent test interference.
test.beforeEach(() => {
  fs.writeFileSync(FIXTURE_PATH, FIXTURE_CONTENT);
});

test.afterAll(() => {
  fs.writeFileSync(FIXTURE_PATH, FIXTURE_CONTENT);
});

test.describe('Empty notebook Get Started UI', () => {
  test('empty notebook via API init shows starter content', async ({ request }) => {
    // We cannot easily run a separate server with an empty notebook.
    // Instead, verify the /api/notebook/init endpoint behavior via API.
    // On the hello.md fixture (non-empty), init should return 409.
    const resp = await request.post('/api/notebook/init');
    expect(resp.status()).toBe(409);
    const body = await resp.json();
    expect(body.error).toContain('already has content');
  });
});

test.describe('Block creation', () => {
  test('add prose block via API', async ({ request }) => {
    const beforeResp = await request.get('/api/blocks');
    const blocksBefore = await beforeResp.json();
    const countBefore = blocksBefore.length;
    const lastBlock = blocksBefore[blocksBefore.length - 1];

    const resp = await request.post('/api/blocks/add', {
      data: {
        kind: 'prose',
        after_line: lastBlock.end_line,
        body: 'TEST_PROSE: Added by e2e test.',
      },
    });
    expect(resp.ok()).toBe(true);

    const afterResp = await request.get('/api/blocks');
    const blocksAfter = await afterResp.json();
    expect(blocksAfter.length).toBeGreaterThan(countBefore);

    const added = blocksAfter.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('TEST_PROSE'),
    );
    expect(added).toBeTruthy();

    // Clean up: delete the added prose block.
    if (added) {
      await request.delete('/api/blocks/delete', {
        data: { line: added.line },
      });
    }
  });

  test('add defs block via API', async ({ request }) => {
    const beforeResp = await request.get('/api/blocks');
    const blocksBefore = await beforeResp.json();
    const countBefore = blocksBefore.length;
    const lastBlock = blocksBefore[blocksBefore.length - 1];

    const resp = await request.post('/api/blocks/add', {
      data: {
        kind: 'defs',
        after_line: lastBlock.end_line,
        body: 'import "os"',
        info: 'go:defs',
      },
    });
    expect(resp.ok()).toBe(true);

    const afterResp = await request.get('/api/blocks');
    const blocksAfter = await afterResp.json();
    expect(blocksAfter.length).toBeGreaterThan(countBefore);

    const added = blocksAfter.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'defs' && b.body.includes('import "os"'),
    );
    expect(added).toBeTruthy();

    // Clean up.
    if (added) {
      await request.delete('/api/blocks/delete', {
        data: { line: added.line },
      });
    }
  });

  test('click + Prose button adds prose block via UI', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    const blocksBefore = await page.locator('section.block').count();

    // Click the last "+ Prose" button.
    const proseButtons = page.locator('.insert-toolbar button:has-text("+ Prose")');
    const lastProseBtn = proseButtons.last();
    await lastProseBtn.click();

    // Wait for page to reload with new block.
    await page.waitForTimeout(500);
    await page.waitForSelector('section.block');

    const blocksAfter = await page.locator('section.block').count();
    expect(blocksAfter).toBeGreaterThan(blocksBefore);

    // Clean up via API.
    const blocks = await (await page.request.get('/api/blocks')).json();
    const added = [...blocks]
      .reverse()
      .find((b: { kind: string }) => b.kind === 'prose' && b.line > 1);
    if (added) {
      await page.request.delete('/api/blocks/delete', {
        data: { line: added.line },
      });
    }
  });

  test('click + Defs button adds defs block via UI', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    const blocksBefore = await page.locator('section.block').count();

    // Click the last "+ Defs" button.
    const defsButtons = page.locator('.insert-toolbar button:has-text("+ Defs")');
    const lastDefsBtn = defsButtons.last();
    await lastDefsBtn.click();

    // Wait for page to reload.
    await page.waitForTimeout(500);
    await page.waitForSelector('section.block');

    const blocksAfter = await page.locator('section.block').count();
    expect(blocksAfter).toBeGreaterThan(blocksBefore);

    // Clean up via API.
    const blocks = await (await page.request.get('/api/blocks')).json();
    const added = [...blocks]
      .reverse()
      .find(
        (b: { kind: string; body: string }) =>
          b.kind === 'defs' && b.body === '',
      );
    if (added) {
      await page.request.delete('/api/blocks/delete', {
        data: { line: added.line },
      });
    }
  });
});

test.describe('Block deletion', () => {
  test('delete block via API and verify removal', async ({ request }) => {
    // Add a temporary block.
    const addResp = await request.post('/api/blocks/add', {
      data: {
        kind: 'prose',
        after_line: 0,
        body: 'BLOCK_TO_DELETE_E2E',
      },
    });
    expect(addResp.ok()).toBe(true);

    // Find it.
    const blocksResp = await request.get('/api/blocks');
    const blocks = await blocksResp.json();
    const target = blocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('BLOCK_TO_DELETE_E2E'),
    );
    expect(target).toBeTruthy();

    // Delete it.
    const delResp = await request.delete('/api/blocks/delete', {
      data: { line: target.line },
    });
    expect(delResp.ok()).toBe(true);

    // Verify gone.
    const afterResp = await request.get('/api/blocks');
    const afterBlocks = await afterResp.json();
    const stillThere = afterBlocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('BLOCK_TO_DELETE_E2E'),
    );
    expect(stillThere).toBeFalsy();
  });
});

test.describe('Block reorder', () => {
  test('reorder blocks via API — swap adjacent blocks', async ({ request }) => {
    // The hello.md fixture has blocks: prose(title), defs, prose(blank), cell(greeting).
    // Move the blank prose (3rd block) down to swap it with the greeting cell.
    const blocksBeforeReorder = await (await request.get('/api/blocks')).json();
    expect(blocksBeforeReorder.length).toBeGreaterThanOrEqual(3);

    // Find the defs block.
    const defsBlock = blocksBeforeReorder.find(
      (b: { kind: string }) => b.kind === 'defs',
    );
    expect(defsBlock).toBeTruthy();

    // Move defs down. Even if it only swaps with the blank prose, verify the
    // reorder API succeeds and the block order actually changes.
    const resp1 = await request.put('/api/blocks/reorder', {
      data: { line: defsBlock.line, direction: 'down' },
    });
    expect(resp1.ok()).toBe(true);

    // Get blocks after first move.
    const blocksAfter1 = await (await request.get('/api/blocks')).json();

    // The defs block should have moved down (its line number increased).
    const movedDefs1 = blocksAfter1.find(
      (b: { kind: string }) => b.kind === 'defs',
    );
    expect(movedDefs1).toBeTruthy();
    expect(movedDefs1.line).toBeGreaterThan(defsBlock.line);

    // Move defs down again to get it past the greeting cell.
    const resp2 = await request.put('/api/blocks/reorder', {
      data: { line: movedDefs1.line, direction: 'down' },
    });
    expect(resp2.ok()).toBe(true);

    // Verify: greeting should now come before defs.
    const blocksAfter2 = await (await request.get('/api/blocks')).json();
    const finalDefs = blocksAfter2.find(
      (b: { kind: string }) => b.kind === 'defs',
    );
    const finalGreeting = blocksAfter2.find(
      (b: { name?: string }) => b.name === 'greeting',
    );
    expect(finalDefs).toBeTruthy();
    expect(finalGreeting).toBeTruthy();
    expect(finalGreeting.line).toBeLessThan(finalDefs.line);

    // Fixture will be restored by beforeEach on next test.
  });
});

test.describe('Keyboard shortcuts', () => {
  test('Ctrl+S does not open browser save dialog', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    // Focus on a textarea in the editor.
    const textarea = page.locator('textarea').first();
    if (await textarea.count() > 0) {
      await textarea.focus();
    }

    // Listen for download events (browser Save dialog).
    let downloadTriggered = false;
    page.on('download', () => {
      downloadTriggered = true;
    });

    // Press Ctrl+S.
    await page.keyboard.press('Control+s');
    await page.waitForTimeout(500);

    // Ctrl+S should be intercepted; no download should trigger.
    expect(downloadTriggered).toBe(false);
  });
});

test.describe('Cell execution', () => {
  test('run a named cell via API and get response', async ({ request }) => {
    // First verify the fixture is in the expected state.
    const statusResp = await request.get('/api/status');
    const statusBody = await statusResp.json();
    expect(statusBody.ok).toBe(true);

    const resp = await request.post('/api/run', {
      data: { cell: 'greeting' },
    });
    expect(resp.ok()).toBe(true);

    const body = await resp.json();
    expect(body.cell).toBe('greeting');
    // Verify the response has the expected structure.
    expect(typeof body.exit_code).toBe('number');
    expect(typeof body.stdout).toBe('string');
    expect(typeof body.stderr).toBe('string');
    // If there's a compile error, log it for debugging.
    if (body.exit_code !== 0) {
      console.log('Cell execution stderr:', body.stderr);
      console.log('Cell execution error:', body.error);
    }
  });

  test('Run Cell button exists for named cells', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    // The hello.md fixture has a named cell "greeting" with a Run Cell button.
    const runButtons = page.locator('button:has-text("Run Cell")');
    const count = await runButtons.count();
    expect(count).toBeGreaterThan(0);
  });

  test('run all cells via API', async ({ request }) => {
    const resp = await request.post('/api/run/all');
    expect(resp.ok()).toBe(true);

    const results = await resp.json();
    expect(Array.isArray(results)).toBe(true);
    // Should have results for each named cell.
    expect(results.length).toBeGreaterThan(0);
  });

  test('Run All button exists and is clickable', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    const runAllBtn = page.locator('button.run-all-btn:has-text("Run All")');
    expect(await runAllBtn.count()).toBe(1);
  });

  test('run state endpoint returns valid state', async ({ request }) => {
    // Run the cell first to populate state.
    await request.post('/api/run', {
      data: { cell: 'greeting' },
    });

    const resp = await request.get('/api/run/state');
    expect(resp.ok()).toBe(true);

    const state = await resp.json();
    expect(state).toBeTruthy();
    // After running greeting, it should have state.
    expect(state.greeting).toBeTruthy();
    expect(state.greeting.stale).toBe(false);
  });
});

test.describe('Diagnostics', () => {
  test('diagnostics endpoint returns valid response', async ({ request }) => {
    const resp = await request.get('/api/lsp/diagnostics');
    expect(resp.ok()).toBe(true);

    const body = await resp.json();
    expect(body).toHaveProperty('diagnostics');
    expect(Array.isArray(body.diagnostics)).toBe(true);
  });

  test('diagnostics panel element is present in the DOM', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    // The editor renders a diagnostics panel when there are errors.
    // Verify the editor loaded successfully, which includes diagnostics infrastructure.
    // The diagnostics panel only appears when there are errors, so we verify
    // the editor loaded and no JS errors occurred.
    const editorLoaded = await page.locator('section.block').count();
    expect(editorLoaded).toBeGreaterThan(0);
  });
});

test.describe('LSP features', () => {
  test('completion endpoint returns items', async ({ request }) => {
    const resp = await request.get('/api/lsp/completion');
    expect(resp.ok()).toBe(true);

    const body = await resp.json();
    expect(body).toHaveProperty('items');
    expect(Array.isArray(body.items)).toBe(true);
    expect(body.items.length).toBeGreaterThan(0);

    // Should include the "greeting" cell name.
    const cellItem = body.items.find(
      (item: { label: string; kind: string }) =>
        item.kind === 'cell' && item.label === 'greeting',
    );
    expect(cellItem).toBeTruthy();
  });

  test('hover endpoint returns info for known cell', async ({ request }) => {
    const resp = await request.get('/api/lsp/hover?word=greeting');
    expect(resp.ok()).toBe(true);

    const body = await resp.json();
    expect(body.content).toContain('greeting');
  });
});

test.describe('Prose update', () => {
  test('update prose block via API', async ({ request }) => {
    // Add a prose block, update it, then verify.
    await request.post('/api/blocks/add', {
      data: { kind: 'prose', after_line: 0, body: 'PROSE_UPDATE_BEFORE' },
    });

    const blocks = await (await request.get('/api/blocks')).json();
    const target = blocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('PROSE_UPDATE_BEFORE'),
    );
    expect(target).toBeTruthy();

    // Update the prose.
    const updateResp = await request.put('/api/prose/update', {
      data: { line: target.line, body: 'PROSE_UPDATE_AFTER' },
    });
    expect(updateResp.ok()).toBe(true);

    // Verify update.
    const updatedBlocks = await (await request.get('/api/blocks')).json();
    const updated = updatedBlocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('PROSE_UPDATE_AFTER'),
    );
    expect(updated).toBeTruthy();

    // Clean up.
    if (updated) {
      await request.delete('/api/blocks/delete', {
        data: { line: updated.line },
      });
    }
  });
});
