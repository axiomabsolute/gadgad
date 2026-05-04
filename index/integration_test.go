package index_test

import (
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
	"github.com/axiomabsolute/gadgad/index"
)

func TestIntegration_FileSourceWordCount(t *testing.T) {
	src := dict.File("../testdata/small_words.txt")
	words, err := src.Words()
	if err != nil {
		t.Fatalf("Words(): %v", err)
	}

	idx, err := index.Build(src)
	if err != nil {
		t.Fatalf("Build(): %v", err)
	}

	if idx.Stats().WordCount != len(words) {
		t.Errorf("WordCount: expected %d, got %d", len(words), idx.Stats().WordCount)
	}
}

func TestIntegration_KnownWordsViaFileSource(t *testing.T) {
	idx, err := index.Build(dict.File("../testdata/small_words.txt"))
	if err != nil {
		t.Fatalf("Build(): %v", err)
	}

	cases := []struct {
		word   string
		anchor rune
	}{
		{"CARE", 'A'},
		{"CARED", 'R'},
		{"CRANE", 'N'},
		{"STARE", 'T'},
		{"GRACE", 'G'},
	}
	for _, tc := range cases {
		found := false
		for w := range idx.Traverse(tc.anchor) {
			if w == tc.word {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("word %q not found via anchor %c", tc.word, tc.anchor)
		}
	}
}

func TestIntegration_KnownWordsViaSliceSource(t *testing.T) {
	words := []string{"SPARE", "SPARED", "SHARE", "SHARED", "STARE", "SNARE"}
	idx, err := index.Build(dict.Slice(words))
	if err != nil {
		t.Fatalf("Build(): %v", err)
	}

	for _, word := range words {
		for i, anchor := range word {
			found := false
			for w := range idx.Traverse(anchor) {
				if w == word {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("word %q not found via anchor %c (position %d)", word, anchor, i)
			}
		}
	}
}
