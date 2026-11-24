// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import * as vscode from "vscode";
import { LanguageClient, LanguageClientOptions, ServerOptions } from "vscode-languageclient/node";
import { findGobook } from "./gobook";

let client: LanguageClient | undefined;

/**
 * Activate the Go Notebook extension.
 * Starts the gobook lsp language server for completion, hover, and diagnostics.
 */
export async function activate(context: vscode.ExtensionContext): Promise<void> {
	const outputChannel = vscode.window.createOutputChannel("Go Notebook");
	outputChannel.appendLine("Go Notebook extension activated");
	context.subscriptions.push(outputChannel);

	const gobookPath = findGobook();
	if (!gobookPath) {
		const msg = "gobook binary not found. Install it or set goNotebook.gobookPath in settings.";
		outputChannel.appendLine(msg);
		void vscode.window.showWarningMessage(msg);
		return;
	}

	outputChannel.appendLine("Using gobook: " + gobookPath);

	const serverOptions: ServerOptions = {
		command: gobookPath,
		args: ["lsp"],
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [
			{ scheme: "file", language: "markdown" },
		],
		outputChannel,
	};

	client = new LanguageClient(
		"goNotebook",
		"Go Notebook LSP",
		serverOptions,
		clientOptions,
	);

	try {
		await client.start();
		outputChannel.appendLine("Go Notebook LSP started");
	} catch (err: unknown) {
		const msg = err instanceof Error ? err.message : String(err);
		outputChannel.appendLine("Failed to start LSP: " + msg);
		void vscode.window.showErrorMessage("Go Notebook: failed to start language server: " + msg);
	}
}

/**
 * Deactivate the extension. Stops the language client.
 */
export async function deactivate(): Promise<void> {
	if (client) {
		await client.stop();
		client = undefined;
	}
}
