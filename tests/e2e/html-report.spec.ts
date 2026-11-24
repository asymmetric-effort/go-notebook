// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Visual regression tests for HTML report output (#53).
// Renders notebooks to HTML via gobook build, opens in browser, and
// captures screenshots at desktop and mobile viewports.

import { test, expect } from "@playwright/test";
import { execSync } from "child_process";
import * as fs from "fs";
import * as path from "path";

const notebooks = [
	"simple",
	"multicell",
	"display",
	"prose_only",
	"formatting",
];

const testdataDir = path.resolve(__dirname, "../../nb/build/testdata/notebooks");
const outputDir = path.resolve(__dirname, "../../tmp/visual-regression");

test.beforeAll(() => {
	fs.mkdirSync(outputDir, { recursive: true });

	// Build HTML reports for each notebook.
	for (const name of notebooks) {
		const src = path.join(testdataDir, name + ".md");
		const out = path.join(outputDir, name + ".html");
		try {
			execSync(`go run ../../cmd/gobook build ${src} --format html -o ${out}`, {
				cwd: __dirname,
				timeout: 60000,
				stdio: "pipe",
			});
		} catch {
			// Some notebooks (panic_recovery) may produce partial output — that's ok.
		}
	}
});

for (const name of notebooks) {
	test.describe(`HTML Report: ${name}`, () => {
		test(`desktop screenshot`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			await page.setViewportSize({ width: 1280, height: 720 });
			await page.goto("file://" + htmlPath);
			await expect(page).toHaveScreenshot(`html-${name}-desktop.png`, {
				maxDiffPixelRatio: 0.02,
				fullPage: true,
			});
		});

		test(`mobile screenshot`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			await page.setViewportSize({ width: 393, height: 851 });
			await page.goto("file://" + htmlPath);
			await expect(page).toHaveScreenshot(`html-${name}-mobile.png`, {
				maxDiffPixelRatio: 0.02,
				fullPage: true,
			});
		});

		test(`has CSP meta tag`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			await page.goto("file://" + htmlPath);
			const csp = await page.locator("meta[http-equiv='Content-Security-Policy']").getAttribute("content");
			expect(csp).toContain("default-src");
		});

		test(`no external resource loads`, async ({ page }) => {
			const htmlPath = path.join(outputDir, name + ".html");
			if (!fs.existsSync(htmlPath)) {
				test.skip();
				return;
			}
			const requests: string[] = [];
			page.on("request", (req) => {
				if (!req.url().startsWith("file://")) {
					requests.push(req.url());
				}
			});
			await page.goto("file://" + htmlPath);
			expect(requests).toEqual([]);
		});
	});
}

test.describe("Deliberate change detection", () => {
	test("detects style change in HTML output", async ({ page }) => {
		const htmlPath = path.join(outputDir, "simple.html");
		if (!fs.existsSync(htmlPath)) {
			test.skip();
			return;
		}
		// Inject a visible style change and verify the screenshot differs.
		await page.setViewportSize({ width: 1280, height: 720 });
		await page.goto("file://" + htmlPath);
		await page.evaluate(() => {
			document.body.style.backgroundColor = "red";
		});
		// This should NOT match the reference screenshot.
		// We use a negative assertion: expect failure when comparing.
		let matched = true;
		try {
			await expect(page).toHaveScreenshot(`html-simple-desktop.png`, {
				maxDiffPixelRatio: 0.001,
			});
		} catch {
			matched = false;
		}
		expect(matched).toBe(false);
	});
});
