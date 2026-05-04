package index_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
	"github.com/axiomabsolute/gadgad/index"
)

// smallWords is a minimal deterministic word set for unit tests.
var smallWords = []string{"CARE", "CARED", "RACE", "RACED", "ACE", "ACRE", "CRANE"}

func buildSmall(t *testing.T) *index.Index {
	t.Helper()
	idx, err := index.Build(dict.Slice(smallWords))
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	return idx
}

// --- Build ---

func TestBuild_NonEmpty(t *testing.T) {
	idx, err := index.Build(dict.Slice(smallWords))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx == nil {
		t.Fatal("expected non-nil index")
	}
}

func TestBuild_Empty(t *testing.T) {
	idx, err := index.Build(dict.Slice(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx == nil {
		t.Fatal("expected non-nil index")
	}
}

func TestBuild_Deduplication(t *testing.T) {
	withDups := append(smallWords, smallWords...)
	idx1, _ := index.Build(dict.Slice(smallWords))
	idx2, _ := index.Build(dict.Slice(withDups))
	if idx1.Stats().WordCount != idx2.Stats().WordCount {
		t.Errorf("dedup: expected same WordCount, got %d vs %d",
			idx1.Stats().WordCount, idx2.Stats().WordCount)
	}
}

func TestBuild_NonASCIIError(t *testing.T) {
	_, err := index.Build(dict.Slice([]string{"CAFÉ"}))
	if err == nil {
		t.Fatal("expected error for non-ASCII word, got nil")
	}
}

func TestBuild_SourceError(t *testing.T) {
	_, err := index.Build(dict.File("/no/such/file.txt"))
	if err == nil {
		t.Fatal("expected error from missing file")
	}
}

// --- Stats ---

func TestStats_WordCount(t *testing.T) {
	idx := buildSmall(t)
	if idx.Stats().WordCount != len(smallWords) {
		t.Errorf("expected WordCount %d, got %d", len(smallWords), idx.Stats().WordCount)
	}
}

// --- TraverseRaw ---

func TestTraverseRaw_SeparatorPresent(t *testing.T) {
	idx := buildSmall(t)
	for rot := range idx.TraverseRaw('C') {
		if !strings.Contains(rot, "+") {
			t.Errorf("TraverseRaw: rotation %q missing '+'", rot)
		}
		if strings.Count(rot, "+") != 1 {
			t.Errorf("TraverseRaw: rotation %q has more than one '+'", rot)
		}
	}
}

func TestTraverseRaw_AnchorIsFirstChar(t *testing.T) {
	idx := buildSmall(t)
	for _, anchor := range []rune{'A', 'C', 'R', 'E'} {
		for rot := range idx.TraverseRaw(anchor) {
			if []rune(rot)[0] != anchor {
				t.Errorf("TraverseRaw(%c): first char of %q is not anchor", anchor, rot)
			}
		}
	}
}

func TestTraverseRaw_UnknownAnchorEmpty(t *testing.T) {
	idx := buildSmall(t)
	count := 0
	for range idx.TraverseRaw('Z') {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 results for unknown anchor Z, got %d", count)
	}
}

// --- Traverse ---

func TestTraverse_NoSeparatorInResults(t *testing.T) {
	idx := buildSmall(t)
	for _, anchor := range []rune{'A', 'C', 'R', 'E', 'N'} {
		for word := range idx.Traverse(anchor) {
			if strings.Contains(word, "+") {
				t.Errorf("Traverse(%c): word %q contains '+'", anchor, word)
			}
		}
	}
}

func TestTraverse_NoDuplicates(t *testing.T) {
	idx := buildSmall(t)
	for _, anchor := range []rune{'A', 'C', 'R', 'E'} {
		seen := make(map[string]struct{})
		for word := range idx.Traverse(anchor) {
			if _, ok := seen[word]; ok {
				t.Errorf("Traverse(%c): duplicate word %q", anchor, word)
			}
			seen[word] = struct{}{}
		}
	}
}

func TestTraverse_UnknownAnchorEmpty(t *testing.T) {
	idx := buildSmall(t)
	count := 0
	for range idx.Traverse('Z') {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 results for unknown anchor, got %d", count)
	}
}

func TestTraverse_SingleLetterWord(t *testing.T) {
	idx, err := index.Build(dict.Slice([]string{"A"}))
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	words := index.Collect(idx.Traverse('A'))
	if len(words) != 1 || words[0] != "A" {
		t.Errorf("expected [A], got %v", words)
	}
}

// TestTraverse_AllWordsReachable checks that every word in the index is
// reachable via every one of its letter positions.
func TestTraverse_AllWordsReachable(t *testing.T) {
	idx := buildSmall(t)
	for _, word := range smallWords {
		for i, anchor := range word {
			found := false
			for w := range idx.Traverse(anchor) {
				if w == word {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("word %q not reachable from anchor %c (position %d)", word, anchor, i)
			}
		}
	}
}

// TestTraverse_ConsistencyWithRaw asserts that reconstructing words from
// TraverseRaw via RotationToWord produces the same set as Traverse for each anchor.
func TestTraverse_ConsistencyWithRaw(t *testing.T) {
	idx := buildSmall(t)
	for _, anchor := range []rune{'A', 'C', 'R', 'E', 'N'} {
		fromRaw := make(map[string]struct{})
		for rot := range idx.TraverseRaw(anchor) {
			fromRaw[index.RotationToWord(rot)] = struct{}{}
		}
		fromTraverse := make(map[string]struct{})
		for word := range idx.Traverse(anchor) {
			fromTraverse[word] = struct{}{}
		}
		if len(fromRaw) != len(fromTraverse) {
			t.Errorf("anchor %c: TraverseRaw reconstructed %d words, Traverse returned %d",
				anchor, len(fromRaw), len(fromTraverse))
			continue
		}
		for word := range fromTraverse {
			if _, ok := fromRaw[word]; !ok {
				t.Errorf("anchor %c: Traverse returned %q but TraverseRaw did not", anchor, word)
			}
		}
	}
}

// --- RotationToWord ---

func TestRotationToWord_SingleLetter(t *testing.T) {
	if got := index.RotationToWord("A+"); got != "A" {
		t.Errorf("expected %q, got %q", "A", got)
	}
}

func TestRotationToWord_AnchorAtFirst(t *testing.T) {
	if got := index.RotationToWord("C+ARED"); got != "CARED" {
		t.Errorf("expected %q, got %q", "CARED", got)
	}
}

func TestRotationToWord_AnchorAtMiddle(t *testing.T) {
	if got := index.RotationToWord("RAC+ED"); got != "CARED" {
		t.Errorf("expected %q, got %q", "CARED", got)
	}
}

func TestRotationToWord_AnchorAtLast(t *testing.T) {
	if got := index.RotationToWord("DERAC+"); got != "CARED" {
		t.Errorf("expected %q, got %q", "CARED", got)
	}
}

func TestRotationToWord_NoSeparator(t *testing.T) {
	if got := index.RotationToWord("CARED"); got != "CARED" {
		t.Errorf("expected %q, got %q", "CARED", got)
	}
}

// --- Collect ---

func TestCollect(t *testing.T) {
	idx := buildSmall(t)
	words := index.Collect(idx.Traverse('C'))
	if len(words) == 0 {
		t.Fatal("expected non-empty Collect result")
	}
	// Verify all returned words contain 'C'.
	for _, w := range words {
		if !strings.ContainsRune(w, 'C') {
			t.Errorf("Collect returned %q which does not contain 'C'", w)
		}
	}
}

// --- Property test ---

func TestProperty_EveryWordReachableFromEveryPosition(t *testing.T) {
	idx, err := index.Build(dict.File("../testdata/small_words.txt"))
	if err != nil {
		t.Fatalf("Build from file: %v", err)
	}
	words, _ := dict.File("../testdata/small_words.txt").Words()

	for _, word := range words {
		for i, anchor := range word {
			results := index.Collect(idx.Traverse(anchor))
			sort.Strings(results)
			pos := sort.SearchStrings(results, word)
			if pos >= len(results) || results[pos] != word {
				t.Errorf("word %q not reachable from anchor %c at position %d", word, anchor, i)
			}
		}
	}
}
