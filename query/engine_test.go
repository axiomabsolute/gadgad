package query_test

import (
	"slices"
	"testing"

	"github.com/axiomabsolute/gadgad/dict"
	"github.com/axiomabsolute/gadgad/index"
	"github.com/axiomabsolute/gadgad/query"
)

// smallWords: CARE, CARED, RACE, RACED, ACE, ACRE, CRANE
func buildSmall(t *testing.T) *index.Index {
	t.Helper()
	idx, err := index.Build(dict.Slice([]string{
		"CARE", "CARED", "RACE", "RACED", "ACE", "ACRE", "CRANE",
	}))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return idx
}

func collect(seq func(yield func(string) bool)) []string {
	return query.NewIterator(seq).Collect()
}

// --- Mode: Crossword ---

func TestSearch_Crossword_FixedAndAny(t *testing.T) {
	idx := buildSmall(t)
	// Pattern: Any Any Fixed('A') Any Fixed('E') — 5-letter words with A at pos 2 and E at pos 4.
	// CRANE: C=0 R=1 A=2 N=3 E=4 ✓
	p := query.Pattern{
		Slots:  []query.Slot{query.Any, query.Any, query.Fixed('A'), query.Any, query.Fixed('E')},
		Anchor: query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "CRANE") {
		t.Errorf("crossword: expected CRANE in results, got %v", results)
	}
	for _, w := range results {
		runes := []rune(w)
		if len(runes) != 5 || runes[2] != 'A' || runes[4] != 'E' {
			t.Errorf("crossword: %q does not match pattern __A_E", w)
		}
	}
}

func TestSearch_Crossword_NoRackConstraint(t *testing.T) {
	idx := buildSmall(t)
	// Fixed('C') at pos 0, Any elsewhere, length 4.
	// Should include CARE, CARED filtered to length 4.
	p := query.Pattern{
		Slots:  []query.Slot{query.Fixed('C'), query.Any, query.Any, query.Fixed('E')},
		Anchor: query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "CARE") {
		t.Errorf("crossword: expected CARE in results %v", results)
	}
	for _, w := range results {
		runes := []rune(w)
		if runes[0] != 'C' || runes[len(runes)-1] != 'E' {
			t.Errorf("crossword: %q does not match C___E pattern", w)
		}
	}
}

// --- Mode: Scrabble ---

func TestSearch_Scrabble_RackConstraint(t *testing.T) {
	idx := buildSmall(t)
	// Pattern: Fixed('A') at pos 0, Free Free — 3-letter words starting with A
	// using tiles from rack {C, E}.
	// ACE: needs C and E ✓
	// ACRE: length 4, excluded by slot count.
	avail := query.NewLetterSet("CE")
	p := query.Pattern{
		Slots:     []query.Slot{query.Fixed('A'), query.Free, query.Free},
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "ACE") {
		t.Errorf("scrabble: expected ACE in results %v", results)
	}
	for _, w := range results {
		if len([]rune(w)) != 3 {
			t.Errorf("scrabble: %q has wrong length", w)
		}
	}
}

func TestSearch_Scrabble_RackDepletion(t *testing.T) {
	idx := buildSmall(t)
	// Rack has one R. Pattern: Free Fixed('A') Fixed('C') Fixed('E') — 4 letters, _ACE.
	// RACE: R A C E — needs R(1) from rack ✓
	// Nothing else in smallWords matches _ACE.
	avail := query.NewLetterSet("R")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Fixed('A'), query.Fixed('C'), query.Fixed('E')},
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "RACE") {
		t.Errorf("scrabble rack: expected RACE in results %v", results)
	}
}

func TestSearch_Scrabble_RackExhausted(t *testing.T) {
	idx := buildSmall(t)
	// Rack has no R. Pattern needs R in a Free slot — RACE should not appear.
	avail := query.NewLetterSet("XYZ")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Fixed('A'), query.Fixed('C'), query.Fixed('E')},
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if slices.Contains(results, "RACE") {
		t.Errorf("scrabble rack: RACE should not appear when R is not in rack")
	}
}

// --- Mode: Wordle ---

func TestSearch_Wordle_FixedAndExclusion(t *testing.T) {
	idx := buildSmall(t)
	// Five-letter word ending in E, with known letters {C,R,A,N} and gray letter D.
	// CRANE: C R A N E — contains all known letters, does not contain D ✓
	avail := query.NewLetterSet("CRAN")
	excluded := query.NewLetterSet("D")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Free, query.Free, query.Free, query.Fixed('E')},
		Available: &avail,
		Excluded:  &excluded,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "CRANE") {
		t.Errorf("wordle: expected CRANE in results %v", results)
	}
	for _, w := range results {
		runes := []rune(w)
		if runes[len(runes)-1] != 'E' {
			t.Errorf("wordle: %q does not end in E", w)
		}
	}
}

func TestSearch_Wordle_GrayLetterExcluded(t *testing.T) {
	idx := buildSmall(t)
	// Gray letters: C, R — CRANE should not appear (contains C and R in Free slots).
	avail := query.NewLetterSet("ANED")
	excluded := query.NewLetterSet("CR")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Free, query.Free, query.Free, query.Fixed('E')},
		Available: &avail,
		Excluded:  &excluded,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if slices.Contains(results, "CRANE") {
		t.Errorf("wordle: CRANE should not appear when C and R are excluded")
	}
}

// --- Mode: Anagram ---

func TestSearch_Anagram_ExactLength(t *testing.T) {
	idx := buildSmall(t)
	// All Free, length 3, available = {A, C, E}.
	// ACE: A C E — uses all three tiles ✓
	avail := query.NewLetterSet("ACE")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Free, query.Free},
		Length:    query.Exact(3),
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	if !slices.Contains(results, "ACE") {
		t.Errorf("anagram: expected ACE in results %v", results)
	}
	for _, w := range results {
		if len([]rune(w)) != 3 {
			t.Errorf("anagram: %q has wrong length (want 3)", w)
		}
	}
}

func TestSearch_Anagram_NoExtraResults(t *testing.T) {
	idx := buildSmall(t)
	// Available = {A, C, E} — CARE (needs R) and CRANE (needs R,N) must not appear.
	avail := query.NewLetterSet("ACE")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Free, query.Free},
		Length:    query.Exact(3),
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	for _, w := range results {
		if w == "CARE" || w == "CRANE" {
			t.Errorf("anagram: %q must not appear (requires letters not in rack)", w)
		}
	}
}

func TestSearch_Anagram_NoDuplicates(t *testing.T) {
	idx := buildSmall(t)
	avail := query.NewLetterSet("CARNED")
	p := query.Pattern{
		Slots:     []query.Slot{query.Free, query.Free, query.Free, query.Free, query.Free},
		Length:    query.Exact(5),
		Available: &avail,
		Anchor:    query.Auto,
	}
	results := collect(query.Search(idx, p))
	seen := make(map[string]struct{})
	for _, w := range results {
		if _, ok := seen[w]; ok {
			t.Errorf("anagram: duplicate word %q", w)
		}
		seen[w] = struct{}{}
	}
}

// --- No results ---

func TestSearch_NoResults_UnknownLetter(t *testing.T) {
	idx := buildSmall(t)
	p := query.Pattern{
		Slots:  []query.Slot{query.Fixed('Q'), query.Free, query.Free},
		Anchor: query.Auto,
	}
	results := collect(query.Search(idx, p))
	if len(results) != 0 {
		t.Errorf("expected empty results for pattern with Q, got %v", results)
	}
}

func TestSearch_NoResults_EmptyIteratorNoError(t *testing.T) {
	idx := buildSmall(t)
	// Impossible rack: requires Z but available is empty.
	avail := query.NewLetterSet("XYZ")
	p := query.Pattern{
		Slots:     []query.Slot{query.Fixed('A'), query.Free, query.Free, query.Free},
		Available: &avail,
		Anchor:    query.Auto,
	}
	var count int
	for range query.Search(idx, p) {
		count++
	}
	// Just verifying no panic — empty result is fine.
	_ = count
}

// --- Validate ---

func TestPattern_Validate_ManualOnFreeReturnsError(t *testing.T) {
	p := query.Pattern{
		Slots:  []query.Slot{query.Free, query.Fixed('A')},
		Anchor: query.Manual(0),
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for Manual anchor on Free slot, got nil")
	}
}

func TestPattern_Validate_ManualOnFixedOK(t *testing.T) {
	p := query.Pattern{
		Slots:  []query.Slot{query.Free, query.Fixed('A')},
		Anchor: query.Manual(1),
	}
	if err := p.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPattern_Validate_OutOfRange(t *testing.T) {
	p := query.Pattern{
		Slots:  []query.Slot{query.Fixed('A')},
		Anchor: query.Manual(5),
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for out-of-range Manual position, got nil")
	}
}
