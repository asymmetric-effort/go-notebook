// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package crypto

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// MinPassphraseLength is the minimum number of Unicode code points required.
const MinPassphraseLength = 12

// ValidatePassphrase checks that a passphrase meets the minimum length
// requirement (12 Unicode code points, not bytes).
func ValidatePassphrase(passphrase string) error {
	length := utf8.RuneCountInString(passphrase)
	if length < MinPassphraseLength {
		return fmt.Errorf("passphrase is %d characters, minimum is %d", length, MinPassphraseLength)
	}
	return nil
}

// ReadPassphraseFromFile reads a passphrase from a file, trimming the
// trailing newline.
func ReadPassphraseFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read passphrase file: %w", err)
	}
	return strings.TrimRight(string(data), "\n\r"), nil
}

// PromptPassphrase prompts for a passphrase on the terminal with echo
// disabled. If confirm is true, prompts twice and rejects mismatches.
func PromptPassphrase(confirm bool) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", fmt.Errorf("stdin is not a terminal. Use --passphrase-file for unattended use")
	}

	fmt.Fprint(os.Stderr, "Enter passphrase: ")
	pass1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read passphrase: %w", err)
	}

	if confirm {
		fmt.Fprint(os.Stderr, "Confirm passphrase: ")
		pass2, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("read confirmation: %w", err)
		}
		if string(pass1) != string(pass2) {
			return "", fmt.Errorf("passphrases do not match")
		}
	}

	return string(pass1), nil
}
