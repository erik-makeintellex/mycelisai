package memory

import (
	"fmt"
	"strings"
)

// pqTextArray formats a Go string slice as a Postgres TEXT[] literal.
func pqTextArray(arr []string) string {
	if len(arr) == 0 {
		return "{}"
	}
	escaped := make([]string, len(arr))
	for i, s := range arr {
		escaped[i] = `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(escaped, ",") + "}"
}

// pqScanArray returns a scanner that reads a Postgres TEXT[] into a Go string slice.
func pqScanArray(dest *[]string) interface{ Scan(src any) error } {
	return &pgTextArrayScanner{dest: dest}
}

type pgTextArrayScanner struct {
	dest *[]string
}

func (s *pgTextArrayScanner) Scan(src any) error {
	if src == nil {
		*s.dest = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return s.parseArray(string(v))
	case string:
		return s.parseArray(v)
	default:
		return fmt.Errorf("unsupported pg array type: %T", src)
	}
}

func (s *pgTextArrayScanner) parseArray(raw string) error {
	// Handle empty array
	raw = strings.TrimSpace(raw)
	if raw == "{}" || raw == "" {
		*s.dest = nil
		return nil
	}
	// Strip outer braces and split
	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		if p != "" {
			result = append(result, p)
		}
	}
	*s.dest = result
	return nil
}

// formatVector converts a float64 slice to pgvector string format: "[0.1,0.2,0.3]"
func formatVector(v []float64) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
