// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

func FuzzParseInfoString(f *testing.F) {
	f.Add("go:notebook name=load exported(rows []Row) reads(sales.csv:embed) data=embedded")
	f.Add("go:notebook")
	f.Add("go:notebook name=test timeout=2m")
	f.Add("go:notebook secrets(db, key) writes(out.png)")
	f.Add("")
	f.Add("go:notebook name=")
	f.Add("go:notebook exported(unclosed")
	f.Add("go:notebook data=bad")

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic, errors are fine.
		ParseInfoString(input, 1)
	})
}
