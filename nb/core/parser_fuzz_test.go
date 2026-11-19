// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

func FuzzParse(f *testing.F) {
	// Seed corpus with representative inputs.
	f.Add("# Hello\n\n```go:defs\nimport \"fmt\"\n```\n")
	f.Add("```go:notebook name=load exported(rows []Row)\nf, err := ctx.Open(\"sales.csv\")\n```\n")
	f.Add("```output name=totals\n- East: 120400.00\n```\n")
	f.Add("```go\nfmt.Println(\"hello\")\n```\n")
	f.Add("````go:notebook name=test\n```\nstill body\n```\n````\n")
	f.Add("   ```go:notebook name=test\n   body\n   ```\n")
	f.Add("~~~go:notebook name=test\nbody\n~~~\n")
	f.Add("```go:notebook name=test\nbody\n")
	f.Add("")
	f.Add("no fences at all")
	f.Add("```\n```\n```\n```\n")
	f.Add("``````````go:notebook\n```go:notebook\n``````````\n")

	f.Fuzz(func(t *testing.T, input string) {
		result := Parse(input)
		// Verify no panic occurred and basic invariants hold.
		for _, b := range result.Blocks {
			if b.StartPos.Line < 1 {
				t.Errorf("block has StartPos.Line < 1: %d", b.StartPos.Line)
			}
			if b.Kind < BlockProse || b.Kind > BlockOutput {
				t.Errorf("block has invalid kind: %d", b.Kind)
			}
		}
	})
}
