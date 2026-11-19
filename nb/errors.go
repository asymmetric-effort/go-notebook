// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package nb

import "fmt"

// CellError wraps an error with the cell name and Markdown line for
// human-readable reporting.
type CellError struct {
	Cell    string
	Line    int
	Wrapped error
}

func (e *CellError) Error() string {
	return fmt.Sprintf("cell %q failed at line %d: %v", e.Cell, e.Line, e.Wrapped)
}

func (e *CellError) Unwrap() error {
	return e.Wrapped
}

// NewCellError creates a CellError.
func NewCellError(cell string, line int, err error) *CellError {
	return &CellError{Cell: cell, Line: line, Wrapped: err}
}

// PanicError converts a recovered panic value into an error.
// Generated code calls this in the deferred recovery function.
func PanicError(r any) error {
	if err, ok := r.(error); ok {
		return fmt.Errorf("panic: %w", err)
	}
	return fmt.Errorf("panic: %v", r)
}
