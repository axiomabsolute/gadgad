package query

import (
	"iter"

	"github.com/axiomabsolute/gadgad/index"
)

// Search returns a lazy iterator over all words in idx that satisfy p.
//
// Filtering is applied inline as candidates are yielded from the GADDAG:
//  1. Length constraint
//  2. Word length must equal len(p.Slots) when Slots is non-empty
//  3. Fixed slots must match exactly
//  4. Free slots must not contain letters in p.Excluded
//  5. Free slots must be satisfiable from p.Available (if non-nil), consuming
//     each tile at most as many times as it appears in the rack
//  6. Any slots match any letter and do not consume from p.Available
//
// If the anchor cannot be determined (no Fixed slots and Available is nil or
// empty), Search returns an empty iterator.
//
// Callers that prefer eager results can use NewIterator(Search(...)).Collect().
func Search(idx *index.Index, p Pattern) iter.Seq[string] {
	ar, err := selectAnchor(p)
	if err != nil {
		return func(yield func(string) bool) {}
	}

	return func(yield func(string) bool) {
		var source iter.Seq[string]
		if ar.position < 0 {
			// Anagram / all-Free: anchor letter can be at any position.
			source = idx.Traverse(ar.letter)
		} else {
			source = idx.TraverseAt(ar.letter, ar.position)
		}

		seen := make(map[string]struct{})

		for word := range source {
			runes := []rune(word)
			wlen := len(runes)

			// Length constraint.
			if !p.Length.Matches(wlen) {
				continue
			}
			// Slot count must equal word length when Slots is specified.
			if len(p.Slots) > 0 && wlen != len(p.Slots) {
				continue
			}

			// Per-slot checks: Fixed match, Free exclusion + rack consumption.
			if !matchSlots(runes, p) {
				continue
			}

			// Deduplication (needed when Traverse is used for all-Free patterns).
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

// matchSlots checks all slot constraints for word runes against p.
// It returns false if any constraint is violated.
func matchSlots(runes []rune, p Pattern) bool {
	if len(p.Slots) == 0 {
		return true
	}

	// Copy the rack so we can deplete it per-word without touching p.Available.
	var rack LetterSet
	if p.Available != nil {
		rack = *p.Available
	}

	for i, s := range p.Slots {
		r := runes[i]
		switch {
		case s.isFixed():
			if r != s.fixedLetter() {
				return false
			}
		case s.isFree():
			if p.Excluded != nil && p.Excluded.Contains(r) {
				return false
			}
			if p.Available != nil {
				if rack.Count(r) == 0 {
					return false
				}
				rack[r-'A']--
			}
		default:
			// Any letter accepted; no rack consumption.
		}
	}
	return true
}
