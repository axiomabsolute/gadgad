package index

import (
	"fmt"
	"iter"
	"sort"
	"strings"

	"github.com/axiomabsolute/gadgad/dict"
)

// traversalOrder is the fixed DFS edge iteration order for TraverseRaw: Separator then A–Z.
// Allocated once at init; never modified.
var traversalOrder = func() []rune {
	out := make([]rune, 27)
	out[0] = Separator
	for i := range 26 {
		out[1+i] = rune('A' + i)
	}
	return out
}()

// Index is an immutable GADDAG index built from a word list.
// It supports single-anchor traversal to find all words containing a given
// letter at any position.
type Index struct {
	root      *node
	wordCount int
	nodeCount int
	edgeCount int
}

// IndexStats holds aggregate metrics about the index.
type IndexStats struct {
	WordCount int
	NodeCount int
	EdgeCount int
}

// Build constructs a GADDAG index from the words provided by src.
// Words must contain only uppercase ASCII letters (A–Z); other input causes an error.
// Build normalizes words to uppercase (delegating to the Source) and
// deduplicates before construction.
func Build(src dict.Source) (*Index, error) {
	words, err := src.Words()
	if err != nil {
		return nil, err
	}

	// Validate and deduplicate.
	seen := make(map[string]struct{}, len(words))
	unique := make([]string, 0, len(words))
	for _, w := range words {
		if w == "" {
			continue
		}
		if !isASCIIUpper(w) {
			return nil, fmt.Errorf("gadgad: word %q contains characters outside A–Z", w)
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

	// Pre-allocate reusable buffers for insertSuffix (sized for a typical word length).
	pathBuf := make([]*node, 0, 20)
	edgesBuf := make([]rune, 0, 20)

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
		lastPath, lastEdges = insertSuffix(root, rot, lcpLen, pathBuf, edgesBuf)
		// Update buffer headers so any growth from insertSuffix is retained.
		pathBuf = lastPath[:0]
		edgesBuf = lastEdges[:0]
		lastRot = rot
	}

	// Minimize the final path.
	if len(lastPath) > 0 {
		m.replaceOrRegister(lastPath, lastEdges)
	}

	nodes, edges := countNodesEdges(root, make(map[*node]struct{}))
	return &Index{root: root, wordCount: len(words), nodeCount: nodes, edgeCount: edges}, nil
}

// Stats returns aggregate metrics about the index. It is a free struct copy;
// graph metrics are computed once during Build.
func (idx *Index) Stats() IndexStats {
	return IndexStats{
		WordCount: idx.wordCount,
		NodeCount: idx.nodeCount,
		EdgeCount: idx.edgeCount,
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
			for _, r := range traversalOrder {
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
			word := RotationToWord(rot)
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

// TraverseAt returns an iterator over all words in the index where anchor
// appears at exactly the given zero-indexed position. The '+' separator never
// appears in yielded strings. Results contain no duplicates.
//
// This is more efficient than Traverse for position-constrained queries: it
// prunes wrong-position branches during the DFS rather than post-hoc, and
// returns fully decoded words so callers never need to parse rotation strings.
func (idx *Index) TraverseAt(anchor rune, position int) iter.Seq[string] {
	return func(yield func(string) bool) {
		start := idx.root.child(anchor)
		if start == nil || position < 0 {
			return
		}

		// walkSuffix traverses edges after the '+' separator, building suffixBuf.
		// prefixBuf is the reversed prefix (anchor + chars before anchor, reversed)
		// accumulated during walkPrefix; it is read-only here.
		var walkSuffix func(n *node, prefixBuf, suffixBuf []rune) bool
		walkSuffix = func(n *node, prefixBuf, suffixBuf []rune) bool {
			if n.terminal {
				// Reconstruct word: reverse prefixBuf then append suffixBuf.
				wordBuf := make([]rune, len(prefixBuf)+len(suffixBuf))
				for i, r := range prefixBuf {
					wordBuf[len(prefixBuf)-1-i] = r
				}
				copy(wordBuf[len(prefixBuf):], suffixBuf)
				if !yield(string(wordBuf)) {
					return false
				}
			}
			for r := rune('A'); r <= 'Z'; r++ {
				child := n.child(r)
				if child != nil {
					if !walkSuffix(child, prefixBuf, append(suffixBuf, r)) {
						return false
					}
				}
			}
			return true
		}

		// walkPrefix traverses non-separator edges until depth == position,
		// then takes the separator edge and hands off to walkSuffix.
		var walkPrefix func(n *node, depth int, prefixBuf []rune) bool
		walkPrefix = func(n *node, depth int, prefixBuf []rune) bool {
			if depth == position {
				sepChild := n.child(Separator)
				if sepChild == nil {
					return true
				}
				return walkSuffix(sepChild, prefixBuf, nil)
			}
			for r := rune('A'); r <= 'Z'; r++ {
				child := n.child(r)
				if child != nil {
					if !walkPrefix(child, depth+1, append(prefixBuf, r)) {
						return false
					}
				}
			}
			return true
		}

		walkPrefix(start, 0, []rune{anchor})
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

// RotationToWord reconstructs the original word from a GADDAG rotation string.
// The rotation format is: <reversed-prefix-including-anchor>+<suffix>
// Reversing the portion before '+' and concatenating with the portion after
// recovers the original word.
func RotationToWord(rot string) string {
	sepIdx := strings.IndexRune(rot, '+')
	if sepIdx < 0 {
		return rot
	}
	prefix := []rune(rot[:sepIdx])
	suffix := rot[sepIdx+1:]
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
		total += len(w) // all ASCII: byte length equals rune length
	}
	all := make([]string, 0, total)
	for _, w := range words {
		all = append(all, rotations(w)...)
	}
	return all
}

// insertSuffix inserts the portion of rot starting at offset lcpLen into the
// trie rooted at root, following existing nodes for the common prefix portion.
// path and edges are caller-owned buffers; they are re-sliced to [:0] on entry.
// Returns path and edges populated with the full inserted path.
// rot must contain only single-byte ASCII characters.
func insertSuffix(root *node, rot string, lcpLen int, path []*node, edges []rune) ([]*node, []rune) {
	path = path[:0]
	edges = edges[:0]
	cur := root
	path = append(path, cur)
	for i := 0; i < lcpLen; i++ {
		r := rune(rot[i])
		cur = cur.child(r)
		path = append(path, cur)
		edges = append(edges, r)
	}
	for i := lcpLen; i < len(rot); i++ {
		r := rune(rot[i])
		next := newNode()
		cur.setChild(r, next)
		cur = next
		path = append(path, cur)
		edges = append(edges, r)
	}
	cur.terminal = true
	return path, edges
}

// commonPrefixLen returns the length of the longest common prefix of a and b.
// Both strings must contain only single-byte ASCII characters.
func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// isASCIIUpper reports whether every byte in s is an uppercase ASCII letter (A–Z).
func isASCIIUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'Z' {
			return false
		}
	}
	return true
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
