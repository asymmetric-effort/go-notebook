// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"strings"
	"testing"
)

func TestCSPScriptHash(t *testing.T) {
	hash := CSPScriptHash("console.log('hello')")
	if !strings.HasPrefix(hash, "'sha256-") {
		t.Errorf("hash = %q, want sha256- prefix", hash)
	}
	if !strings.HasSuffix(hash, "'") {
		t.Error("hash should be single-quoted")
	}
}

func TestCSPScriptHashDeterministic(t *testing.T) {
	h1 := CSPScriptHash("test")
	h2 := CSPScriptHash("test")
	if h1 != h2 {
		t.Error("same script should produce same hash")
	}
}

func TestCSPScriptHashDifferent(t *testing.T) {
	h1 := CSPScriptHash("script1")
	h2 := CSPScriptHash("script2")
	if h1 == h2 {
		t.Error("different scripts should produce different hashes")
	}
}

func TestCSPMetaWithScript(t *testing.T) {
	hash := CSPScriptHash("test()")
	meta := CSPMetaWithScript(hash)
	if !strings.Contains(meta, "Content-Security-Policy") {
		t.Error("missing CSP header")
	}
	if !strings.Contains(meta, "script-src") {
		t.Error("missing script-src directive")
	}
	if !strings.Contains(meta, "sha256-") {
		t.Error("missing hash in script-src")
	}
	if !strings.Contains(meta, "default-src 'none'") {
		t.Error("missing default-src none")
	}
}
