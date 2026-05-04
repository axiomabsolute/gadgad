package index

import (
	"fmt"
	"iter"
	"sort"
	"strings"

	"github.com/axiomabsolute/gadgad/dict"
)

// Index is an immutable GADDAG index built from a word list.
// It supports single-anchor traversal to find all words containing a given
// letter at any position.
type Index struct {
	root      *node
	wordCount int
}

// IndexStats holds aggregate metrics about the index.
type IndexStats struct {
	WordCount int
	NodeCount int
	EdgeCount int
}

// Build constructs a GADDAG index from the words provided by src.
// Words must be ASCII-encodeable; non-ASCII words cause an error.
// Build normalizes words to uppercase (delegating to the Source) and
// deduplicates before construction.
func Build(src dict.Source) (*Index, error) {
	words, err := src.Words()
	if err != nil {
		return nil, err
	}

	// Validate and deduplicate.
	seen := make(map[string]struct{}, len(words))
	unique := words[:0]
	for _, w := range words {
		if w == "" {
			continue
		}
		if !isASCIIUpper(w) {
			return nil, fmt.Errorf("gadgad: word %q contains non-ASCII or non-uppercase runes", w)
		}
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			unique = append(unique, w)
		}
	}
	words = unique

	// Generate and sort all rotation strings.
	rotList := buildRotationList(words)
	sort.Strings(rotList)

	// Build the minimized DAWG using Daciuk's algorithm.
	root := newNode()
	m := newMinimizer()
	var lastRot string
	var lastPath []*node
	var lastEdges []rune

	for _, rot := range rotList {
		// Find the longest common prefix with the previous rotation.
		lcpLen := commonPrefixLen(lastRot, rot)

		// Minimize the suffix of the last path that diverges from this rotation.
		// Only nodes at positions lcpLen and beyond are not shared with the
		// current rotation, so we start replaceOrRegister from lastPath[lcpLen].
		if lcpLen < len(lastPath)-1 {
			m.replaceOrRegister(lastPath[lcpLen:], lastEdges[lcpLen:])
		}

		// Insert the new suffix into the trie starting from the common prefix node.
		path, edges := insertSuffix(root, rot, lcpLen)
		lastRot = rot
		lastPath = path
		lastEdges = edges
	}

	// Minimize the final path.
	if len(lastPath) > 0 {
		m.replaceOrRegister(lastPath, lastEdges)
	}

	idx := &Index{root: root, wordCount: len(words)}
	return idx, nil
}

// Stats returns aggregate metrics about the index.
func (idx *Index) Stats() IndexStats {
	nodes, edges := countNodesEdges(idx.root, make(map[*node]struct{}))
	return IndexStats{
		WordCount: idx.wordCount,
		NodeCount: nodes,
		EdgeCount: edges,
	}
}

// TraverseRaw returns an iterator over raw GADDAG rotation strings for all
// rotations whose anchor letter (first character) equals anchor.
// Each yielded string contains exactly one '+' separator.
// The anchor letter is always the first character of each yielded string.
func (idx *Index) TraverseRaw(anchor rune) iter.Seq[string] {
	return func(yield func(string) bool) {
		start := idx.root.child(anchor)
		if start == nil {
			return
		}
		var walk func(n *node, buf []rune) bool
		walk = func(n *node, buf []rune) bool {
			if n.terminal {
				if !yield(string(buf)) {
					return false
				}
			}
			for _, r := range append([]rune{Separator}, alphabetRunes()...) {
				child := n.child(r)
				if child != nil {
					if !walk(child, append(buf, r)) {
						return false
					}
				}
			}
			return true
		}
		walk(start, []rune{anchor})
	}
}

// Traverse returns an iterator over all words in the index that contain
// anchor at any letter position. The '+' separator never appears in yielded
// strings. Results may arrive in any order but contain no duplicates.
func (idx *Index) Traverse(anchor rune) iter.Seq[string] {
	return func(yield func(string) bool) {
		seen := make(map[string]struct{})
		for rot := range idx.TraverseRaw(anchor) {
			word := rotationToWord(rot)
			if _, ok := seen[word]; ok {
				continue
			}
			seen[word] = struct{}{}
			if !yield(word) {
				return
			}
		}
	}
}

// Collect materializes all values from seq into a slice.
func Collect(seq iter.Seq[string]) []string {
	var out []string
	for s := range seq {
		out = append(out, s)
	}
	return out
}

// rotationToWord reconstructs the original word from a GADDAG rotation string.
// The rotation format is: <anchor><reversed-prefix>+<suffix>
// The original word is: reverse(<anchor><reversed-prefix>) + <suffix>
// = <prefix><anchor> ... reversed ... + suffix
// Concretely: join(reverse(part before '+'), part after '+')
func rotationToWord(rot string) string {
	sepIdx := strings.IndexRune(rot, '+')
	if sepIdx < 0 {
		return rot
	}
	prefix := []rune(rot[:sepIdx])
	suffix := rot[sepIdx+1:]
	// Reverse the prefix to recover the original word order up to the anchor.
	for i, j := 0, len(prefix)-1; i < j; i, j = i+1, j-1 {
		prefix[i], prefix[j] = prefix[j], prefix[i]
	}
	return string(prefix) + suffix
}

// buildRotationList generates all GADDAG rotations for every word and returns
// the combined unsorted list.
func buildRotationList(words []string) []string {
	total := 0
	for _, w := range words {
		total += len([]rune(w))
	}
	all := make([]string, 0, total)
	for _, w := range words {
		all = append(all, rotations(w)...)
	}
	return all
}

// insertSuffix inserts the portion of rot starting at offset lcpLen into the
// trie rooted at root, following existing nodes for the common prefix portion.
// Returns the full path from root to the new terminal node, and the edges
// (rune labels) connecting each consecutive pair.
func insertSuffix(root *node, rot string, lcpLen int) ([]*node, []rune) {
	runes := []rune(rot)
	// Walk the existing common prefix nodes.
	path := make([]*node, 0, len(runes)+1)
	edges := make([]rune, 0, len(runes))
	cur := root
	path = append(path, cur)
	for i := 0; i < lcpLen; i++ {
		cur = cur.child(runes[i])
		path = append(path, cur)
		edges = append(edges, runes[i])
	}
	// Insert new nodes for the novel suffix.
	for i := lcpLen; i < len(runes); i++ {
		next := newNode()
		cur.setChild(runes[i], next)
		cur = next
		path = append(path, cur)
		edges = append(edges, runes[i])
	}
	cur.terminal = true
	return path, edges
}

// commonPrefixLen returns the length of the longest common prefix of a and b.
func commonPrefixLen(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := len(ra)
	if len(rb) < n {
		n = len(rb)
	}
	for i := 0; i < n; i++ {
		if ra[i] != rb[i] {
			return i
		}
	}
	return n
}

// countNodesEdges counts unique nodes and edges in the graph via DFS.
func countNodesEdges(n *node, visited map[*node]struct{}) (nodes, edges int) {
	if _, ok := visited[n]; ok {
		return 0, 0
	}
	visited[n] = struct{}{}
	nodes = 1
	for _, child := range n.children {
		edges++
		cn, ce := countNodesEdges(child, visited)
		nodes += cn
		edges += ce
	}
	return nodes, edges
}
