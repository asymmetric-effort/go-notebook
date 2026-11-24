// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { test, expect } from "@playwright/test";

test.describe("Editor Page", () => {
	test("loads and shows notebook blocks", async ({ page }) => {
		await page.goto("/api/editor");
		await expect(page.locator("h1")).toContainText("go-notebook editor");

		// Should have at least one block section.
		const blocks = page.locator("section.block");
		await expect(blocks.first()).toBeVisible();
	});

	test("shows prose blocks", async ({ page }) => {
		await page.goto("/api/editor");
		const proseBlocks = page.locator("section.prose-block");
		// The multicell notebook has prose between cells.
		await expect(proseBlocks.first()).toBeVisible();
	});

	test("shows cell blocks with names", async ({ page }) => {
		await page.goto("/api/editor");
		const cellBlocks = page.locator("section.cell-block");
		await expect(cellBlocks.first()).toBeVisible();

		// Should show cell names.
		const heading = cellBlocks.first().locator("h2");
		await expect(heading).toContainText("Cell:");
	});

	test("cell has Run Cell button", async ({ page }) => {
		await page.goto("/api/editor");
		const runBtn = page.locator("button.run-btn").first();
		await expect(runBtn).toBeVisible();
		await expect(runBtn).toContainText("Run Cell");
	});

	test("has Run All button in toolbar", async ({ page }) => {
		await page.goto("/api/editor");
		const runAllBtn = page.locator("button.run-all-btn");
		await expect(runAllBtn).toBeVisible();
		await expect(runAllBtn).toContainText("Run All");
	});

	test("has secret prompt section", async ({ page }) => {
		await page.goto("/api/editor");
		const secretBlock = page.locator("section.secret-block");
		await expect(secretBlock).toBeVisible();

		// Should have a masked input.
		const passwordInput = secretBlock.locator("input[type='password']");
		await expect(passwordInput).toBeVisible();
	});
});

test.describe("Editor CSP", () => {
	test("CSP header is present", async ({ page }) => {
		const response = await page.goto("/api/editor");
		const csp = response?.headers()["content-security-policy"];
		expect(csp).toBeDefined();
		expect(csp).toContain("default-src 'self'");
		expect(csp).toContain("script-src 'self'");
	});

	test("X-Content-Type-Options header is set", async ({ page }) => {
		const response = await page.goto("/api/editor");
		expect(response?.headers()["x-content-type-options"]).toBe("nosniff");
	});
});

test.describe("Editor API", () => {
	test("status endpoint returns OK", async ({ request }) => {
		const response = await request.get("/api/status");
		expect(response.ok()).toBeTruthy();
		const body = await response.json();
		expect(body.ok).toBe(true);
	});

	test("blocks endpoint returns blocks", async ({ request }) => {
		const response = await request.get("/api/blocks");
		expect(response.ok()).toBeTruthy();
		const blocks = await response.json();
		expect(Array.isArray(blocks)).toBeTruthy();
		expect(blocks.length).toBeGreaterThan(0);
	});

	test("diagnostics endpoint works", async ({ request }) => {
		const response = await request.get("/api/lsp/diagnostics");
		expect(response.ok()).toBeTruthy();
		const body = await response.json();
		expect(body).toHaveProperty("diagnostics");
	});

	test("completion endpoint returns items", async ({ request }) => {
		const response = await request.get("/api/lsp/completion");
		expect(response.ok()).toBeTruthy();
		const body = await response.json();
		expect(body).toHaveProperty("items");
		expect(body.items.length).toBeGreaterThan(0);
	});
});

test.describe("Editor Responsiveness", () => {
	test("desktop viewport renders all blocks", async ({ page }) => {
		await page.setViewportSize({ width: 1280, height: 720 });
		await page.goto("/api/editor");
		const blocks = page.locator("section.block");
		await expect(blocks.first()).toBeVisible();

		// Measure page load time.
		const timing = await page.evaluate(() => {
			const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming;
			return {
				domContentLoaded: nav.domContentLoadedEventEnd - nav.startTime,
				loadComplete: nav.loadEventEnd - nav.startTime,
			};
		});
		console.log(`Desktop: DOMContentLoaded=${timing.domContentLoaded}ms, Load=${timing.loadComplete}ms`);
	});

	test("mobile viewport renders all blocks", async ({ page }) => {
		await page.setViewportSize({ width: 393, height: 851 });
		await page.goto("/api/editor");
		const blocks = page.locator("section.block");
		await expect(blocks.first()).toBeVisible();

		const timing = await page.evaluate(() => {
			const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming;
			return {
				domContentLoaded: nav.domContentLoadedEventEnd - nav.startTime,
				loadComplete: nav.loadEventEnd - nav.startTime,
			};
		});
		console.log(`Mobile: DOMContentLoaded=${timing.domContentLoaded}ms, Load=${timing.loadComplete}ms`);
	});
});
