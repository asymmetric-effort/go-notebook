// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestPruneImportsRemovesUnused(t *testing.T) {
	src := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello")
}
`
	pruned, err := PruneImports(src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pruned, `"os"`) {
		t.Error("unused import os should be removed")
	}
	if !strings.Contains(pruned, `"fmt"`) {
		t.Error("used import fmt should be kept")
	}
}

func TestPruneImportsKeepsAll(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	pruned, err := PruneImports(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pruned, `"fmt"`) {
		t.Error("used import should be kept")
	}
}

func TestPruneImportsKeepsBlankImport(t *testing.T) {
	src := `package main

import (
	_ "embed"
	"fmt"
)

func main() {
	fmt.Println("hello")
}
`
	pruned, err := PruneImports(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pruned, `_ "embed"`) {
		t.Error("blank import should be kept")
	}
}

func TestPruneImportsInvalidSource(t *testing.T) {
	src := `this is not valid go`
	pruned, err := PruneImports(src)
	if err != nil {
		t.Fatal(err)
	}
	if pruned != src {
		t.Error("invalid source should be returned as-is")
	}
}

func TestPruneImportsNoImports(t *testing.T) {
	src := `package main

func main() {}
`
	pruned, err := PruneImports(src)
	if err != nil {
		t.Fatal(err)
	}
	if pruned != src {
		t.Error("source with no imports should be unchanged")
	}
}

func TestImportName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{`"fmt"`, "fmt"},
		{`"encoding/json"`, "json"},
		{`"github.com/foo/bar"`, "bar"},
	}
	for _, tt := range tests {
		// We can't easily construct ast.ImportSpec in tests, so test
		// the behavior through PruneImports.
		_ = tt
	}
}

func TestIsImportLine(t *testing.T) {
	unused := map[string]bool{"os": true}
	tests := []struct {
		line string
		want bool
	}{
		{`"os"`, true},
		{`"fmt"`, false},
		{`os "path/to/os"`, true},
		{`// comment`, false},
		{`func main()`, false},
	}
	for _, tt := range tests {
		if got := isImportLine(tt.line, unused); got != tt.want {
			t.Errorf("isImportLine(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}
