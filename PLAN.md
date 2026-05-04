# gadgad — Project Plan

## Overview

`gadgad` is a Go library for efficient word search in text-based language games
such as Scrabble, Wordle, and crossword puzzles. Its inaugural feature is an
optimized index that, given a dictionary, answers queries of the form:

> "What words can be formed when I know some letters, some positions, and have
> a set of available tiles?"

The library is designed to be both a **pragmatic tool** for game developers and
AI agents, and a **learning resource** for students interested in data
structures, graph algorithms, and applied combinatorics.

Module: `github.com/axiomabsolute/gadgad`

---

## Design Philosophy

- **Layered API** — Users should be able to engage at whatever depth they need.
  A game developer can call a single top-level function; a student or researcher
  can import the index layer directly and work with traversal primitives.
- **Clean public surface** — Every exported type and function should be
  documented and intentional. Nothing is exported by accident.
- **Algorithm transparency** — The GADDAG is not widely understood outside
  competitive Scrabble circles. This library should demystify it: clear naming,
  inline references to the source paper, and a thorough README explainer.
- **Static dictionary** — The index is built once and queried many times.
  Dynamic updates are explicitly out of scope for v1 (see Future Considerations).

---

## Architecture

The library is organized into three layers:

```
┌────────────────────────────────────────────────────────┐
│  Layer 3: Game adapters (future)                       │
│  ScrabbleBoard, CrosswordGrid, WordleGame, etc.        │
│  Translates game-specific state into Layer 2 queries   │
├────────────────────────────────────────────────────────┤
│  Layer 2: Query engine          (query/)               │
│  Pattern type: fixed positions, available letter set,  │
│  exclusions, length constraints, anchor selection      │
├────────────────────────────────────────────────────────┤
│  Layer 1: GADDAG index          (index/)               │
│  Graph construction, minimization, traversal           │
│  primitives. No game semantics.                        │
└────────────────────────────────────────────────────────┘
        ▲                   ▲
        │                   │
   dict.Source         gadgad.go
   (dict/)             top-level convenience wrapper
```

### Layer 1 — GADDAG Index (`index/`)

The core data structure is a **GADDAG** (Graph Automaton for Dictionary
Directed by Gordon), introduced by Steven Gordon (1994). It is a variant of a
DAWG (Directed Acyclic Word Graph) with a specialized node encoding that
supports efficient bidirectional traversal from any anchor letter in a word.

**Key insight:** a standard trie handles prefix queries efficiently, and a
reverse trie handles suffix queries, but neither handles the general case of
"find all words containing letter X at position N with letters Y, Z, ... on
either side." GADDAG encodes this by storing, for every word, all rotations
around each letter position, separated by a reserved `+` character.

For the word `CARED`:

```
Rotate around position 0:  C + A R E D
Rotate around position 1:  A C + R E D
Rotate around position 2:  R A C + E D
Rotate around position 3:  E R A C + D
Rotate around position 4:  D E R A C +
```

All rotations are inserted into a single minimized directed acyclic graph. To
find all words containing `A` at position 2 (0-indexed), traverse from node
`A`, travel left (before the `+`) to fill positions 0–1, cross the separator,
then travel right to fill positions 3–N.

**Construction algorithm:** Daciuk et al. (2000) for online DAWG construction,
applied after generating all rotations. This is the most algorithmically
complex part of the library and should be implemented and verified in isolation
before any layers are built on top.

**Layer 1 exposes:**

- `index.Build(words []string) (*Index, error)` — construct the GADDAG
- `index.Index.Traverse(anchor rune, backward, forward TraversalConstraint) Iterator`
  — primitive traversal from a single anchor letter
- `index.Index.Size() IndexStats` — node count, edge count, word count

### Layer 2 — Query Engine (`query/`)

Translates structured queries into one or more Layer 1 traversals, applies
constraints during traversal, and returns results.

**Pattern type:**

```
Pattern
  ├── Slots []Slot         each position: Fixed(rune) | Free | Any
  ├── Length  LengthConstraint (exact, min, max, or unconstrained)
  ├── Available LetterSet  tiles available to fill Free slots (nil = no constraint)
  ├── Excluded  LetterSet  letters that must NOT appear in Free slots
  └── Anchor    AnchorStrategy (Auto | Manual(position))
```

Auto anchor selection picks the anchor that minimizes search fan-out — prefer
the rarest or most positionally constrained fixed letter.

**Query modes supported:**

| Mode | Pattern shape | Available | Excluded |
|------|---------------|-----------|----------|
| Crossword | `__A__ED` | nil | nil |
| Scrabble | `__A__ED` | `[P,R,I,Z,S,O,U]` | nil |
| Wordle (hard mode) | `____E` | known good letters | gray letters |
| Anagram | all Free, exact length | `[A,E,R,T,S]` | nil |

**Output:**

- `Iterator` — lazy, implements `range`-over-func (Go 1.23+) for idiomatic
  `for word := range results { ... }` usage. Documented with the Go version
  requirement and rationale.
- `Iterator.Collect() []string` — materializes all results eagerly.

### Top-Level Convenience (`gadgad.go`)

```go
index, err := gadgad.Load(dict.File("words.txt"))
results := gadgad.Search(index, "??A??ED", gadgad.WithAvailable("PRIZSO"))
for word := range results { ... }
```

Users who only need this surface never need to import `index/` or `query/`
directly.

---

## Package Structure

```
github.com/axiomabsolute/gadgad/
│
├── gadgad.go              top-level Load + Search convenience API
│
├── index/
│   ├── index.go           Index type, Build(), public traversal API
│   ├── node.go            internal node + edge representation
│   └── minimize.go        DAWG minimization (Daciuk algorithm)
│
├── query/
│   ├── pattern.go         Pattern, Slot, LengthConstraint, LetterSet types
│   ├── engine.go          Search(): pattern → traversal → results
│   ├── anchor.go          AnchorStrategy and auto-selection logic
│   └── iterator.go        Iterator type, range-over-func, Collect()
│
└── dict/
    ├── dict.go            Source interface: Words() ([]string, error)
    ├── file.go            FileSource: line-delimited text file
    └── slice.go           SliceSource: immediate []string
```

---

## Dictionary Loading

The `dict.Source` interface decouples the index from how words are provided:

```go
type Source interface {
    Words() ([]string, error)
}
```

**v1 implementations:**

- `dict.File(path string) Source` — reads a UTF-8 text file, one word per line,
  trims whitespace, skips blank lines and `#` comments.
- `dict.Slice(words []string) Source` — wraps a `[]string` directly; useful for
  tests, embedded wordlists, and programmatic construction.

Both normalize words to uppercase before returning them, matching the GADDAG's
internal representation.

---

## Implementation Phases

### Phase 1 — Core GADDAG (highest risk, do first)

Goal: a correct, tested GADDAG that can be constructed from `[]string` and
traversed from a single anchor with no constraints.

Deliverables:
- `index.Build()` and internal graph construction
- DAWG minimization (`minimize.go`)
- Single-anchor unconstrained traversal
- Unit tests: for a small fixed word list, every word is reachable from every
  anchor letter in every rotation
- Benchmark: construction time and memory for a standard wordlist (TWL06, ~180k
  words)

This phase is intentionally isolated. Nothing else is built until Phase 1 is
correct.

### Phase 2 — Dictionary Loading

Goal: `dict.Source` implementations wired into `index.Build()`.

Deliverables:
- `dict.Source` interface
- `dict.FileSource` with normalization and basic error handling
- `dict.SliceSource`
- Integration test: load TWL06 from file, confirm word count and a sample of
  known words

### Phase 3 — Query Engine

Goal: structured queries with constraints; lazy and eager output.

Deliverables:
- `query.Pattern` type and all constraint fields
- `query.Search()` translating patterns to traversals
- Anchor auto-selection
- Available letter set constraint (rack-aware pruning during traversal)
- Exclusion constraint
- `Iterator` with range-over-func and `Collect()`
- Tests covering all four query modes from the table above
- Benchmark: query latency for representative patterns on TWL06

### Phase 4 — Public Surface and Documentation

Goal: library is usable by strangers, not just its author.

Deliverables:
- `gadgad.go` convenience wrapper
- Godoc on all exported symbols
- Runnable `Example*` test functions covering common use cases
- `README.md` with:
  - Quick start
  - GADDAG algorithm explainer with ASCII diagrams
  - References to source papers
  - Query mode examples (Scrabble, crossword, Wordle, anagram)
- Benchmarks published in README

---

## Testing Strategy

- **Unit tests** for each layer in isolation (package-level `_test.go` files)
- **Property tests** for GADDAG correctness: for any word in the dictionary,
  every rotation must be reachable via the corresponding anchor traversal
- **Integration tests** using real wordlists (TWL06 or SOWPODS) checked into
  `testdata/`
- **Benchmark tests** (`Benchmark*`) for construction and query hot paths;
  establish baselines in Phase 1 and track regressions
- **Example tests** (`Example*`) as both documentation and regression guards

---

## Future Considerations

### Dynamic Dictionary Updates

The GADDAG's space efficiency comes from graph minimization — a batch operation
over the full word set. Online insertions invalidate shared subgraph nodes and
require local re-minimization, which is complex and potentially expensive.
Deletions are harder still.

If dynamic updates become a requirement, the likely path is a separate
implementation: an unminimized trie with the same GADDAG rotation encoding,
accepting worse memory characteristics in exchange for mutability. The
`dict.Source` interface and `query.Pattern` API are designed to be
implementation-agnostic, so a dynamic index would be a drop-in alternative to
`index.Build()` without breaking the upper layers.

### Game Adapters (Layer 3)

- `ScrabbleBoard`: encodes a board state, derives patterns for each open anchor
  position, and ranks candidates by tile score
- `CrosswordGrid`: resolves intersecting clues into multi-anchor patterns
- `WordleGame`: maintains gray/yellow/green state across guesses, derives
  exclusion + inclusion constraints automatically

These are explicitly out of scope for v1 but the layered design accommodates
them without changes to Layers 1 or 2.

### Alternative Encodings

The DAWG minimization produces near-optimal space usage for the static case. If
memory pressure is not a concern (e.g., the dictionary is small or memory is
abundant), an unminimized trie with the GADDAG rotation encoding is simpler to
implement and easier to understand. A `BuildUnminimized()` constructor could
serve as a reference implementation for students comparing the two.

---

## References

- Gordon, S. (1994). *A faster Scrabble move generation algorithm.* Software —
  Practice and Experience, 24(2), 219–232.
- Daciuk, J., Mihov, S., Watson, B. W., & Watson, R. E. (2000). *Incremental
  construction of minimal acyclic finite-state automata.* Computational
  Linguistics, 26(1), 3–16.
- Appel, A. W., & Jacobson, G. J. (1988). *The world's fastest Scrabble
  program.* Communications of the ACM, 31(5), 572–578.
