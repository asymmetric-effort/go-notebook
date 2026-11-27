// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// version manages semantic version tags.
//
// Usage:
//
//	go run ./tools/version          # print current version
//	go run ./tools/version major    # bump major, tag, push
//	go run ./tools/version minor    # bump minor, tag, push
//	go run ./tools/version patch    # bump patch, tag, push
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "version: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	current, err := currentVersion()
	if err != nil {
		return err
	}

	if len(os.Args) < 2 {
		fmt.Println(current)
		return nil
	}

	bump := os.Args[1]
	next, err := bumpVersion(current, bump)
	if err != nil {
		return err
	}

	tag := "v" + next
	fmt.Printf("%s → %s\n", current, tag)

	if err := gitTag(tag); err != nil {
		return err
	}
	if err := gitPush(tag); err != nil {
		return err
	}

	fmt.Printf("Tagged and pushed %s\n", tag)
	return nil
}

// currentVersion returns the latest semver tag or "v0.0.0" if none exist.
func currentVersion() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match", "v*")
	out, err := cmd.Output()
	if err != nil {
		return "v0.0.0", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// bumpVersion increments the given semver component.
func bumpVersion(current, bump string) (string, error) {
	v := strings.TrimPrefix(current, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version %q", current)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid major %q: %w", parts[0], err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid minor %q: %w", parts[1], err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch %q: %w", parts[2], err)
	}

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", fmt.Errorf("unknown bump %q (use major, minor, or patch)", bump)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

// gitTag creates an annotated tag.
func gitTag(tag string) error {
	cmd := exec.Command("git", "tag", "-a", tag, "-m", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitPush pushes the tag to origin.
func gitPush(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
