// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Grammar snapshot tests for the TextMate injection grammar.
// These test the regex patterns in the grammar directly.

import * as assert from "assert";
import * as fs from "fs";
import * as path from "path";

interface GrammarPattern {
	begin?: string;
	end?: string;
	name?: string;
}

interface Grammar {
	scopeName: string;
	injectionSelector: string;
	repository: Record<string, GrammarPattern>;
}

function loadGrammar(): Grammar {
	const grammarPath = path.join(__dirname, "..", "..", "syntaxes", "gobook.tmLanguage.json");
	const content = fs.readFileSync(grammarPath, "utf-8");
	return JSON.parse(content) as Grammar;
}

describe("GoBook TextMate Grammar", () => {
	const grammar = loadGrammar();

	it("has correct scope name", () => {
		assert.strictEqual(grammar.scopeName, "source.gobook.injection");
	});

	it("injects into Markdown", () => {
		assert.strictEqual(grammar.injectionSelector, "L:text.html.markdown");
	});

	it("has notebook block pattern", () => {
		const pattern = grammar.repository["gobook-notebook-block"];
		assert.ok(pattern, "notebook block pattern should exist");
		assert.ok(pattern.begin, "notebook block should have begin pattern");
	});

	it("has defs block pattern", () => {
		const pattern = grammar.repository["gobook-defs-block"];
		assert.ok(pattern, "defs block pattern should exist");
		assert.ok(pattern.begin, "defs block should have begin pattern");
	});

	it("notebook pattern matches go:notebook fence", () => {
		const pattern = grammar.repository["gobook-notebook-block"];
		const regex = new RegExp(pattern.begin as string, "m");
		assert.ok(regex.test("```go:notebook name=test exported(val int)"));
		assert.ok(regex.test("```go:notebook"));
		assert.ok(regex.test("````go:notebook name=test"));
	});

	it("notebook pattern does not match plain go fence", () => {
		const pattern = grammar.repository["gobook-notebook-block"];
		const regex = new RegExp(pattern.begin as string, "m");
		assert.ok(!regex.test("```go"), "should not match plain go fence");
		assert.ok(!regex.test("```python"), "should not match python fence");
		assert.ok(!regex.test("```"), "should not match bare fence");
	});

	it("defs pattern matches go:defs fence", () => {
		const pattern = grammar.repository["gobook-defs-block"];
		const regex = new RegExp(pattern.begin as string, "m");
		assert.ok(regex.test("```go:defs"));
		assert.ok(regex.test("````go:defs"));
	});

	it("defs pattern does not match notebook fence", () => {
		const pattern = grammar.repository["gobook-defs-block"];
		const regex = new RegExp(pattern.begin as string, "m");
		assert.ok(!regex.test("```go:notebook"), "should not match notebook fence");
	});

	it("closing fence pattern exists", () => {
		const pattern = grammar.repository["gobook-notebook-block"];
		assert.ok(pattern.end, "notebook block should have end pattern");
		// The end pattern uses TextMate \\G anchors which don't translate
		// directly to JavaScript regex. We verify the pattern string contains
		// the backtick matching group.
		assert.ok((pattern.end as string).includes("\\3"), "end pattern should reference opening fence length");
	});

	it("notebook block has Go content name", () => {
		const pattern = grammar.repository["gobook-notebook-block"];
		assert.strictEqual(
			(pattern as Record<string, string>)["contentName"],
			"meta.embedded.block.gobook",
		);
	});
});
