package index

// Separator is the GADDAG rotation separator character.
// It is U+002B '+', which sorts before all uppercase ASCII letters (A–Z),
// satisfying the lexicographic ordering required by the Daciuk construction algorithm.
const Separator rune = '+'

// node is an internal GADDAG graph node.
type node struct {
	children map[rune]*node
	terminal bool
}

func newNode() *node {
	return &node{children: make(map[rune]*node)}
}

// child returns the child node for edge r, or nil if none exists.
func (n *node) child(r rune) *node {
	return n.children[r]
}

// setChild adds or replaces the child node for edge r.
func (n *node) setChild(r rune, child *node) {
	n.children[r] = child
}

// isASCIIUpper reports whether every rune in s is an uppercase ASCII letter (A–Z).
func isASCIIUpper(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// alphabetRunes returns the 26 uppercase ASCII letter runes A–Z.
func alphabetRunes() []rune {
	out := make([]rune, 26)
	for i := range 26 {
		out[i] = rune('A' + i)
	}
	return out
}
