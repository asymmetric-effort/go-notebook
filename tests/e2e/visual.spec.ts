// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { test, expect } from "@playwright/test";

test.describe("Visual Regression", () => {
	test("editor desktop screenshot", async ({ page }) => {
		await page.setViewportSize({ width: 1280, height: 720 });
		await page.goto("/api/editor");
		// Wait for blocks to load.
		await page.locator("section.block").first().waitFor();
		await expect(page).toHaveScreenshot("editor-desktop.png", {
			maxDiffPixelRatio: 0.01,
		});
	});

	test("editor mobile screenshot", async ({ page }) => {
		await page.setViewportSize({ width: 393, height: 851 });
		await page.goto("/api/editor");
		await page.locator("section.block").first().waitFor();
		await expect(page).toHaveScreenshot("editor-mobile.png", {
			maxDiffPixelRatio: 0.01,
		});
	});
});
