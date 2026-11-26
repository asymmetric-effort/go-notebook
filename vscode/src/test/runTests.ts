// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Test runner for grammar and extension tests.
// This file is the entry point for `bun run test`.

import Mocha from "mocha";
import * as path from "path";
import * as fs from "fs";

function run(): void {
	const mocha = new Mocha({ ui: "bdd", color: true });
	const testsRoot = path.resolve(__dirname);

	const files = fs.readdirSync(testsRoot).filter((f: string) => f.endsWith(".test.js"));
	for (const f of files) {
		mocha.addFile(path.resolve(testsRoot, f));
	}

	mocha.run((failures: number) => {
		if (failures > 0) {
			process.exit(1);
		}
	});
}

run();
