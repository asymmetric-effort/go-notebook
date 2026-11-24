// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import * as vscode from "vscode";

/**
 * Activate the Go Notebook extension.
 * Currently provides syntax highlighting via the TextMate grammar.
 * LSP connection will be added in a future story.
 */
export function activate(context: vscode.ExtensionContext): void {
	const outputChannel = vscode.window.createOutputChannel("Go Notebook");
	outputChannel.appendLine("Go Notebook extension activated");
	context.subscriptions.push(outputChannel);
}

/**
 * Deactivate the extension.
 */
export function deactivate(): void {
	// Nothing to clean up yet.
}
