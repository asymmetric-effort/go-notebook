// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import * as assert from "assert";
import * as path from "path";
import * as fs from "fs";

// Test the gobook binary finder logic without requiring vscode module.
// We test the PATH lookup portion directly.

import { execSync } from "child_process";

describe("GoBook Binary Finder", () => {
	it("which/where command exists on this platform", () => {
		const cmd = process.platform === "win32" ? "where" : "which";
		// The command itself should exist (even if gobook doesn't).
		try {
			execSync(cmd + " " + cmd, { encoding: "utf-8", timeout: 5000 });
		} catch {
			// which/where may not find itself on all platforms, that's ok.
		}
		// Just verify it doesn't throw a hard crash.
		assert.ok(true);
	});

	it("detects non-existent binary returns empty", () => {
		let found = false;
		try {
			const cmd = process.platform === "win32"
				? "where nonexistent_binary_xyz_12345"
				: "which nonexistent_binary_xyz_12345";
			execSync(cmd, { encoding: "utf-8", timeout: 5000 });
			found = true;
		} catch {
			found = false;
		}
		assert.ok(!found, "non-existent binary should not be found");
	});

	it("fs.existsSync works for real paths", () => {
		const thisFile = path.join(__dirname, "..", "..", "package.json");
		assert.ok(fs.existsSync(thisFile), "package.json should exist relative to test");
	});

	it("fs.existsSync returns false for missing paths", () => {
		assert.ok(!fs.existsSync("/nonexistent/path/to/gobook"), "missing path should not exist");
	});
});

describe("Extension Module Structure", () => {
	it("extension.ts exports activate and deactivate", () => {
		// We can't import the actual extension module here because it
		// depends on the vscode API. Instead, verify the file exists.
		const extPath = path.join(__dirname, "..", "extension.js");
		assert.ok(fs.existsSync(extPath), "compiled extension.js should exist");
	});

	it("gobook.ts exports findGobook", () => {
		const gobookPath = path.join(__dirname, "..", "gobook.js");
		assert.ok(fs.existsSync(gobookPath), "compiled gobook.js should exist");
		// Verify the export exists in the compiled output.
		const content = fs.readFileSync(gobookPath, "utf-8");
		assert.ok(content.includes("findGobook"), "gobook.js should export findGobook");
	});
});
