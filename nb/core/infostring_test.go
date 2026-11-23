// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"testing"
	"time"
)

func TestParseInfoStringEmpty(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "" {
		t.Errorf("Name = %q, want empty", attrs.Name)
	}
}

func TestParseInfoStringName(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook name=load", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "load" {
		t.Errorf("Name = %q, want %q", attrs.Name, "load")
	}
}

func TestParseInfoStringExported(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook name=load exported(rows []Row, total float64)", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Exported != "rows []Row, total float64" {
		t.Errorf("Exported = %q", attrs.Exported)
	}
}

func TestParseInfoStringReads(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook reads(sales.csv:embed, regions.csv:runtime)", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(attrs.Reads) != 2 {
		t.Fatalf("Reads = %d, want 2", len(attrs.Reads))
	}
	if attrs.Reads[0].Name != "sales.csv" || attrs.Reads[0].Mode != FileModeEmbed {
		t.Errorf("Reads[0] = %+v", attrs.Reads[0])
	}
	if attrs.Reads[1].Name != "regions.csv" || attrs.Reads[1].Mode != FileModeRuntime {
		t.Errorf("Reads[1] = %+v", attrs.Reads[1])
	}
}

func TestParseInfoStringWrites(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook writes(chart.png, report.html)", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(attrs.Writes) != 2 {
		t.Fatalf("Writes = %d, want 2", len(attrs.Writes))
	}
	if attrs.Writes[0] != "chart.png" {
		t.Errorf("Writes[0] = %q", attrs.Writes[0])
	}
}

func TestParseInfoStringSecrets(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook secrets(db, api_key)", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(attrs.Secrets) != 2 {
		t.Fatalf("Secrets = %d, want 2", len(attrs.Secrets))
	}
}

func TestParseInfoStringDataEmbedded(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook data=embedded", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Data != DataEmbedded {
		t.Errorf("Data = %v, want embedded", attrs.Data)
	}
}

func TestParseInfoStringDataLive(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook data=live", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Data != DataLive {
		t.Errorf("Data = %v, want live", attrs.Data)
	}
}

func TestParseInfoStringTimeout(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook timeout=2m", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want 2m", attrs.Timeout)
	}
}

func TestParseInfoStringFullExample(t *testing.T) {
	info := "go:notebook name=load exported(rows []Row) reads(sales.csv:embed) data=embedded timeout=30s"
	attrs, err := ParseInfoString(info, 5)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "load" {
		t.Errorf("Name = %q", attrs.Name)
	}
	if attrs.Exported != "rows []Row" {
		t.Errorf("Exported = %q", attrs.Exported)
	}
	if len(attrs.Reads) != 1 {
		t.Fatalf("Reads = %d", len(attrs.Reads))
	}
	if attrs.Data != DataEmbedded {
		t.Errorf("Data = %v", attrs.Data)
	}
	if attrs.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v", attrs.Timeout)
	}
}

func TestParseInfoStringErrors(t *testing.T) {
	tests := []struct {
		info    string
		wantErr string
	}{
		{"go:notebook name=", "empty name"},
		{"go:notebook exported(unclosed", "unclosed exported"},
		{"go:notebook reads(file.txt)", "missing :embed or :runtime"},
		{"go:notebook reads(file.txt:bad)", "unknown mode"},
		{"go:notebook reads(unclosed", "unclosed reads"},
		{"go:notebook writes(unclosed", "unclosed writes"},
		{"go:notebook secrets(unclosed", "unclosed secrets"},
		{"go:notebook data=wrong", "not valid"},
		{"go:notebook timeout=xyz", "invalid timeout"},
		{"go:notebook badattr", "unknown attribute"},
	}
	for _, tt := range tests {
		_, err := ParseInfoString(tt.info, 1)
		if err == nil {
			t.Errorf("ParseInfoString(%q) = nil error, want error containing %q", tt.info, tt.wantErr)
			continue
		}
		if !containsSubstring(err.Error(), tt.wantErr) {
			t.Errorf("ParseInfoString(%q) error = %q, want to contain %q", tt.info, err.Error(), tt.wantErr)
		}
	}
}

func TestParseInfoStringErrorLine(t *testing.T) {
	_, err := ParseInfoString("go:notebook name=", 42)
	if err == nil {
		t.Fatal("expected error")
	}
	isErr, ok := err.(*InfoStringError)
	if !ok {
		t.Fatalf("error type = %T, want *InfoStringError", err)
	}
	if isErr.Line != 42 {
		t.Errorf("Line = %d, want 42", isErr.Line)
	}
}

func TestDataModeString(t *testing.T) {
	tests := []struct {
		mode DataMode
		want string
	}{
		{DataNone, ""},
		{DataEmbedded, "embedded"},
		{DataLive, "live"},
		{DataMode(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestNextWord(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello"},
		{"hello", "hello"},
		{"", ""},
		{"  ", ""},
	}
	for _, tt := range tests {
		if got := nextWord(tt.input); got != tt.want {
			t.Errorf("nextWord(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSplitTrimmed(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"a, b, c", 3},
		{"a", 1},
		{"", 0},
		{" , , ", 0},
	}
	for _, tt := range tests {
		got := splitTrimmed(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitTrimmed(%q) = %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestParseInfoStringTrailingWhitespace(t *testing.T) {
	attrs, err := ParseInfoString("go:notebook name=test   ", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "test" {
		t.Errorf("Name = %q, want %q", attrs.Name, "test")
	}
}

func TestParseInfoStringNoPrefix(t *testing.T) {
	// Should handle input without go:notebook prefix.
	attrs, err := ParseInfoString("name=test", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Name != "test" {
		t.Errorf("Name = %q, want %q", attrs.Name, "test")
	}
}

func TestParseInfoStringNestedParens(t *testing.T) {
	// exported() with nested parens like map[string]func(int)
	attrs, err := ParseInfoString("go:notebook exported(m map[string]func(int))", 1)
	if err != nil {
		t.Fatal(err)
	}
	if attrs.Exported != "m map[string]func(int)" {
		t.Errorf("Exported = %q", attrs.Exported)
	}
}

func TestFormatInfoStringEmpty(t *testing.T) {
	got := FormatInfoString(CellAttrs{})
	if got != "go:notebook" {
		t.Errorf("FormatInfoString({}) = %q, want %q", got, "go:notebook")
	}
}

func TestFormatInfoStringName(t *testing.T) {
	got := FormatInfoString(CellAttrs{Name: "load"})
	if got != "go:notebook name=load" {
		t.Errorf("got %q", got)
	}
}

func TestFormatInfoStringFull(t *testing.T) {
	attrs := CellAttrs{
		Name:     "compute",
		Exported: "result float64",
		Reads:    []ReadFile{{Name: "data.csv", Mode: FileModeEmbed}, {Name: "cfg.json", Mode: FileModeRuntime}},
		Writes:   []string{"out.png"},
		Secrets:  []string{"api_key"},
		Data:     DataEmbedded,
		Timeout:  30 * time.Second,
	}
	got := FormatInfoString(attrs)
	want := "go:notebook name=compute exported(result float64) reads(data.csv:embed, cfg.json:runtime) writes(out.png) secrets(api_key) data=embedded timeout=30s"
	if got != want {
		t.Errorf("FormatInfoString() = %q\nwant %q", got, want)
	}
}

func TestFormatInfoStringRoundTrip(t *testing.T) {
	cases := []string{
		"go:notebook",
		"go:notebook name=load",
		"go:notebook name=load exported(rows []Row, total float64)",
		"go:notebook reads(sales.csv:embed, regions.csv:runtime)",
		"go:notebook writes(chart.png, report.html)",
		"go:notebook secrets(db, api_key)",
		"go:notebook data=embedded",
		"go:notebook data=live",
		"go:notebook timeout=2m0s",
		"go:notebook name=load exported(rows []Row) reads(sales.csv:embed) data=embedded timeout=30s",
	}
	for _, tc := range cases {
		attrs, err := ParseInfoString(tc, 1)
		if err != nil {
			t.Errorf("ParseInfoString(%q) error: %v", tc, err)
			continue
		}
		got := FormatInfoString(attrs)
		// Re-parse the formatted output and compare attrs.
		attrs2, err := ParseInfoString(got, 1)
		if err != nil {
			t.Errorf("ParseInfoString(FormatInfoString(%q)) error: %v", tc, err)
			continue
		}
		got2 := FormatInfoString(attrs2)
		if got != got2 {
			t.Errorf("round-trip not idempotent for %q:\n  first:  %q\n  second: %q", tc, got, got2)
		}
	}
}

func TestFormatThenParseRoundTrip(t *testing.T) {
	attrs := CellAttrs{
		Name:     "myCell",
		Exported: "x int, y string",
		Reads:    []ReadFile{{Name: "input.csv", Mode: FileModeRuntime}},
		Writes:   []string{"output.json"},
		Secrets:  []string{"token"},
		Data:     DataLive,
		Timeout:  5 * time.Minute,
	}
	formatted := FormatInfoString(attrs)
	parsed, err := ParseInfoString(formatted, 1)
	if err != nil {
		t.Fatalf("ParseInfoString(%q) error: %v", formatted, err)
	}
	if parsed.Name != attrs.Name {
		t.Errorf("Name: got %q want %q", parsed.Name, attrs.Name)
	}
	if parsed.Exported != attrs.Exported {
		t.Errorf("Exported: got %q want %q", parsed.Exported, attrs.Exported)
	}
	if len(parsed.Reads) != 1 || parsed.Reads[0].Name != "input.csv" || parsed.Reads[0].Mode != FileModeRuntime {
		t.Errorf("Reads: got %+v want %+v", parsed.Reads, attrs.Reads)
	}
	if len(parsed.Writes) != 1 || parsed.Writes[0] != "output.json" {
		t.Errorf("Writes: got %v want %v", parsed.Writes, attrs.Writes)
	}
	if len(parsed.Secrets) != 1 || parsed.Secrets[0] != "token" {
		t.Errorf("Secrets: got %v want %v", parsed.Secrets, attrs.Secrets)
	}
	if parsed.Data != DataLive {
		t.Errorf("Data: got %v want %v", parsed.Data, attrs.Data)
	}
	if parsed.Timeout != 5*time.Minute {
		t.Errorf("Timeout: got %v want %v", parsed.Timeout, attrs.Timeout)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
