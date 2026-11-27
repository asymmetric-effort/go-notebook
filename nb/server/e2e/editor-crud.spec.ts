// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { test, expect } from '@playwright/test';

test.describe('Editor CRUD — API tests', () => {
  // These tests use the hello.md fixture served by the webServer config.
  // They test the new CRUD endpoints via the API.

  test('POST /api/blocks/add with kind=cell adds a cell', async ({ request }) => {
    // First, get current blocks.
    const before = await request.get('/api/blocks');
    const blocksBefore = await before.json();
    const countBefore = blocksBefore.length;

    // Add a cell after the last block.
    const lastBlock = blocksBefore[blocksBefore.length - 1];
    const resp = await request.post('/api/blocks/add', {
      data: {
        kind: 'cell',
        after_line: lastBlock.end_line,
        body: 'testVal := 123',
        info: 'go:notebook name=testcrud',
      },
    });
    expect(resp.ok()).toBe(true);

    // Verify block was added.
    const after = await request.get('/api/blocks');
    const blocksAfter = await after.json();
    expect(blocksAfter.length).toBeGreaterThan(countBefore);

    // Clean up: delete the added block.
    const addedBlock = blocksAfter.find(
      (b: { kind: string; name?: string }) => b.kind === 'cell' && b.name === 'testcrud',
    );
    if (addedBlock) {
      await request.delete('/api/blocks/delete', {
        data: { line: addedBlock.line },
      });
    }
  });

  test('DELETE /api/blocks/delete removes a block', async ({ request }) => {
    // Add a block to delete.
    const addResp = await request.post('/api/blocks/add', {
      data: {
        kind: 'prose',
        after_line: 0,
        body: 'DELETEME prose block',
      },
    });
    expect(addResp.ok()).toBe(true);

    // Find it.
    const blocksResp = await request.get('/api/blocks');
    const blocks = await blocksResp.json();
    const target = blocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('DELETEME'),
    );
    expect(target).toBeTruthy();

    // Delete it.
    const delResp = await request.delete('/api/blocks/delete', {
      data: { line: target.line },
    });
    expect(delResp.ok()).toBe(true);

    // Verify it's gone.
    const afterResp = await request.get('/api/blocks');
    const afterBlocks = await afterResp.json();
    const found = afterBlocks.find(
      (b: { kind: string; body: string }) =>
        b.kind === 'prose' && b.body.includes('DELETEME'),
    );
    expect(found).toBeFalsy();
  });

  test('POST /api/notebook/init on non-empty file returns error', async ({ request }) => {
    const resp = await request.post('/api/notebook/init');
    // hello.md is non-empty, so this should fail.
    expect(resp.status()).toBe(409);
  });
});

test.describe('Editor CRUD — UI tests', () => {
  test('non-empty notebook shows insert toolbars', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    // Should have insert toolbars with "+ Cell" buttons.
    const insertButtons = page.locator('.insert-toolbar button');
    const count = await insertButtons.count();
    expect(count).toBeGreaterThan(0);

    // Check for "+ Cell" text in at least one button.
    const cellBtnCount = await page.locator('.insert-toolbar button:has-text("+ Cell")').count();
    expect(cellBtnCount).toBeGreaterThan(0);
  });

  test('blocks have delete buttons', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    const deleteButtons = page.locator('.block-actions .delete-btn');
    const count = await deleteButtons.count();
    expect(count).toBeGreaterThan(0);
  });

  test('blocks have reorder buttons', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    // At least some blocks should have up/down buttons.
    const blockActions = page.locator('.block-actions button');
    const count = await blockActions.count();
    // With 3+ blocks, there should be several action buttons.
    expect(count).toBeGreaterThan(2);
  });

  test('click + Cell adds a cell block', async ({ page }) => {
    await page.goto('/editor');
    await page.waitForSelector('section.block');

    const blocksBefore = await page.locator('section.block').count();

    // Click the last "+ Cell" button.
    const cellButtons = page.locator('.insert-toolbar button:has-text("+ Cell")');
    const lastCellBtn = cellButtons.last();
    await lastCellBtn.click();

    // Wait for reload.
    await page.waitForTimeout(500);
    await page.waitForSelector('section.block');

    const blocksAfter = await page.locator('section.block').count();
    expect(blocksAfter).toBeGreaterThan(blocksBefore);

    // Clean up: delete the new cell block (the last cell-block).
    // Use API directly to avoid dialog issues.
    const blocks = await (await page.request.get('/api/blocks')).json();
    const lastCell = [...blocks]
      .reverse()
      .find((b: { kind: string; name?: string }) => b.kind === 'cell' && b.name === 'cell1');
    if (lastCell) {
      await page.request.delete('/api/blocks/delete', {
        data: { line: lastCell.line },
      });
    }
  });
});
