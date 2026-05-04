package dict

import "strings"

// Source is the interface for loading a word collection into the gadgad index.
type Source interface {
	Words() ([]string, error)
}

// normalize trims whitespace and converts s to uppercase.
func normalize(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
