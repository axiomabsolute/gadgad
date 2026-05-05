package query

import "fmt"

// Slot represents a single position in a pattern.
type Slot struct {
	kind   slotKind
	letter rune // only valid when kind == slotFixed
}

type slotKind int

const (
	slotFixed slotKind = iota
	slotFree
	slotAny
)

// Fixed returns a slot that requires exactly letter r.
func Fixed(r rune) Slot { return Slot{kind: slotFixed, letter: r} }

// Free is a slot filled by a letter consumed from the Available rack.
var Free = Slot{kind: slotFree}

// Any is a slot that matches any letter without consuming from Available.
var Any = Slot{kind: slotAny}

func (s Slot) isFixed() bool { return s.kind == slotFixed }
func (s Slot) isFree() bool  { return s.kind == slotFree }

// letter panics if s is not Fixed.
func (s Slot) fixedLetter() rune {
	if s.kind != slotFixed {
		panic("query: fixedLetter called on non-Fixed slot")
	}
	return s.letter
}

// LengthConstraint limits which word lengths are accepted.
type LengthConstraint struct {
	kind lcKind
	n    int
}

type lcKind int

const (
	lcUnconstrained lcKind = iota
	lcExact
	lcMin
	lcMax
)

// Exact requires the word to be exactly n letters long.
func Exact(n int) LengthConstraint { return LengthConstraint{kind: lcExact, n: n} }

// Min requires the word to be at least n letters long.
func Min(n int) LengthConstraint { return LengthConstraint{kind: lcMin, n: n} }

// Max requires the word to be at most n letters long.
func Max(n int) LengthConstraint { return LengthConstraint{kind: lcMax, n: n} }

// Unconstrained accepts words of any length.
var Unconstrained = LengthConstraint{kind: lcUnconstrained}

// Matches reports whether wordLen satisfies the constraint.
func (lc LengthConstraint) Matches(wordLen int) bool {
	switch lc.kind {
	case lcExact:
		return wordLen == lc.n
	case lcMin:
		return wordLen >= lc.n
	case lcMax:
		return wordLen <= lc.n
	default:
		return true
	}
}

// LetterSet is a multiset of uppercase ASCII letters (A–Z), backed by a
// fixed-size array for zero-allocation arithmetic.
type LetterSet [26]int

// Add increments the count of r by delta.
func (ls *LetterSet) Add(r rune, delta int) { ls[r-'A'] += delta }

// Count returns how many times r appears in the set.
func (ls LetterSet) Count(r rune) int { return ls[r-'A'] }

// Contains reports whether r appears at least once.
func (ls LetterSet) Contains(r rune) bool { return ls[r-'A'] > 0 }

// NewLetterSet builds a LetterSet from a string of uppercase ASCII letters.
func NewLetterSet(s string) LetterSet {
	var ls LetterSet
	for _, r := range s {
		ls[r-'A']++
	}
	return ls
}

// AnchorStrategy controls how the anchor letter and position are chosen.
type AnchorStrategy struct {
	kind     anchorKind
	position int
}

type anchorKind int

const (
	anchorAuto anchorKind = iota
	anchorManual
)

// Auto selects the anchor by choosing the rarest fixed letter (lowest
// frequency), minimizing the number of GADDAG candidates before filtering.
var Auto = AnchorStrategy{kind: anchorAuto}

// Manual pins the anchor to the given slot position.  The slot at position
// must be Fixed; Validate() will catch violations.
func Manual(position int) AnchorStrategy {
	return AnchorStrategy{kind: anchorManual, position: position}
}

// Pattern describes a structured word query.
//
//   - Slots defines per-position constraints; len(Slots) is the required word
//     length when it is non-zero.
//   - Length is an additional or alternative length constraint.
//   - Available is the multiset of tiles available to fill Free slots; nil
//     means no rack constraint.
//   - Excluded lists letters that must NOT fill Free slots; nil means no
//     exclusion constraint.
//   - Anchor controls which Fixed slot becomes the GADDAG traversal root.
type Pattern struct {
	Slots     []Slot
	Length    LengthConstraint
	Available *LetterSet
	Excluded  *LetterSet
	Anchor    AnchorStrategy
}

// Validate checks the pattern for structural correctness.  It returns an error
// if a Manual anchor points at a Free or Any slot, or if the position is out
// of range.
func (p Pattern) Validate() error {
	if p.Anchor.kind != anchorManual {
		return nil
	}
	pos := p.Anchor.position
	if pos < 0 || pos >= len(p.Slots) {
		return fmt.Errorf("query: Manual anchor position %d is out of range [0, %d)", pos, len(p.Slots))
	}
	if !p.Slots[pos].isFixed() {
		return fmt.Errorf("query: Manual anchor at position %d must be a Fixed slot", pos)
	}
	return nil
}
