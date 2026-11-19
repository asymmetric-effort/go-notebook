// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// jsonlint checks JSON files for valid syntax and canonical formatting.
// Canonical formatting is json.Indent with a tab indent and no prefix.
// The linter never modifies files.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// canonicalJSON returns the canonical formatting of a JSON document.
// Canonical is defined as json.Indent with tab indentation.
func canonicalJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "\t"); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// lintFile checks a single JSON file for syntax and formatting.
// It returns an error describing the problem, or nil if the file is valid
// and canonically formatted.
func lintFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// Trim trailing whitespace for consistent comparison.
	trimmed := bytes.TrimRight(data, " \t\n\r")

	if !json.Valid(trimmed) {
		offset := findSyntaxErrorOffset(trimmed)
		line, col := offsetToLineCol(data, offset)
		return fmt.Errorf("%s:%d:%d: invalid JSON syntax", path, line, col)
	}

	canonical, err := canonicalJSON(trimmed)
	if err != nil {
		return fmt.Errorf("%s: format: %w", path, err)
	}

	if !bytes.Equal(data, canonical) {
		return fmt.Errorf("%s: not in canonical format", path)
	}

	return nil
}

// findSyntaxErrorOffset attempts to find the byte offset of a JSON syntax error.
func findSyntaxErrorOffset(data []byte) int64 {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		if synErr, ok := err.(*json.SyntaxError); ok {
			return synErr.Offset
		}
	}
	return 0
}

// offsetToLineCol converts a byte offset to a 1-based line and column.
func offsetToLineCol(data []byte, offset int64) (int, int) {
	if offset <= 0 || int(offset) > len(data) {
		return 1, 1
	}
	line := 1
	col := 1
	for i := int64(0); i < offset; i++ {
		if data[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// findJSONFiles walks dir and returns all .json files.
func findJSONFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".json" {
			base := filepath.Base(path)
			// Skip npm-generated files which use their own formatting convention.
			if base == "package-lock.json" {
				return nil
			}
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func run(args []string) int {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	files, err := findJSONFiles(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	failed := false
	for _, f := range files {
		if err := lintFile(f); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			failed = true
		}
	}

	if failed {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:]))
}
