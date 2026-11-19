// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// DataMode specifies how a cell's data reaches the report binary.
type DataMode int

const (
	DataNone     DataMode = iota // No data attribute specified.
	DataEmbedded                 // data=embedded: captured at build time.
	DataLive                     // data=live: fetched at render time.
)

// String returns the data mode as a string.
func (d DataMode) String() string {
	switch d {
	case DataNone:
		return ""
	case DataEmbedded:
		return "embedded"
	case DataLive:
		return "live"
	default:
		return "unknown"
	}
}

// FileMode specifies how a reads() file is accessed.
type FileMode int

const (
	FileModeEmbed   FileMode = iota // name:embed
	FileModeRuntime                 // name:runtime
)

// ReadFile is a file declared in reads().
type ReadFile struct {
	Name string
	Mode FileMode
}

// CellAttrs holds the parsed attributes from a cell's info string.
type CellAttrs struct {
	Name     string        // name=<identifier>
	Exported string        // exported(...) raw parameter declarations
	Reads    []ReadFile    // reads(file:mode, ...)
	Writes   []string      // writes(file, ...)
	Secrets  []string      // secrets(name, ...)
	Data     DataMode      // data=embedded or data=live
	Timeout  time.Duration // timeout=<duration>
}

// InfoStringError is a parse error with a Markdown line reference.
type InfoStringError struct {
	Line    int
	Message string
}

func (e *InfoStringError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// ParseInfoString parses a go:notebook info string into structured attributes.
// The info string has the form:
//
//	go:notebook [name=<id>] [exported(...)] [reads(...)] [writes(...)] [secrets(...)] [data=embedded|live] [timeout=<dur>]
//
// line is the 1-based Markdown line number for error reporting.
func ParseInfoString(info string, line int) (CellAttrs, error) {
	var attrs CellAttrs

	// Strip the "go:notebook" prefix.
	rest := info
	if strings.HasPrefix(rest, "go:notebook") {
		rest = strings.TrimSpace(rest[len("go:notebook"):])
	}

	for rest != "" {
		rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
		if rest == "" {
			break
		}

		if strings.HasPrefix(rest, "name=") {
			name, remaining, err := parseNameAttr(rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Name = name
			rest = remaining
		} else if strings.HasPrefix(rest, "exported(") {
			val, remaining, err := parseParenAttr("exported", rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Exported = val
			rest = remaining
		} else if strings.HasPrefix(rest, "reads(") {
			val, remaining, err := parseParenAttr("reads", rest, line)
			if err != nil {
				return attrs, err
			}
			files, err := parseReadsValue(val, line)
			if err != nil {
				return attrs, err
			}
			attrs.Reads = files
			rest = remaining
		} else if strings.HasPrefix(rest, "writes(") {
			val, remaining, err := parseParenAttr("writes", rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Writes = splitTrimmed(val)
			rest = remaining
		} else if strings.HasPrefix(rest, "secrets(") {
			val, remaining, err := parseParenAttr("secrets", rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Secrets = splitTrimmed(val)
			rest = remaining
		} else if strings.HasPrefix(rest, "data=") {
			mode, remaining, err := parseDataAttr(rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Data = mode
			rest = remaining
		} else if strings.HasPrefix(rest, "timeout=") {
			dur, remaining, err := parseTimeoutAttr(rest, line)
			if err != nil {
				return attrs, err
			}
			attrs.Timeout = dur
			rest = remaining
		} else {
			// Unknown attribute.
			word := nextWord(rest)
			return attrs, &InfoStringError{Line: line, Message: fmt.Sprintf("unknown attribute %q", word)}
		}
	}

	return attrs, nil
}

// parseNameAttr parses "name=<identifier>" and returns the name and remaining string.
func parseNameAttr(s string, line int) (string, string, error) {
	s = s[len("name="):]
	name := nextWord(s)
	if name == "" {
		return "", "", &InfoStringError{Line: line, Message: "empty name attribute"}
	}
	return name, strings.TrimSpace(s[len(name):]), nil
}

// parseParenAttr parses "<attr>(<content>)" and returns the content and remaining string.
func parseParenAttr(attr string, s string, line int) (string, string, error) {
	prefix := attr + "("
	s = s[len(prefix):]

	depth := 1
	i := 0
	for i < len(s) && depth > 0 {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
		}
		if depth > 0 {
			i++
		}
	}
	if depth != 0 {
		return "", "", &InfoStringError{Line: line, Message: fmt.Sprintf("unclosed %s(...)", attr)}
	}

	content := s[:i]
	remaining := strings.TrimSpace(s[i+1:])
	return content, remaining, nil
}

// parseReadsValue parses the contents of reads(...) into ReadFile entries.
func parseReadsValue(val string, line int) ([]ReadFile, error) {
	parts := splitTrimmed(val)
	var files []ReadFile
	for _, p := range parts {
		colonIdx := strings.LastIndex(p, ":")
		if colonIdx < 0 {
			return nil, &InfoStringError{Line: line, Message: fmt.Sprintf("reads entry %q missing :embed or :runtime mode", p)}
		}
		name := p[:colonIdx]
		mode := p[colonIdx+1:]
		switch mode {
		case "embed":
			files = append(files, ReadFile{Name: name, Mode: FileModeEmbed})
		case "runtime":
			files = append(files, ReadFile{Name: name, Mode: FileModeRuntime})
		default:
			return nil, &InfoStringError{Line: line, Message: fmt.Sprintf("reads entry %q has unknown mode %q (want embed or runtime)", p, mode)}
		}
	}
	return files, nil
}

// parseDataAttr parses "data=embedded" or "data=live".
func parseDataAttr(s string, line int) (DataMode, string, error) {
	s = s[len("data="):]
	word := nextWord(s)
	switch word {
	case "embedded":
		return DataEmbedded, strings.TrimSpace(s[len(word):]), nil
	case "live":
		return DataLive, strings.TrimSpace(s[len(word):]), nil
	default:
		return DataNone, "", &InfoStringError{Line: line, Message: fmt.Sprintf("data=%q is not valid (want embedded or live)", word)}
	}
}

// parseTimeoutAttr parses "timeout=<duration>".
func parseTimeoutAttr(s string, line int) (time.Duration, string, error) {
	s = s[len("timeout="):]
	word := nextWord(s)
	dur, err := time.ParseDuration(word)
	if err != nil {
		return 0, "", &InfoStringError{Line: line, Message: fmt.Sprintf("invalid timeout %q: %v", word, err)}
	}
	return dur, strings.TrimSpace(s[len(word):]), nil
}

// nextWord returns the next whitespace-delimited word from s.
func nextWord(s string) string {
	i := strings.IndexFunc(s, unicode.IsSpace)
	if i < 0 {
		return s
	}
	return s[:i]
}

// splitTrimmed splits a comma-separated string and trims each part.
func splitTrimmed(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
