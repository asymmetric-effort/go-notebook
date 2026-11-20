// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package nb

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// PromptFunc is the function used to prompt for secrets on a terminal.
// It can be replaced for testing.
var PromptFunc = defaultPrompt

func defaultPrompt(name string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("stdin is not a terminal, cannot prompt for secret %q", name)
	}
	fmt.Fprintf(os.Stderr, "Enter secret %q: ", name)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read secret %q: %w", name, err)
	}
	return string(password), nil
}

// ResolveSecret reads a secret from the first available source:
// 1. Environment variable GOBOOK_SECRET_<NAME> (uppercased)
// 2. Secrets file <notebook>.secrets (must be owner-only permissions)
// 3. Terminal prompt (stdin must be a terminal)
//
// Returns the secret value or an error naming the secret and sources tried.
func ResolveSecret(name string, notebookPath string) (string, error) {
	// Source 1: Environment variable.
	envKey := "GOBOOK_SECRET_" + strings.ToUpper(name)
	if val := os.Getenv(envKey); val != "" {
		return val, nil
	}

	// Source 2: Secrets file.
	secretsFile := notebookPath + ".secrets"
	val, err := readSecretsFile(secretsFile, name)
	if err == nil && val != "" {
		return val, nil
	}

	// Source 3: Terminal prompt.
	val, err = PromptFunc(name)
	if err == nil {
		return val, nil
	}

	return "", fmt.Errorf("secret %q not found. Set %s, add it to %s, or run with a terminal for a prompt",
		name, envKey, secretsFile)
}

// readSecretsFile reads a secret from a .secrets file. The file format is
// one NAME=VALUE per line. The file must have owner-only permissions.
func readSecretsFile(path string, name string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err // File doesn't exist — that's fine, try next source.
	}

	// Check permissions: must be owner-only (0600 or 0400).
	// Skip on Windows where Unix permissions don't apply.
	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			return "", fmt.Errorf("secrets file %s has permissions %o, must be owner-only (0600). Fix with: chmod 600 %s",
				path, mode, path)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == name {
			return strings.TrimSpace(parts[1]), nil
		}
	}
	return "", fmt.Errorf("secret %q not found in %s", name, path)
}
