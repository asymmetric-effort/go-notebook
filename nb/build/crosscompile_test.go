// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package build

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestDefaultCrossCompile(t *testing.T) {
	opts := DefaultCrossCompile()
	if opts.GOOS != runtime.GOOS {
		t.Errorf("GOOS = %q, want %q", opts.GOOS, runtime.GOOS)
	}
	if opts.GOARCH != runtime.GOARCH {
		t.Errorf("GOARCH = %q, want %q", opts.GOARCH, runtime.GOARCH)
	}
	if opts.CGOEnabled != "1" {
		t.Errorf("CGOEnabled = %q, want %q", opts.CGOEnabled, "1")
	}
}

func TestSetEnv(t *testing.T) {
	opts := CrossCompileOptions{GOOS: "linux", GOARCH: "arm64", CGOEnabled: "0"}
	cmd := exec.Command("true")
	opts.SetEnv(cmd)
	found := map[string]bool{}
	for _, e := range cmd.Env {
		if e == "GOOS=linux" {
			found["GOOS"] = true
		}
		if e == "GOARCH=arm64" {
			found["GOARCH"] = true
		}
		if e == "CGO_ENABLED=0" {
			found["CGO_ENABLED"] = true
		}
	}
	for _, k := range []string{"GOOS", "GOARCH", "CGO_ENABLED"} {
		if !found[k] {
			t.Errorf("missing %s in env", k)
		}
	}
}

func TestCrossCompileString(t *testing.T) {
	opts := CrossCompileOptions{GOOS: "darwin", GOARCH: "arm64", CGOEnabled: "1"}
	s := opts.String()
	if s != "darwin/arm64 (cgo=1)" {
		t.Errorf("String() = %q", s)
	}
}
