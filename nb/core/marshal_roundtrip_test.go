// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMarshalRoundTrip generates a package that marshals a cell struct,
// unmarshals it, and verifies the values survived the round-trip.
// This is a real integration test — the marshal methods are compiled
// and executed, not just string-compared.
func TestMarshalRoundTrip(t *testing.T) {
	src := "```go:notebook name=data exported(val int, msg string)\nval = 42\nmsg = \"hello\"\n```\n"

	result := Parse(src)
	gen, err := Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "marshalpkg")
	err = WritePackage(gen, pkgDir, "test.md", "marshalpkg")
	if err != nil {
		t.Fatal(err)
	}

	// Read the generated main.go and replace the main function with one
	// that tests the round-trip.
	mainPath := filepath.Join(pkgDir, "main.go")
	mainSrc, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	// Verify marshal methods are present.
	if !strings.Contains(string(mainSrc), "MarshalBinary") {
		t.Fatal("generated code missing MarshalBinary method")
	}
	if !strings.Contains(string(mainSrc), "UnmarshalBinary") {
		t.Fatal("generated code missing UnmarshalBinary method")
	}

	// Replace the entire main function with a round-trip test.
	// Find "func main() {" and everything after it, replace completely.
	mainIdx := strings.Index(string(mainSrc), "func main() {")
	if mainIdx < 0 {
		t.Fatal("could not find func main() in generated code")
	}
	src2 := string(mainSrc)[:mainIdx] + `func main() {
	// Test marshal round-trip.
	original := dataCell{val: 42, msg: "hello"}
	data, err := original.MarshalBinary()
	if err != nil {
		panic("marshal: " + err.Error())
	}
	var restored dataCell
	if err := restored.UnmarshalBinary(data); err != nil {
		panic("unmarshal: " + err.Error())
	}
	if restored.val != 42 {
		panic(fmt.Sprintf("val = %d, want 42", restored.val))
	}
	if restored.msg != "hello" {
		panic(fmt.Sprintf("msg = %q, want hello", restored.msg))
	}
	fmt.Println("ROUNDTRIP_OK")
}
`
	os.WriteFile(mainPath, []byte(src2), 0o644)

	// Add replace directive.
	moduleRoot := findModuleRoot(t)
	goMod := "module marshalpkg\n\ngo 1.27.4\n\nrequire github.com/asymmetric-effort/go-notebook v0.0.0\n\n" +
		"replace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0o644)

	// Build and run.
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(dir, "marshaltest"), ".")
	buildCmd.Dir = pkgDir
	buildCmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		genSrc, _ := os.ReadFile(mainPath)
		t.Fatalf("go build failed:\n%s\n---generated---\n%s", out, genSrc)
	}

	runCmd := exec.Command(filepath.Join(dir, "marshaltest"))
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("binary failed:\n%s\n%v", runOut, err)
	}

	if !strings.Contains(string(runOut), "ROUNDTRIP_OK") {
		t.Errorf("round-trip test did not print ROUNDTRIP_OK.\nOutput: %s", runOut)
	}
	t.Logf("Marshal round-trip verified: %s", strings.TrimSpace(string(runOut)))
}
