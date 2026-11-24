// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import * as vscode from "vscode";
import { execSync } from "child_process";
import * as fs from "fs";

/**
 * Find the gobook binary. Checks in order:
 * 1. goNotebook.gobookPath setting
 * 2. PATH lookup via `which gobook` / `where gobook`
 *
 * Returns the path to gobook, or undefined if not found.
 */
export function findGobook(): string | undefined {
	// Check the user setting first.
	const config = vscode.workspace.getConfiguration("goNotebook");
	const configPath = config.get<string>("gobookPath", "");
	if (configPath && configPath !== "gobook") {
		if (fs.existsSync(configPath)) {
			return configPath;
		}
	}

	// Look in PATH.
	try {
		const cmd = process.platform === "win32" ? "where gobook" : "which gobook";
		const result = execSync(cmd, { encoding: "utf-8", timeout: 5000 }).trim();
		if (result && fs.existsSync(result)) {
			return result;
		}
	} catch {
		// Not found in PATH.
	}

	return undefined;
}
