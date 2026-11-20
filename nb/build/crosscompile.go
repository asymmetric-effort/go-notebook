// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// CrossCompileOptions holds the target platform for go build.
type CrossCompileOptions struct {
	GOOS       string
	GOARCH     string
	CGOEnabled string
}

// DefaultCrossCompile returns options for the current platform.
func DefaultCrossCompile() CrossCompileOptions {
	return CrossCompileOptions{
		GOOS:       runtime.GOOS,
		GOARCH:     runtime.GOARCH,
		CGOEnabled: "1",
	}
}

// SetEnv applies the cross-compilation environment to a command.
func (o CrossCompileOptions) SetEnv(cmd *exec.Cmd) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", o.GOOS),
		fmt.Sprintf("GOARCH=%s", o.GOARCH),
		fmt.Sprintf("CGO_ENABLED=%s", o.CGOEnabled),
	)
}

// String returns a human-readable description.
func (o CrossCompileOptions) String() string {
	return fmt.Sprintf("%s/%s (cgo=%s)", o.GOOS, o.GOARCH, o.CGOEnabled)
}
