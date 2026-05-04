package query

import "fmt"

// letterFreq is the Scrabble tile count for each letter (A–Z, index = letter-'A').
// Lower count means rarer; we pick the rarest Fixed letter as anchor to
// minimize GADDAG fan-out.
var letterFreq = [26]int{
	// A   B  C  D   E  F  G  H  I  J  K  L  M
	9, 2, 2, 4, 12, 2, 3, 2, 9, 1, 1, 4, 2,
	// N  O  P  Q  R  S   T  U  V  W  X  Y  Z
	6, 8, 2, 1, 6, 4, 6, 4, 2, 2, 1, 2, 1,
}

// anchorResult holds the chosen anchor letter and its word-position.
// position == -1 means "any position" — the caller should use Traverse
// rather than TraverseAt (used for all-Free/anagram patterns).
type anchorResult struct {
	letter   rune
	position int
}

// selectAnchor returns the best anchor for p.
//
// For Auto strategy:
//   - Among Fixed slots, pick the one with the lowest letterFreq value.
//   - If no Fixed slots exist (anagram/all-Free), pick the rarest letter in
//     Available and return position == -1 (position-unconstrained).
//
// Tiebreaks are resolved alphabetically (earlier letter wins) for determinism.
//
// For Manual strategy: use the given position directly.
func selectAnchor(p Pattern) (anchorResult, error) {
	if p.Anchor.kind == anchorManual {
		pos := p.Anchor.position
		return anchorResult{letter: p.Slots[pos].fixedLetter(), position: pos}, nil
	}

	// Auto: scan Fixed slots for the rarest letter.
	bestPos := -1
	bestFreq := int(^uint(0) >> 1) // maxInt
	bestLetter := rune(0)
	for i, s := range p.Slots {
		if !s.isFixed() {
			continue
		}
		r := s.fixedLetter()
		freq := letterFreq[r-'A']
		if freq < bestFreq || (freq == bestFreq && r < bestLetter) {
			bestFreq = freq
			bestPos = i
			bestLetter = r
		}
	}
	if bestPos >= 0 {
		return anchorResult{letter: bestLetter, position: bestPos}, nil
	}

	// No Fixed slots — anagram case. Pick rarest letter in Available.
	if p.Available == nil {
		return anchorResult{}, fmt.Errorf("query: cannot select anchor: pattern has no Fixed slots and Available is nil")
	}
	bestIdx := -1
	bestFreq = int(^uint(0) >> 1)
	for i := 0; i < 26; i++ {
		if (*p.Available)[i] == 0 {
			continue
		}
		freq := letterFreq[i]
		if freq < bestFreq || (freq == bestFreq && (bestIdx == -1 || i < bestIdx)) {
			bestFreq = freq
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return anchorResult{}, fmt.Errorf("query: cannot select anchor: pattern has no Fixed slots and Available is empty")
	}
	return anchorResult{letter: rune('A' + bestIdx), position: -1}, nil
}
