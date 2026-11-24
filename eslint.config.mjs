// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

import tseslint from "@typescript-eslint/eslint-plugin";
import tsparser from "@typescript-eslint/parser";

export default [
	{
		files: ["**/*.ts"],
		languageOptions: {
			parser: tsparser,
			parserOptions: {
				ecmaVersion: "latest",
				sourceType: "module",
			},
		},
		plugins: {
			"@typescript-eslint": tseslint,
		},
		rules: {
			...tseslint.configs.recommended.rules,
			"no-unused-vars": "off",
			"@typescript-eslint/no-unused-vars": "error",
			"@typescript-eslint/explicit-function-return-type": "error",
			"@typescript-eslint/no-explicit-any": "error",
		},
	},
	{
		ignores: ["node_modules/", "vscode/"],
	},
];
