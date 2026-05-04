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
