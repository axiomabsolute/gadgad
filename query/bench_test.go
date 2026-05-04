package query_test

import (
	"os"
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
	"github.com/axiomabsolute/gadgad/index"
	"github.com/axiomabsolute/gadgad/query"
)

func buildTWL06(b *testing.B) *index.Index {
	b.Helper()
	path := "../testdata/twl06.txt"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		b.Skip("testdata/twl06.txt not present")
	}
	idx, err := index.Build(dict.File(path))
	if err != nil {
		b.Fatal(err)
	}
	return idx
}

// BenchmarkSearch_Crossword_TWL06 measures a 7-slot crossword pattern with
// three fixed positions and four Any slots.
// Pattern: _ _ A _ _ E D (e.g. CRASHED, CRASHED, etc.)
func BenchmarkSearch_Crossword_TWL06(b *testing.B) {
	idx := buildTWL06(b)
	p := query.Pattern{
		Slots: []query.Slot{
			query.Any, query.Any, query.Fixed('A'), query.Any, query.Any,
			query.Fixed('E'), query.Fixed('D'),
		},
		Anchor: query.Auto,
	}
	b.ResetTimer()
	for range b.N {
		for range query.Search(idx, p) {
		}
	}
}

// BenchmarkSearch_Scrabble_TWL06 measures a Scrabble-style query: two fixed
// positions, five Free slots, with a common 7-tile rack.
// Pattern: _ _ A _ _ E D, rack = PRINTSO
func BenchmarkSearch_Scrabble_TWL06(b *testing.B) {
	idx := buildTWL06(b)
	avail := query.NewLetterSet("PRINTSO")
	p := query.Pattern{
		Slots: []query.Slot{
			query.Free, query.Free, query.Fixed('A'), query.Free, query.Free,
			query.Fixed('E'), query.Fixed('D'),
		},
		Available: &avail,
		Anchor:    query.Auto,
	}
	b.ResetTimer()
	for range b.N {
		for range query.Search(idx, p) {
		}
	}
}

// BenchmarkSearch_Anagram_TWL06_Common measures an anagram query with a rack
// of common letters (AEINTRS). This is a worst-case for the Traverse-based
// anagram path: S is the selected anchor but appears in most TWL06 words.
func BenchmarkSearch_Anagram_TWL06_Common(b *testing.B) {
	idx := buildTWL06(b)
	avail := query.NewLetterSet("AEINTRS")
	p := query.Pattern{
		Slots: []query.Slot{
			query.Free, query.Free, query.Free, query.Free,
			query.Free, query.Free, query.Free,
		},
		Length:    query.Exact(7),
		Available: &avail,
		Anchor:    query.Auto,
	}
	b.ResetTimer()
	for range b.N {
		for range query.Search(idx, p) {
		}
	}
}

// BenchmarkSearch_Anagram_TWL06_Moderate measures an anagram query with a rack
// containing rarer letters (ABORVXZ). The selected anchor (X or Z, freq=1)
// appears in far fewer words, representing a more typical mixed-rack case.
func BenchmarkSearch_Anagram_TWL06_Moderate(b *testing.B) {
	idx := buildTWL06(b)
	avail := query.NewLetterSet("ABORVXZ")
	p := query.Pattern{
		Slots: []query.Slot{
			query.Free, query.Free, query.Free, query.Free,
			query.Free, query.Free, query.Free,
		},
		Length:    query.Exact(7),
		Available: &avail,
		Anchor:    query.Auto,
	}
	b.ResetTimer()
	for range b.N {
		for range query.Search(idx, p) {
		}
	}
}
