package index

import "fmt"

// minimize implements Daciuk's algorithm for incremental DAWG minimization.
//
// The algorithm works on a sorted sequence of rotation strings. After inserting
// each string, it walks back up the last-inserted path and collapses any
// suffixes that are structurally equivalent (same children, same terminal flag)
// into a single shared node.
//
// This file manages the "register" of known minimized nodes and the
// replaceOrRegister procedure used after each insertion.

// nodeKey produces a canonical key for a node based on its structure.
// Two nodes with the same key are equivalent and can be merged.
func nodeKey(n *node) string {
	key := make([]byte, 0, 16)
	if n.terminal {
		key = append(key, '1')
	} else {
		key = append(key, '0')
	}
	// Encode children in sorted edge order for a stable key.
	// We use the pointer address of each child as its identity after
	// minimization (children are already minimized when we process their parent).
	for r := 'A'; r <= 'Z'; r++ {
		if child := n.children[r]; child != nil {
			key = append(key, []byte(fmt.Sprintf("%c:%p;", r, child))...)
		}
	}
	if child := n.children[Separator]; child != nil {
		key = append(key, []byte(fmt.Sprintf("+:%p;", child))...)
	}
	return string(key)
}

// minimizer holds the register of already-minimized nodes used by Daciuk's algorithm.
type minimizer struct {
	register map[string]*node
}

func newMinimizer() *minimizer {
	return &minimizer{register: make(map[string]*node)}
}

// replaceOrRegister walks the suffix of the last-inserted path from the
// deepest node upward, replacing each node with its registered equivalent
// (or registering it if none exists yet).
//
// path is the sequence of nodes from root to the last-inserted terminal,
// and edges is the corresponding sequence of rune labels on each arc.
func (m *minimizer) replaceOrRegister(path []*node, edges []rune) {
	// Walk from deepest node toward root.
	for i := len(path) - 1; i >= 1; i-- {
		child := path[i]
		parentEdge := edges[i-1]
		parent := path[i-1]

		key := nodeKey(child)
		if existing, ok := m.register[key]; ok {
			// Replace child with the already-registered equivalent.
			parent.setChild(parentEdge, existing)
		} else {
			m.register[key] = child
		}
	}
}
