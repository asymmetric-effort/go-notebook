// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// CSPScriptHash computes the SHA-256 hash of a script for use in a
// Content-Security-Policy script-src directive. The hash allows exactly
// one inline script to run while blocking all others.
func CSPScriptHash(script string) string {
	hash := sha256.Sum256([]byte(script))
	encoded := base64.StdEncoding.EncodeToString(hash[:])
	return fmt.Sprintf("'sha256-%s'", encoded)
}

// CSPMetaWithScript returns a CSP meta tag that allows inline styles,
// data: images, and exactly the given script hash. All other scripts
// are blocked.
func CSPMetaWithScript(scriptHash string) string {
	return fmt.Sprintf(
		`<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:; script-src %s;">`,
		scriptHash,
	)
}
