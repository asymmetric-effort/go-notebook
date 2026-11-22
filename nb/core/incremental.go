// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

// CellChange describes a changed cell between two versions of a notebook.
type CellChange struct {
	Index    int    // Cell index in the notebook.
	Name     string // Cell name (empty for unnamed cells).
	OldBody  string // Previous body (empty for new cells).
	NewBody  string // Current body (empty for deleted cells).
	IsNew    bool
	IsDelete bool
}

// DiffCells compares two parse results and returns the cells that changed.
// Cells are matched by name for named cells and by position for unnamed cells.
func DiffCells(old, new ParseResult) []CellChange {
	oldCells := extractCellBodies(old)
	newCells := extractCellBodies(new)

	var changes []CellChange

	// Build maps for named cells.
	oldByName := make(map[string]cellBody)
	newByName := make(map[string]cellBody)
	var oldUnnamed, newUnnamed []cellBody

	for _, c := range oldCells {
		if c.name != "" {
			oldByName[c.name] = c
		} else {
			oldUnnamed = append(oldUnnamed, c)
		}
	}
	for _, c := range newCells {
		if c.name != "" {
			newByName[c.name] = c
		} else {
			newUnnamed = append(newUnnamed, c)
		}
	}

	// Check named cells for changes.
	for name, nc := range newByName {
		oc, existed := oldByName[name]
		if !existed {
			changes = append(changes, CellChange{
				Index: nc.index, Name: name, NewBody: nc.body, IsNew: true,
			})
		} else if oc.body != nc.body {
			changes = append(changes, CellChange{
				Index: nc.index, Name: name, OldBody: oc.body, NewBody: nc.body,
			})
		}
	}

	// Check for deleted named cells.
	for name, oc := range oldByName {
		if _, exists := newByName[name]; !exists {
			changes = append(changes, CellChange{
				Index: oc.index, Name: name, OldBody: oc.body, IsDelete: true,
			})
		}
	}

	// Compare unnamed cells by position.
	maxUnnamed := len(oldUnnamed)
	if len(newUnnamed) > maxUnnamed {
		maxUnnamed = len(newUnnamed)
	}
	for i := 0; i < maxUnnamed; i++ {
		if i >= len(oldUnnamed) {
			changes = append(changes, CellChange{
				Index: newUnnamed[i].index, NewBody: newUnnamed[i].body, IsNew: true,
			})
		} else if i >= len(newUnnamed) {
			changes = append(changes, CellChange{
				Index: oldUnnamed[i].index, OldBody: oldUnnamed[i].body, IsDelete: true,
			})
		} else if oldUnnamed[i].body != newUnnamed[i].body {
			changes = append(changes, CellChange{
				Index: newUnnamed[i].index, OldBody: oldUnnamed[i].body, NewBody: newUnnamed[i].body,
			})
		}
	}

	return changes
}

// NeedsFullRegeneration returns true if the changes require a full regeneration
// (e.g., defs changed, cells added/removed, or upstream cells changed).
func NeedsFullRegeneration(old, new ParseResult) bool {
	// If defs blocks changed, full regen needed.
	oldDefs := extractDefs(old)
	newDefs := extractDefs(new)
	if oldDefs != newDefs {
		return true
	}

	changes := DiffCells(old, new)
	for _, c := range changes {
		if c.IsNew || c.IsDelete {
			return true // Structural change.
		}
	}

	return false
}

type cellBody struct {
	name  string
	body  string
	index int
}

func extractCellBodies(result ParseResult) []cellBody {
	var cells []cellBody
	idx := 0
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			attrs, _ := ParseInfoString(b.Info, b.StartPos.Line)
			cells = append(cells, cellBody{
				name:  attrs.Name,
				body:  b.Body,
				index: idx,
			})
			idx++
		}
	}
	return cells
}

func extractDefs(result ParseResult) string {
	var defs []string
	for _, b := range result.Blocks {
		if b.Kind == BlockDefs {
			defs = append(defs, b.Body)
		}
	}
	if len(defs) == 0 {
		return ""
	}
	return joinStrings(defs, "\n\n")
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += sep + s
	}
	return result
}
