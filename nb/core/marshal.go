// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"strings"
)

// GenerateMarshal generates encode/decode methods for a named cell struct.
// These methods are used for deep copy (serialize-and-deserialize) and for
// passing values between processes in notebook mode.
func GenerateMarshal(structName string, fields []string) string {
	var buf strings.Builder

	// MarshalBinary method.
	fmt.Fprintf(&buf, "func (c %s) MarshalBinary() ([]byte, error) {\n", structName)
	fmt.Fprintf(&buf, "\tvar buf bytes.Buffer\n")
	fmt.Fprintf(&buf, "\tenc := gob.NewEncoder(&buf)\n")
	for _, f := range fields {
		name := fieldName(f)
		fmt.Fprintf(&buf, "\tif err := enc.Encode(c.%s); err != nil {\n", name)
		fmt.Fprintf(&buf, "\t\treturn nil, fmt.Errorf(\"marshal %s.%s: %%w\", err)\n", structName, name)
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\treturn buf.Bytes(), nil\n")
	fmt.Fprintf(&buf, "}\n\n")

	// UnmarshalBinary method.
	fmt.Fprintf(&buf, "func (c *%s) UnmarshalBinary(data []byte) error {\n", structName)
	fmt.Fprintf(&buf, "\tdec := gob.NewDecoder(bytes.NewReader(data))\n")
	for _, f := range fields {
		name := fieldName(f)
		fmt.Fprintf(&buf, "\tif err := dec.Decode(&c.%s); err != nil {\n", name)
		fmt.Fprintf(&buf, "\t\treturn fmt.Errorf(\"unmarshal %s.%s: %%w\", err)\n", structName, name)
		fmt.Fprintf(&buf, "\t}\n")
	}
	fmt.Fprintf(&buf, "\treturn nil\n")
	fmt.Fprintf(&buf, "}\n")

	return buf.String()
}

// fieldName extracts the field name from a declaration like "rows []Row".
func fieldName(field string) string {
	parts := strings.Fields(field)
	if len(parts) >= 1 {
		return parts[0]
	}
	return field
}

// DeepCopyCode generates a DeepCopy function that serializes and
// deserializes a cell struct to create an independent copy.
func DeepCopyCode(structName string) string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "func deepCopy%s(src %s) %s {\n", capitalize(structName), structName, structName)
	fmt.Fprintf(&buf, "\tdata, err := src.MarshalBinary()\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tpanic(fmt.Sprintf(\"deep copy %s: %%v\", err))\n", structName)
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tvar dst %s\n", structName)
	fmt.Fprintf(&buf, "\tif err := dst.UnmarshalBinary(data); err != nil {\n")
	fmt.Fprintf(&buf, "\t\tpanic(fmt.Sprintf(\"deep copy %s: %%v\", err))\n", structName)
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\treturn dst\n")
	fmt.Fprintf(&buf, "}\n")
	return buf.String()
}
