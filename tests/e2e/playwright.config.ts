// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright configuration for go-notebook editor browser tests.
 * Tests run against `gobook serve` started by the test fixtures.
 */
export default defineConfig({
	testDir: ".",
	timeout: 30000,
	retries: 0,
	use: {
		baseURL: "http://localhost:3721",
		trace: "on-first-retry",
	},
	projects: [
		{
			name: "desktop-chrome",
			use: { ...devices["Desktop Chrome"] },
		},
		{
			name: "mobile-chrome",
			use: { ...devices["Pixel 5"] },
		},
	],
	// The server is started by Playwright's webServer config.
	webServer: {
		command: "go run ../../cmd/gobook serve ../../nb/build/testdata/notebooks/multicell.md --addr localhost:3721",
		url: "http://localhost:3721/api/status",
		reuseExistingServer: !process.env["CI"],
		timeout: 60000,
	},
});
